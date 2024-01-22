package onepassword

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWASM(t *testing.T) {
	ctx := context.TODO()
	value, _ := loadWASM(ctx)

	// check only one module field
	if len(value.Modules) != 1 {
		t.Fatal("1 extism plugin module expected, ", len(value.Modules), " instead")
	}
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

	if count != len(list) {
		t.Fatal("ExportedFunctionDefinitions expected ", len(list), " functions, ", count, " were counted instead")
	}

	// check AllowedHosts field matches allowed1PHosts
	plugin_hosts := sort.StringSlice(value.AllowedHosts)
	op_hosts := sort.StringSlice(allowed1PHosts())

	assert.Equal(t, len(plugin_hosts), len(op_hosts))

	for x := range plugin_hosts {
		assert.Equal(t, plugin_hosts[x], op_hosts[x])
	}
}

func TestInvalidClientConfig(t *testing.T) {
	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	config := NewDefaultConfig() // invalid without setting SAToken field
	_, err := Core.InitClient(core, config)
	require.Error(t, err)
}

func TestInvalidInvoke(t *testing.T) {

	valid_clientId := 0
	valid_methodName := ""
	valid_params := ""
	invalid_clientId := 0
	invalid_methodName := ""
	invalid_params := ""

	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)

	// invalid client id
	invocation1 := Invocation{uint64(invalid_clientId), valid_methodName, valid_params}
	_, err1 := Core.Invoke(core, invocation1)
	println(err1.Error())

	require.Error(t, err1)

	// invalid method name
	invocation2 := Invocation{uint64(valid_clientId), invalid_methodName, valid_params}
	_, err2 := Core.Invoke(core, invocation2)

	require.Error(t, err2)

	// serialized params
	invocation3 := Invocation{uint64(valid_clientId), valid_methodName, invalid_params}
	_, err3 := Core.Invoke(core, invocation3)

	require.Error(t, err3)
}

func TestReleaseClient(t *testing.T) {
	// ensure latest id is not zero
	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	config := ClientConfig{} //TODO: make valid config
	Core.InitClient(core, config)
	latest, _ := Core.InitClient(core, config)

	// release memory
	Core.ReleaseClient(core, *latest)

	// check next initialization has id zero
	value, _ := Core.InitClient(core, config)
	assert.Equal(t, 0, *value)
}
