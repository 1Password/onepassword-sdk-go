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

// randomFillFunc returns an Extism Function that writes a random byte onto the stack memory using crypto/rand.
func randomFillFunc() extism.HostFunction {
	randomFillFunc := extism.NewHostFunctionWithStack("random_fill_imported", func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
		ptr := api.DecodeU32(stack[0])
		len := api.DecodeU32(stack[1])

		b := make([]byte, len)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}

		p.Memory().Write(ptr, b)
	}, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{})
	randomFillFunc.SetNamespace("op-random")

	return randomFillFunc
}
