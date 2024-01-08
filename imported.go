package onepassword

import (
	"context"
	"crypto/rand"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero/api"
)

// ImportedFunctions returns all functions 1Password SDK core must import.
func ImportedFunctions() []extism.HostFunction {
	return []extism.HostFunction{randomFillFunc()}
}

// randomFillFunc returns an Extism Function that writes random bytes into the WASM core's memory using crypto/rand.
func randomFillFunc() extism.HostFunction {
	randomFill := extism.NewHostFunctionWithStack("random_fill_imported", func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
		ptr := api.DecodeU32(stack[0])
		length := api.DecodeU32(stack[1])

		b := make([]byte, length)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}

		p.Memory().Write(ptr, b)
	}, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{})
	randomFill.SetNamespace("op-random")

	return randomFill
}
