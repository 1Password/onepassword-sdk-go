package internal

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
)

type testOwnedPlugin struct {
	init    func()
	invoke  func()
	release func()
	close   func() error
}

func (p *testOwnedPlugin) InitClient(context.Context, []byte) ([]byte, error) {
	if p.init != nil {
		p.init()
	}
	return []byte("1"), nil
}
func (p *testOwnedPlugin) Invoke(context.Context, []byte) ([]byte, error) {
	if p.invoke != nil {
		p.invoke()
	}
	return []byte("ok"), nil
}
func (p *testOwnedPlugin) ReleaseClient([]byte) {
	if p.release != nil {
		p.release()
	}
}
func (p *testOwnedPlugin) Close(context.Context) error {
	if p.close != nil {
		return p.close()
	}
	return nil
}

type testOwnedCache struct {
	wazero.CompilationCache
	close func() error
}

func (c *testOwnedCache) Close(context.Context) error { return c.close() }

func TestOwnedRuntimeConcurrentPrepareAndInit(t *testing.T) {
	r := NewOwnedRuntime("")
	defer r.Close(context.Background())
	var loads, inits atomic.Int32
	r.load = func(context.Context, wazero.CompilationCache) (ownedPlugin, error) {
		loads.Add(1)
		return &testOwnedPlugin{init: func() { inits.Add(1) }}, nil
	}
	var wg sync.WaitGroup
	for range 30 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			assert.NoError(t, r.Prepare(context.Background()))
		}()
		go func() {
			defer wg.Done()
			_, err := r.InitClient(context.Background(), nil)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	require.Equal(t, int32(1), loads.Load())
	require.Equal(t, int32(30), inits.Load())
}

func TestOwnedRuntimeLoadRetryOwnsFreshCache(t *testing.T) {
	ctx := context.Background()
	r := NewOwnedRuntime("")
	failure := errors.New("load failure")
	cleanup := errors.New("cache cleanup failure")
	var creates, closes, loads int
	r.newCache = func() (wazero.CompilationCache, error) {
		creates++
		return &testOwnedCache{close: func() error { closes++; return cleanup }}, nil
	}
	r.load = func(context.Context, wazero.CompilationCache) (ownedPlugin, error) {
		loads++
		if loads == 1 {
			return nil, failure
		}
		return &testOwnedPlugin{}, nil
	}
	err := r.Prepare(ctx)
	require.ErrorIs(t, err, failure)
	require.ErrorIs(t, err, cleanup)
	require.Nil(t, r.plugin)
	require.Nil(t, r.cache)
	require.Equal(t, 1, closes)
	require.NoError(t, r.Prepare(ctx))
	require.NoError(t, r.Prepare(ctx))
	require.Equal(t, 2, creates)
	require.ErrorIs(t, r.Close(ctx), cleanup)
	require.Equal(t, 2, closes)
}

func TestOwnedRuntimeCloseWaitsAndGuardsFinalizerRelease(t *testing.T) {
	ctx := context.Background()
	r := NewOwnedRuntime("")
	entered, finish := make(chan struct{}), make(chan struct{})
	var releases atomic.Int32
	var order []string
	pluginError, cacheError := errors.New("plugin close"), errors.New("cache close")
	r.newCache = func() (wazero.CompilationCache, error) {
		return &testOwnedCache{close: func() error { order = append(order, "cache"); return cacheError }}, nil
	}
	r.load = func(context.Context, wazero.CompilationCache) (ownedPlugin, error) {
		return &testOwnedPlugin{
			invoke:  func() { close(entered); <-finish },
			release: func() { releases.Add(1) },
			close:   func() error { order = append(order, "plugin"); return pluginError },
		}, nil
	}
	require.NoError(t, r.Prepare(ctx))
	callDone := make(chan struct{})
	go func() { defer close(callDone); _, err := r.Invoke(ctx, nil); require.NoError(t, err) }()
	<-entered
	closeDone := make(chan error, 1)
	go func() { closeDone <- r.Close(ctx) }()
	select {
	case <-closeDone:
		t.Fatal("closed during active call")
	case <-time.After(20 * time.Millisecond):
	}
	close(finish)
	<-callDone
	err := <-closeDone
	require.ErrorIs(t, err, pluginError)
	require.ErrorIs(t, err, cacheError)
	require.Equal(t, []string{"plugin", "cache"}, order)
	require.Equal(t, err, r.Close(ctx))
	require.ErrorIs(t, r.Prepare(ctx), ErrRuntimeClosed)
	_, err = r.InitClient(ctx, nil)
	require.ErrorIs(t, err, ErrRuntimeClosed)
	_, err = r.Invoke(ctx, nil)
	require.ErrorIs(t, err, ErrRuntimeClosed)
	var wg sync.WaitGroup
	for range 30 {
		wg.Add(1)
		go func() { defer wg.Done(); r.ReleaseClient([]byte("1")) }()
	}
	wg.Wait()
	require.Zero(t, releases.Load(), "late finalizer called closed plugin")
}

func TestOwnedRuntimeCloseBeforePrepareAndCancellation(t *testing.T) {
	r := NewOwnedRuntime("")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, r.Prepare(ctx), context.Canceled)
	require.Nil(t, r.plugin)
	require.NoError(t, r.Close(ctx))
	require.ErrorIs(t, r.Prepare(context.Background()), ErrRuntimeClosed)
}
