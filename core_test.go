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
