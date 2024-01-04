package onepassword

import (
	"context"
	"crypto/rand"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero/api"
)

// ImportedFunctions returns all functions 1Password SDK core must import.
func ImportedFunctions() []extism.HostFunction {
	return []extism.HostFunction{randomNextByte()}
}

// randomNextByte returns an Extism Function that writes a random byte onto the stack memory using crypto/rand.
func randomNextByte() extism.HostFunction {
	randomNextByteFunc := extism.NewHostFunctionWithStack("random_next_u8", func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
		b := make([]byte, 1)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}
		stack[0] = api.EncodeI32(int32(b[0]))
	}, []api.ValueType{}, []api.ValueType{api.ValueTypeI32})
	randomNextByteFunc.SetNamespace("op-random")

	return randomNextByteFunc
}
