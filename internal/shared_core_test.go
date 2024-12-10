package internal

import (
	"context"
	"sort"
	"testing"

	extism "github.com/extism/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWASM(t *testing.T) {
	ctx := context.TODO()
	core, err := loadWASM(ctx)
	require.NoError(t, err)
	value, err := core.Instance(ctx, extism.PluginInstanceConfig{})
	require.NoError(t, err)

	// check that the main module field is set
	assert.NotNil(t, value.Module())

	// check ExportedFunctionsDefinitions names contain init_client, invoke and release_client
	functions := [3]string{"init_client", "invoke", "release_client"}
	count := 0

	for _, function := range functions {
		if _, exists := value.Module().ExportedFunctions()[function]; exists {
			count++
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
