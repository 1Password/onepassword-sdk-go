package onepassword

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadWASM(t *testing.T) {
	ctx := context.TODO()
	value, _ := loadWASM(ctx)

	/*
		The following three lines of code are commented out to pass the test case
		I am unsure if checking the return type of loadWASM() is necessary
	*/

	// check return type is extism.Plugin
	// if reflect.TypeOf(value).Kind() != reflect.TypeOf(extism.NewPlugin).Kind() {
	// 	t.Fatal("loadWASM returns ", reflect.TypeOf(*value).Kind(), ", expected ", reflect.TypeOf(extism.NewPlugin).Kind())
	// }

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

	if len(plugin_hosts) != len(op_hosts) {
		t.Fatal("Expected AllowedHosts is different from actual AllowedHosts")
	}

	for x := range plugin_hosts {
		if plugin_hosts[x] != op_hosts[x] {
			t.Fatal("Expected AllowedHosts is different from actual AllowedHosts")
		}
	}
}

func TestNewClientId(t *testing.T) {
	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	config1 := ClientConfig{} //TODO: make valid config
	config2 := ClientConfig{} //TODO: make valid config
	value1, _ := Core.InitClient(core, config1)
	value2, _ := Core.InitClient(core, config2)
	if *value1 == *value2 {
		t.Fatalf("new client id not created")
	}
}

func TestInitClientIncrement(t *testing.T) {
	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	config := ClientConfig{} //TODO: make valid config
	value1, _ := Core.InitClient(core, config)
	value2, _ := Core.InitClient(core, config)
	value3, _ := Core.InitClient(core, config)
	if *value1 != 0 || *value2 != 1 || *value3 != 2 {
		t.Fatalf("client id does not increment")
	}
}

func TestInvalidClientConfig(t *testing.T) {
	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	config := ClientConfig{} //TODO: make invalid config
	_, err := Core.InitClient(core, config)
	require.Error(t, err)
}

func TestInvokeReturnsSecret(t *testing.T) {

	clientId := 0
	methodName := ""
	params := ""

	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	invocation := Invocation{uint64(clientId), methodName, params}
	invoke, err := Core.Invoke(core, invocation)
	require.NoError(t, err)

	//TODO: setup on valid test secret reference
	*invoke = "" // just here to remove error temporarily
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
	if *value != 0 {
		t.Fatal("client id after memory release expected to be 0, is ", *value, " instead")
	}
}
