package _password_go_sdk

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	core "github.com/1password/1password-sdk-core/wasm"
	extism "github.com/extism/go-sdk"
)

const (
	invokeFuncName        = "invoke"
	initClientFuncName    = "init_client"
	releaseClientFuncName = "release_client"
	allowedHostsPattern   = "*.1password"
)

var corePlugin *extism.Plugin

type Invocation struct {
	ClientID         uint64 `json:"client"`
	MethodName       string `json:"name"`
	SerializedParams string `json:"data"`
}

// InitClient creates a client instance in the current core module and returns its unique ID.
func InitClient(ctx context.Context, config ClientConfig) (*uint64, error) {
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
		fmt.Printf("Failed to initialize plugin: %v\n", err)
		os.Exit(1)
	}
	corePlugin = plugin

	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	status, res, err := plugin.Call(initClientFuncName, marshaledConfig)
	if err != nil {
		plugin.Log(extism.LogLevelError, fmt.Sprintf("%s exited with status %d", initClientFuncName, status))
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
	status, res, err := corePlugin.Call(invokeFuncName, input)
	if err != nil {
		corePlugin.Log(extism.LogLevelError, fmt.Sprintf("%s exited with status %d", initClientFuncName, status))
		return nil, err
	}

	response := string(res)

	return &response, nil
}

// ReleaseClient releases memory in core associated to the given client ID.
func ReleaseClient(clientId uint64) {
	data, _ := json.Marshal(clientId)
	_, _, _ = corePlugin.Call(releaseClientFuncName, data)
}

func allowed1PHosts() []string {
	var hosts []string
	hosts = append(hosts, allowedHostsPattern+".com")
	hosts = append(hosts, allowedHostsPattern+".ca")
	hosts = append(hosts, allowedHostsPattern+".eu")
	return hosts
}
