package internal

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

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

// GetSharedCore initializes the shared core once and re-uses it on subsequent calls.
func GetSharedCore(ctx context.Context) (*SharedCore, error) {
	if core == nil {
		p, err := loadWASM(ctx)
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
	plugin *extism.Plugin
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func (c SharedCore) InitClient(config ClientConfig) (*uint64, error) {
	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	_, res, err := c.plugin.Call(initClientFuncName, marshaledConfig)
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
func (c SharedCore) Invoke(invokeConfig InvokeConfig) (*string, error) {
	input, err := json.Marshal(invokeConfig)
	if err != nil {
		return nil, err
	}
	_, res, err := c.plugin.Call(invokeFuncName, input)
	if err != nil {
		return nil, err
	}

	response := string(res)

	return &response, nil
}

// ReleaseClient releases memory in the core associated with the given client ID.
func (c SharedCore) ReleaseClient(clientID uint64) {
	marshaledClientID, err := json.Marshal(clientID)
	if err != nil {
		c.plugin.Log(extism.LogLevelError, fmt.Sprintf("memory couldn't be released: %s", err.Error()))
	}
	_, _, err = c.plugin.Call(releaseClientFuncName, marshaledClientID)
	if err != nil {
		c.plugin.Log(extism.LogLevelError, fmt.Sprintf("memory couldn't be released: %s", err.Error()))
	}
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
