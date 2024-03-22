package internal

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	extism "github.com/extism/go-sdk"
)

//go:embed wasm/core.wasm
var coreWASM []byte

const (
	invokeFuncName        = "invoke"
	initClientFuncName    = "init_client"
	releaseClientFuncName = "release_client"
)

var core *SharedCore

// GetSharedCore initializes the shared core once and returns the already existing one on subsequent calls.
func GetSharedCore() (*SharedCore, error) {
	runtimeCtx := context.Background()
	if core == nil {
		p, err := loadWASM(runtimeCtx)
		if err != nil {
			return nil, err
		}
		core = &SharedCore{plugin: p}
	}

	return core, nil
}

func ReleaseCore() {
	core = nil
}

// SharedCore implements Core in such a way that all created client instances share the same core resources.
type SharedCore struct {
	// lock is used to synchronize access to the shared WASM core which is single threaded
	lock sync.Mutex
	// plugin is the Extism plugin which represents the WASM core loaded into memory
	plugin *extism.Plugin
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func (c *SharedCore) InitClient(ctx context.Context, config ClientConfig) (*uint64, error) {
	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	// first return parameter is a sys.Exit code, which we don't need since the error is fully recoverable
	res, err := c.callWithCtx(ctx, initClientFuncName, marshaledConfig)
	if err != nil {
		return nil, err
	}
	var id uint64
	err = json.Unmarshal(res, &id)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// Invoke calls specified business logic from core
func (c *SharedCore) Invoke(ctx context.Context, invokeConfig InvokeConfig) (*string, error) {
	input, err := json.Marshal(invokeConfig)
	if err != nil {
		return nil, err
	}
	res, err := c.callWithCtx(ctx, invokeFuncName, input)
	if err != nil {
		return nil, err
	}

	response := string(res)

	return &response, nil
}

// ReleaseClient releases memory in the core associated with the given client ID.
func (c *SharedCore) ReleaseClient(clientID uint64) {
	marshaledClientID, err := json.Marshal(clientID)
	if err != nil {
		c.plugin.Log(extism.LogLevelWarn, fmt.Sprintf("memory couldn't be released: %s", err.Error()))
	}
	_, err = c.call(releaseClientFuncName, marshaledClientID)
	if err != nil {
		c.plugin.Log(extism.LogLevelWarn, "memory couldn't be released")
	}
}

func (c *SharedCore) callWithCtx(ctx context.Context, functionName string, serializedParameters []byte) ([]byte, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, response, err := c.plugin.CallWithContext(ctx, functionName, serializedParameters)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *SharedCore) call(functionName string, serializedParameters []byte) ([]byte, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, response, err := c.plugin.Call(functionName, serializedParameters)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// `loadWASM` returns the WASM core loaded into an `extism.Plugin`.
func loadWASM(ctx context.Context) (*extism.Plugin, error) {
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: coreWASM,
			},
		},
		AllowedHosts: allowed1PHosts(),
	}

	extismConfig := extism.PluginConfig{}
	plugin, err := extism.NewPlugin(ctx, manifest, extismConfig, ImportedFunctions())
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize plugin: %v\n", err)
	}

	return plugin, nil
}

// `allowed1PHosts` returns all hosts accessible through the WASM core.
func allowed1PHosts() []string {
	return []string{
		"*.1password.com",
		"*.1password.ca",
		"*.1password.eu",
		"*.b5staging.com",
		"*.b5dev.com",
		"*.b5dev.ca",
		"*.b5dev.eu",
		"*.b5test.com",
		"*.b5test.ca",
		"*.b5test.eu",
		"*.b5rev.com",
		"*.b5local.com",
	}
}
