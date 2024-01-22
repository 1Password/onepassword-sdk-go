package onepassword

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadWASM(t *testing.T) {
	ctx := context.TODO()
	value, _ := loadWASM(ctx)

	// check only one module field
	assert.Equal(t, 1, len(value.Modules))

	// check ExportedFunctionsDefinitions names match (init_client, invoke, release_client)
	list := [3]string{"init_client", "invoke", "release_client"}

	count := 0
	for _, x := range list {
		for _, y := range value.Main.ExportedFunctionDefinitions() {
			if x == y.Name() {
				count++
			}
		}
	}

	assert.Equal(t, 3, count)

	// check AllowedHosts field matches allowed1PHosts
	pluginHosts := sort.StringSlice(value.AllowedHosts)
	opHosts := sort.StringSlice(allowed1PHosts())

	assert.Equal(t, len(pluginHosts), len(opHosts))

	for x := range pluginHosts {
		assert.Equal(t, pluginHosts[x], opHosts[x])
	}
}

func TestInvalidClientConfig(t *testing.T) {
	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	config := NewDefaultConfig() // invalid without setting SAToken field
	_, err := Core.InitClient(core, config)
	assert.Equal(t, "invalid service account token", err.Error())
}

func TestInvalidInvoke(t *testing.T) {

	validClientID := 0
	validMethodName := ""
	validParams := ""
	invalidClientID := 0
	invalidMethodName := ""
	invalidParams := ""

	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)

	// invalid client id
	invocation1 := Invocation{uint64(invalidClientID), validMethodName, validParams}
	_, err1 := core.Invoke(invocation1)
	println(err1.Error())

	assert.Equal(t, "wrong method", err1.Error())

	// invalid method name
	invocation2 := Invocation{uint64(validClientID), invalidMethodName, validParams}
	_, err2 := core.Invoke(invocation2)

	assert.Equal(t, "wrong method", err2.Error())

	// serialized params
	invocation3 := Invocation{uint64(validClientID), validMethodName, invalidParams}
	_, err3 := core.Invoke(invocation3)

	assert.Equal(t, "wrong method", err3.Error())
}
