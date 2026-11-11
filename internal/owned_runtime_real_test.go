package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
)

type rejectPrepareNetwork struct{ t *testing.T }

func (r rejectPrepareNetwork) RoundTrip(*http.Request) (*http.Response, error) {
	r.t.Error("Prepare attempted a network request")
	return nil, fmt.Errorf("network disabled in credential-free test")
}

func TestOwnedRuntimeRealCoreCache(t *testing.T) {
	originalTransport := http.DefaultTransport
	http.DefaultTransport = rejectPrepareNetwork{t}
	defer func() { http.DefaultTransport = originalTransport }()
	ctx := context.Background()
	dir := filepath.Join(t.TempDir(), "cache")
	beforeGlobal := core
	r := NewOwnedRuntime(dir)
	defer r.Close(ctx)
	_, err := os.Stat(dir)
	require.True(t, os.IsNotExist(err), "construction must not access the filesystem")
	require.NoError(t, r.Prepare(ctx))
	loaded := r.plugin.(*compiledCore)
	require.Equal(t, allowed1PHosts(), loaded.plugin.AllowedHosts)
	for _, name := range []string{initClientFuncName, invokeFuncName, releaseClientFuncName} {
		require.True(t, loaded.plugin.FunctionExists(name))
	}
	require.NoError(t, r.Close(ctx))
	require.Same(t, beforeGlobal, core, "owned load must not publish global core")
	type entry struct {
		data  []byte
		mtime time.Time
	}
	entries := map[string]entry{}
	require.NoError(t, filepath.WalkDir(dir, func(path string, item os.DirEntry, err error) error {
		if err != nil || item.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := item.Info()
		if err != nil {
			return err
		}
		entries[path] = entry{data, info.ModTime()}
		return nil
	}))
	require.NotEmpty(t, entries)

	// A fresh owner has no shared in-memory engine and reuses the disk entries.
	warm := NewOwnedRuntime(dir)
	defer warm.Close(ctx)
	require.NoError(t, warm.Prepare(ctx))
	require.NoError(t, warm.Close(ctx))
	for path, previous := range entries {
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, previous.data, data)
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, previous.mtime, info.ModTime(), "warm entries should not be rewritten")
	}
}

func TestOwnedRuntimeCacheCreationFailureCanRetry(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	require.NoError(t, os.WriteFile(dir, []byte("not a directory"), 0600))
	r := NewOwnedRuntime(dir)
	defer r.Close(context.Background())
	require.Error(t, r.Prepare(context.Background()))
	require.Nil(t, r.plugin)
	require.Nil(t, r.cache)
	require.NoError(t, os.Remove(dir))
	// A fresh cache factory call is made on retry. Use the test loader to avoid
	// recompiling the real module merely to exercise directory error recovery.
	r.load = func(context.Context, wazero.CompilationCache) (ownedPlugin, error) {
		return &testOwnedPlugin{}, nil
	}
	require.NoError(t, r.Prepare(context.Background()))
}
