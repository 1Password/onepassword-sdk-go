package internal

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWASM(t *testing.T) {
	ctx := context.TODO()
	value, err := loadWASM(ctx)
	require.NoError(t, err)

	// check that there's only one module field
	assert.Equal(t, 1, len(value.Modules))

	// check ExportedFunctionsDefinitions names contain init_client, invoke and release_client
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
