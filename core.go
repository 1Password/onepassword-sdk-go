package onepassword

import (
	"context"
	"encoding/binary"
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

var corePlugins map[uint64]*extism.Plugin

type Invocation struct {
	ClientID         uint64 `json:"client"`
	MethodName       string `json:"name"`
	SerializedParams string `json:"data"`
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func InitClient(ctx context.Context, config ClientConfig) (*uint64, error) {
	if corePlugins == nil {
		corePlugins = make(map[uint64]*extism.Plugin, 1)
	}
	plugin, err := initClient(ctx)
	if err != nil {
		return nil, err
	}
	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	_, res, err := plugin.Call(initClientFuncName, marshaledConfig)
	if err != nil {
		return nil, err
	}
	id := binary.LittleEndian.Uint64(res)
	corePlugins[id] = plugin
	return &id, nil
}

// Invoke calls specified business logic from core
func Invoke(invokeConfig Invocation) (*string, error) {
	input, err := json.Marshal(invokeConfig)
	if err != nil {
		return nil, err
	}
	_, res, err := corePlugins[invokeConfig.ClientID].Call(invokeFuncName, input)
	if err != nil {
		return nil, err
	}

	response := string(res)

	return &response, nil
}

// ReleaseClient releases memory in core associated to the given client ID.
func ReleaseClient(clientID uint64) {
	defer func() {
		corePlugins[clientID].Close()
		corePlugins[clientID] = nil
	}()
	marshaledClientID, err := json.Marshal(clientID)
	if err != nil {
		corePlugins[clientID].Log(extism.LogLevelError, fmt.Sprintf("memory couldn't be released: %s", err.Error()))
	}
	_, _, err = corePlugins[clientID].Call(releaseClientFuncName, marshaledClientID)
	if err != nil {
		corePlugins[clientID].Log(extism.LogLevelError, fmt.Sprintf("memory couldn't be released: %s", err.Error()))
	}
}

func initClient(ctx context.Context) (*extism.Plugin, error) {
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
