package internal

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
)

func TestOwnedRuntimeCloseWaitsForLoad(t *testing.T) {
	ctx := context.Background()
	r := NewOwnedRuntime("")
	entered, finish := make(chan struct{}), make(chan struct{})
	var closed atomic.Bool
	r.load = func(context.Context, wazero.CompilationCache) (ownedPlugin, error) {
		close(entered)
		<-finish
		return &testOwnedPlugin{close: func() error { closed.Store(true); return nil }}, nil
	}
	prepared := make(chan error, 1)
	go func() { prepared <- r.Prepare(ctx) }()
	<-entered
	done := make(chan error, 1)
	go func() { done <- r.Close(ctx) }()
	select {
	case <-done:
		t.Fatal("close did not wait for load")
	case <-time.After(20 * time.Millisecond):
	}
	close(finish)
	require.NoError(t, <-prepared)
	require.NoError(t, <-done)
	require.True(t, closed.Load())
	require.ErrorIs(t, r.Prepare(ctx), ErrRuntimeClosed)
}
