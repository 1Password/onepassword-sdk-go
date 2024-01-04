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

var corePlugin *extism.Plugin

type Invocation struct {
	ClientID         uint64 `json:"client"`
	MethodName       string `json:"name"`
	SerializedParams string `json:"data"`
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func InitClient(ctx context.Context, config ClientConfig) (*uint64, error) {
	if corePlugin == nil {
		err := initClient(ctx)
		if err != nil {
			return nil, err
		}
	}
	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	_, res, err := corePlugin.Call(initClientFuncName, marshaledConfig)
	if err != nil {
		return nil, err
	}
	id := binary.BigEndian.Uint64(res)
	return &id, nil
}

// Invoke calls specified business logic from core
func Invoke(invokeConfig Invocation) (*string, error) {
	input, err := json.Marshal(invokeConfig)
	if err != nil {
		return nil, err
	}
	_, res, err := corePlugin.Call(invokeFuncName, input)
	if err != nil {
		return nil, err
	}

	response := string(res)

	return &response, nil
}

// ReleaseClient releases memory in core associated to the given client ID.
func ReleaseClient(clientID uint64) {
	marshaledClientID, _ := json.Marshal(clientID)
	_, _, err := corePlugin.Call(releaseClientFuncName, marshaledClientID)
	if err != nil {
		corePlugin.Log(extism.LogLevelError, fmt.Sprintf("memory couldn't be released: %s", err.Error()))
	}
}

func initClient(ctx context.Context) error {
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
		return fmt.Errorf("Failed to initialize plugin: %v\n", err)
	}
	corePlugin = plugin

	return nil
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
