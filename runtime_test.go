package onepassword

import (
	"context"
	"errors"
	goruntime "runtime"
	"testing"
	"time"

	"github.com/1password/onepassword-sdk-go/internal"
	"github.com/stretchr/testify/require"
)

func TestRuntimeOptions(t *testing.T) {
	for _, option := range []RuntimeOption{nil, WithCompilationCacheDir("")} {
		r, err := NewRuntime(option)
		require.Error(t, err)
		require.Nil(t, r)
	}
	r, err := NewRuntime()
	require.NoError(t, err)
	require.NoError(t, r.Close(context.Background()))
	require.ErrorIs(t, r.Prepare(context.Background()), ErrRuntimeClosed)
	_, err = r.NewClient(context.Background())
	require.ErrorIs(t, err, ErrRuntimeClosed)
}

func TestRuntimeRejectsDesktopBeforeLoading(t *testing.T) {
	r, err := NewRuntime()
	require.NoError(t, err)
	defer r.Close(context.Background())
	_, err = r.NewClient(context.Background(), func(c *Client) error { account := "test"; c.config.AccountName = &account; return nil })
	require.ErrorContains(t, err, "only service-account")
}

func TestDefaultNewClientOptionsRemainUnchanged(t *testing.T) {
	sentinel := errors.New("option failed")
	_, err := NewClient(context.Background(), func(*Client) error { return sentinel })
	require.ErrorIs(t, err, sentinel)
	_, err = NewClient(context.Background(), WithServiceAccountToken("not-a-credential"), func(c *Client) error { account := "test"; c.config.AccountName = &account; return nil })
	require.ErrorContains(t, err, "cannot use both")
}

// This test double returns a credential-free client ID while retaining the
// exact owned-core release path used by the production client finalizer.
type finalizerOwnedCore struct {
	*internal.OwnedRuntime
	released chan struct{}
}

func (c *finalizerOwnedCore) InitClient(context.Context, []byte) ([]byte, error) {
	return []byte("1"), nil
}

func (c *finalizerOwnedCore) ReleaseClient(id []byte) {
	c.OwnedRuntime.ReleaseClient(id)
	close(c.released)
}

func TestOwnedClientFinalizerAfterRuntimeClose(t *testing.T) {
	r, err := NewRuntime()
	require.NoError(t, err)
	core := &finalizerOwnedCore{OwnedRuntime: r.core, released: make(chan struct{})}
	func() {
		client, err := initClient(context.Background(), internal.CoreWrapper{InnerCore: core}, Client{})
		require.NoError(t, err)
		require.NoError(t, r.Close(context.Background()))
		goruntime.KeepAlive(client)
	}()
	deadline := time.After(5 * time.Second)
	for {
		goruntime.GC()
		select {
		case <-core.released:
			return
		case <-deadline:
			t.Fatal("client finalizer did not run")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestOwnedClientCallsPreserveClosedError(t *testing.T) {
	r, err := NewRuntime()
	require.NoError(t, err)
	require.NoError(t, r.Close(context.Background()))
	core := internal.CoreWrapper{InnerCore: r.core}
	_, err = initClient(context.Background(), core, Client{})
	require.ErrorIs(t, err, ErrRuntimeClosed)
	_, err = clientInvoke(context.Background(), &internal.InnerClient{Core: core}, "unused", nil)
	require.ErrorIs(t, err, ErrRuntimeClosed)
	// Exactly the wrapper retained by the standard client finalizer: safe after Close.
	core.ReleaseClient(1)
}
