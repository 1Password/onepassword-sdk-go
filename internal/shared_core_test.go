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

	// check ExportedFunctionsDefinitions names contain init_client, invoke and release_client
	functions := [3]string{"init_client", "invoke", "release_client"}
	count := 0

	for _, function := range functions {
		if value.FunctionExists(function){
			count++
		}
	}

	assert.Equal(t, len(functions), count)

	// check AllowedHosts field matches allowed1PHosts
	pluginHosts := sort.StringSlice(value.AllowedHosts)
	opHosts := sort.StringSlice(allowed1PHosts())

	assert.Equal(t, len(pluginHosts), len(opHosts))

	for x := range pluginHosts {
		assert.Equal(t, pluginHosts[x], opHosts[x])
	}
}
