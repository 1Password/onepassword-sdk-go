package onepassword

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	core "github.com/1password/1password-sdk-core/wasm"
	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero/api"
)

const (
	invokeFuncName        = "invoke"
	initClientFuncName    = "init_client"
	releaseClientFuncName = "release_client"
	allowAllHosts         = "*.com"
)

var corePlugin *extism.Plugin

type InvokeConfig struct {
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
		AllowedHosts: []string{allowAllHosts},
	}

	extismConfig := extism.PluginConfig{}

	// functions that the WASM core imports
	randomNextByte := extism.NewHostFunctionWithStack("random_next_u8", func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
		b := make([]byte, 1)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}
		stack[0] = api.EncodeI32(int32(b[0]))
	}, []api.ValueType{}, []api.ValueType{api.ValueTypeI32})
	randomNextByte.SetNamespace("op-random")

	plugin, err := extism.NewPlugin(ctx, manifest, extismConfig, []extism.HostFunction{randomNextByte})
	if err != nil {
		fmt.Printf("Failed to initialize plugin: %v\n", err)
		os.Exit(1)
	}
	corePlugin = plugin

	data, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	status, res, err := plugin.Call(initClientFuncName, data)
	if err != nil {
		fmt.Println(status)
		return nil, err
	}
	id := uint64(binary.BigEndian.Uint32(res))
	return &id, nil
}

// Invoke calls specified business logic from core
func Invoke(config InvokeConfig) (*string, error) {
	input, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	status, res, err := corePlugin.Call(invokeFuncName, input)
	if err != nil {
		fmt.Println(status)
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
