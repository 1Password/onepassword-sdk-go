package onepassword

import (
	"context"
	"fmt"

	"github.com/1password/onepassword-sdk-go/internal"
)

// ErrRuntimeClosed is returned by operations on a closed Runtime.
var ErrRuntimeClosed = internal.ErrRuntimeClosed

type runtimeOptions struct{ directory string }

// RuntimeOption configures a Runtime.
type RuntimeOption func(*runtimeOptions) error

// WithCompilationCacheDir persists compiled WASM code in directory, creating it
// if necessary on first use. It does not cache credentials, sessions or secrets.
// The directory must be trusted: cache files contain executable code. Entries
// depend on the runtime version, architecture, CPU features and WASM artifact.
// A missing entry is compiled normally; this is not a require-hit cache.
func WithCompilationCacheDir(directory string) RuntimeOption {
	return func(options *runtimeOptions) error {
		if directory == "" {
			return fmt.Errorf("compilation cache directory must not be empty")
		}
		options.directory = directory
		return nil
	}
}

// Runtime owns a WASM core shared by its service-account clients. Construct it
// with NewRuntime, do not copy it, and close it when its clients are no longer
// needed. It is independent of the core used by the package-level NewClient.
// Operations on one Runtime, including its clients' calls, are serialized.
type Runtime struct{ core *internal.OwnedRuntime }

// NewRuntime creates a lazy runtime without loading WASM, accessing the filesystem
// or authenticating a client. Without options, its compilation cache is in memory.
func NewRuntime(options ...RuntimeOption) (*Runtime, error) {
	config := runtimeOptions{}
	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("nil runtime option")
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}
	return &Runtime{core: internal.NewOwnedRuntime(config.directory)}, nil
}

// Prepare loads and instantiates the embedded WASM without authentication or
// network requests. It can prepopulate a compilation cache before serving clients.
// Failed loads may be retried; no partial core is published.
func (r *Runtime) Prepare(ctx context.Context) error { return r.core.Prepare(ctx) }

// NewClient authenticates a service-account client using this runtime.
// Desktop authentication is unsupported; use the package-level NewClient for it.
func (r *Runtime) NewClient(ctx context.Context, options ...ClientOption) (*Client, error) {
	client := Client{config: internal.NewDefaultConfig()}
	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("nil client option")
		}
		if err := option(&client); err != nil {
			return nil, err
		}
	}
	if client.config.AccountName != nil {
		return nil, fmt.Errorf("owned runtimes support only service-account authentication")
	}
	if err := r.Prepare(ctx); err != nil {
		return nil, err
	}
	return initClient(ctx, internal.CoreWrapper{InnerCore: r.core}, client)
}

// Close waits for active operations rather than cancelling them, then closes the
// core and cache. The context cannot interrupt waiting for the runtime lock.
// Repeated calls return the original cleanup error. Later client calls return
// ErrRuntimeClosed; asynchronous client finalizers cannot access a closed core.
func (r *Runtime) Close(ctx context.Context) error { return r.core.Close(ctx) }
