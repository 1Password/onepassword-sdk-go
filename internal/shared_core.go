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
		core = &SharedCore{
			plugin:         p,
			instByClientID: map[uint64]*pluginInstance{},
		}
	}

	return core, nil
}

func ReleaseCore() {
	core.plugin.Close(context.Background())
	core = nil
}

type pluginInstance struct {
	// lock is used to synchronize access to the WASM core instance which is single threaded
	lock sync.Mutex

	// underlying instance of the WASM core
	plugin *extism.Plugin
}

// SharedCore implements Core in such a way that all created client instances share the same core resources.
type SharedCore struct {
	// lock is used to synchronize access to the shared WASM compiled core which is single threaded as well as
	// the map of plugin instances by client ID
	lock sync.Mutex

	// plugin is the Extism plugin which represents the WASM core loaded into memory
	plugin *extism.CompiledPlugin

	// client instances are the Extism plugin instances spawned from the compiled plugin per client
	instByClientID map[uint64]*pluginInstance
}

func (c *SharedCore) spawnInstance(ctx context.Context) (*extism.Plugin, error) {
	extismConfig := extism.PluginConfig{}
	i, err := c.plugin.Instance(ctx, extism.PluginInstanceConfig{
		ModuleConfig: extismConfig.ModuleConfig,
	})
	if err != nil {
		return nil, err
	}

	return i, nil
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func (c *SharedCore) InitClient(ctx context.Context, config ClientConfig) (*uint64, error) {
	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	// first return parameter is a sys.Exit code, which we don't need since the error is fully recoverable
	res, inst, err := c.callWithCtx(ctx, initClientFuncName, marshaledConfig)
	if err != nil {
		return nil, err
	}
	var id uint64
	err = json.Unmarshal(res, &id)
	if err != nil {
		return nil, err
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	c.instByClientID[id] = &pluginInstance{plugin: inst}
	return &id, nil
}

// Invoke calls specified business logic from core
func (c *SharedCore) Invoke(ctx context.Context, invokeConfig InvokeConfig) (*string, error) {
	input, err := json.Marshal(invokeConfig)
	if err != nil {
		return nil, err
	}
	var res []byte
	var i *pluginInstance
	clientID := invokeConfig.Invocation.ClientID
	if clientID != nil {
		i = c.instByClientID[*clientID]
	}
	if i != nil {
		i.lock.Lock()
		defer i.lock.Unlock()
		_, res, err = i.plugin.CallWithContext(ctx, invokeFuncName, input)
	} else {
		c.lock.Lock()
		defer c.lock.Unlock()
		var inst *extism.Plugin
		res, inst, err = c.callWithCtx(ctx, invokeFuncName, input)
		defer inst.Close(ctx)
	}
	if err != nil {
		return nil, err
	}

	response := string(res)

	return &response, nil
}

// ReleaseClient releases memory in the core associated with the given client ID.
func (c *SharedCore) ReleaseClient(clientID uint64) {
	if i, ok := c.instByClientID[clientID]; ok {
		marshaledClientID, err := json.Marshal(clientID)
		if err != nil {
			i.plugin.Log(extism.LogLevelWarn, fmt.Sprintf("memory couldn't be released: %s", err.Error()))
		}
		_, _, err = i.plugin.Call(releaseClientFuncName, marshaledClientID)
		if err != nil {
			i.plugin.Log(extism.LogLevelWarn, "memory couldn't be released")
		}
		delete(c.instByClientID, clientID)
	}
}

func (c *SharedCore) callWithCtx(ctx context.Context, functionName string, serializedParameters []byte) ([]byte, *extism.Plugin, error) {
	i, err := c.spawnInstance(ctx)
	if err != nil {
		return nil, nil, err
	}

	_, response, err := i.CallWithContext(ctx, functionName, serializedParameters)
	if err != nil {
		return nil, nil, err
	}
	return response, i, nil
}

// `loadWASM` returns the WASM core loaded into an `extism.Plugin`.
func loadWASM(ctx context.Context) (*extism.CompiledPlugin, error) {
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: coreWASM,
			},
		},
		AllowedHosts: allowed1PHosts(),
	}

	extismConfig := extism.PluginConfig{}
	plugin, err := extism.NewCompiledPlugin(ctx, manifest, extismConfig, ImportedFunctions())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %v", err)
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
