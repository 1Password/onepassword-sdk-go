package onepassword

import (
	"context"
	"crypto/rand"
	"fmt"
	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero/api"
)

// ImportedFunctions returns all functions 1Password SDK core must import.
func ImportedFunctions() []extism.HostFunction {
	return []extism.HostFunction{randomFillFunc()}
}

// randomFillFunc returns an Extism Function that writes n (input) random bytes into memory using crypto/rand, and returns their offset.
func randomFillFunc() extism.HostFunction {
	randomFill := extism.NewHostFunctionWithStack("random_fill_imported", func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
		length := api.DecodeU32(stack[0])

		b := make([]byte, length)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}
		stack[0], err = p.WriteBytes(b)
		if err != nil {
			panic(fmt.Errorf("failed to write bytes: %v", err))
		}
	}, []api.ValueType{api.ValueTypeI32}, []api.ValueType{api.ValueTypeI64})
	randomFill.SetNamespace("op-random")

	return randomFill
}
