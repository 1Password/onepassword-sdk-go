package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero"
)

var ErrRuntimeClosed = errors.New("1password runtime is closed")

type ownedPlugin interface {
	Core
	Close(context.Context) error
}

type compiledCore struct {
	*ExtismCore
	compiled *extism.CompiledPlugin
}

func (c *compiledCore) Close(ctx context.Context) error {
	return errors.Join(c.plugin.Close(ctx), c.compiled.Close(ctx))
}

// OwnedRuntime serializes loading, guest calls and closure. Never copy it.
// Its Core implementation also protects clients' asynchronous finalizers.
type OwnedRuntime struct {
	mu       sync.Mutex
	closed   bool
	closeErr error
	plugin   ownedPlugin
	cache    wazero.CompilationCache
	newCache func() (wazero.CompilationCache, error)
	load     func(context.Context, wazero.CompilationCache) (ownedPlugin, error)
}

func NewOwnedRuntime(directory string) *OwnedRuntime {
	return &OwnedRuntime{
		newCache: func() (wazero.CompilationCache, error) {
			if directory == "" {
				return wazero.NewCompilationCache(), nil
			}
			return wazero.NewCompilationCacheWithDir(directory)
		},
		load: loadOwnedWASM,
	}
}

func loadOwnedWASM(ctx context.Context, cache wazero.CompilationCache) (ownedPlugin, error) {
	manifest := extism.Manifest{
		Wasm:         []extism.Wasm{extism.WasmData{Data: coreWASM}},
		AllowedHosts: allowed1PHosts(),
	}
	// Keep this configuration free of WASI, observers and filesystem resources.
	// Extism cannot expose/close its runtime if NewCompiledPlugin fails. The
	// caller closes the attempt's cache on failure; an inaccessible failed
	// constructor runtime cannot be deterministically closed through this API.
	compiled, err := extism.NewCompiledPlugin(ctx, manifest, extism.PluginConfig{
		RuntimeConfig: wazero.NewRuntimeConfig().WithCompilationCache(cache),
	}, ImportedFunctions())
	if err != nil {
		return nil, fmt.Errorf("failed to compile core: %w", err)
	}
	plugin, err := compiled.Instance(ctx, extism.PluginInstanceConfig{})
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to instantiate core: %w", err), compiled.Close(context.Background()))
	}
	return &compiledCore{ExtismCore: &ExtismCore{plugin: plugin}, compiled: compiled}, nil
}

func (r *OwnedRuntime) Prepare(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.prepare(ctx)
}

func (r *OwnedRuntime) prepare(ctx context.Context) error {
	if r.closed {
		return ErrRuntimeClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.plugin != nil {
		return nil
	}
	cache, err := r.newCache()
	if err != nil {
		return err
	}
	plugin, err := r.load(ctx, cache)
	if err != nil {
		return errors.Join(err, cache.Close(context.Background()))
	}
	r.plugin, r.cache = plugin, cache
	return nil
}

func (r *OwnedRuntime) InitClient(ctx context.Context, config []byte) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.prepare(ctx); err != nil {
		return nil, err
	}
	return r.plugin.InitClient(ctx, config)
}

func (r *OwnedRuntime) Invoke(ctx context.Context, config []byte) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil, ErrRuntimeClosed
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.plugin == nil {
		return nil, errors.New("1password runtime is not prepared")
	}
	return r.plugin.Invoke(ctx, config)
}

func (r *OwnedRuntime) ReleaseClient(id []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed && r.plugin != nil {
		r.plugin.ReleaseClient(id)
	}
}

// Close waits for active calls, does not cancel them, and permanently closes
// this owner. Cleanup is attempted in ownership order even when it fails.
func (r *OwnedRuntime) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return r.closeErr
	}
	r.closed = true
	if r.plugin != nil {
		r.closeErr = r.plugin.Close(ctx)
	}
	if r.cache != nil {
		r.closeErr = errors.Join(r.closeErr, r.cache.Close(ctx))
	}
	r.plugin, r.cache = nil, nil
	return r.closeErr
}
