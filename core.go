package onepassword

import (
	"context"
	"encoding/json"
	"fmt"

	core "github.com/1password/1password-sdk-core/wasm"
	extism "github.com/extism/go-sdk"
)

const (
	invokeFuncName        = "invoke"
	initClientFuncName    = "init_client"
	releaseClientFuncName = "release_client"
)

type Core interface {
	InitClient(config ClientConfig) (*uint64, error)
	Invoke(invokeConfig Invocation) (*string, error)
	ReleaseClient(clientID uint64)
}

type ExtismCore struct {
	plugin *extism.Plugin
}

func NewExtismCore(ctx context.Context) (Core, error) {
	p, err := loadWASM(ctx)
	if err != nil {
		return nil, err
	}
	return ExtismCore{plugin: p}, nil
}

// Invocation holds the information required for invoking SDK functionality.
type Invocation struct {
	ClientID         uint64 `json:"client"`
	MethodName       string `json:"name"`
	SerializedParams string `json:"data"`
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func (c ExtismCore) InitClient(config ClientConfig) (*uint64, error) {
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
func (c ExtismCore) Invoke(invokeConfig Invocation) (*string, error) {
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

// `ReleaseClient` releases memory in the core associated with the given client ID.
func (c ExtismCore) ReleaseClient(clientID uint64) {
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
				Data: core.GetWASMCore(),
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
