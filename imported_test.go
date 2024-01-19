package onepassword

import (
	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/wazero/api"
	"testing"
)

func TestRandomFillFunc(t *testing.T) {
	r := randomFillImportedFunc()
	assert.Equal(t, r.Name, "random_fill_imported")
	assert.Equal(t, r.Params, []api.ValueType{api.ValueTypeI32})
	assert.Equal(t, r.Returns, []api.ValueType{api.ValueTypeI64})

	writeBytesToPluginMemoryMock := func(b []byte) (uint64, error) {
		return 25089, nil
	}
	var stack []uint64
	stack = append(stack, 6)
	randomFill(stack, writeBytesToPluginMemoryMock)
	assert.Equal(t, 1, len(stack))
	assert.Equal(t, uint64(25089), stack[0])
}

func TestImportedFunctions(t *testing.T) {
	// initial call
	value := ImportedFunctions()

	// check the returned function name is "random_fill_imported"
	if value[0].Name != "random_fill_imported" {
		t.Fatal("ImportedFunctions() expected to return function 'random_fill_imported', returns ", value[0].Name, " instead")
	}
}
