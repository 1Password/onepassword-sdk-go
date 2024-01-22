package onepassword

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests were designed for CI/CD. If you want to run them locally you must make sure the following dependencies are in place:
// A valid (test) Service Account Token is set in the environment - export OP_SERVICE_ACCOUNT_TOKEN = ...
// Secret references and expected values are matching existing secrets in the test account.

func TestSecretRetrievalFromTestAccount(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	client, err := NewClient(context.TODO(),
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, DefaultIntegrationVersion),
	)
	if err != nil {
		panic(err)
	}

	secret, err := client.Secrets.Resolve("op://tfctuk7dxnrwjwqqhwatuhy3gi/dqtyg7dswx5kvpcxwv32psdbse/password")
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "test_password", *secret)
}

func TestInitClientIncrement(t *testing.T) {

	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	config := NewDefaultConfig()
	config.SAToken = token
	config.IntegrationName = "name"
	config.IntegrationVersion = "version"

	value1, err1 := Core.InitClient(core, config)
	require.NoError(t, err1)
	value2, err2 := Core.InitClient(core, config)
	require.NoError(t, err2)
	value3, err3 := Core.InitClient(core, config)
	require.NoError(t, err3)

	assert.Equal(t, uint64(0), *value1)
	assert.Equal(t, uint64(1), *value2)
	assert.Equal(t, uint64(2), *value3)
}

func TestNewClientId(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	config1 := NewDefaultConfig()
	config1.SAToken = token
	config1.IntegrationName = "name"
	config1.IntegrationVersion = "version"
	config2 := NewDefaultConfig()
	config2.SAToken = token
	config2.IntegrationName = "name"
	config2.IntegrationVersion = "version"
	value1, _ := Core.InitClient(core, config1)
	value2, _ := Core.InitClient(core, config2)
	assert.NotEqual(t, *value1, *value2)
}

//
// invalid NewClient calls
//

func TestNoIntegrationNameOrVersion(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("", ""))
	require.Error(t, err)
}

func TestNoIntegrationName(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("", DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestInvalidIntegrationNameLength(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("12345678901234567890123456789012345678901234567890", DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestInvalidIntegrationNameCharacters(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("$", DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestNoIntegrationVersion(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, ""))
	require.Error(t, err)
}

func TestInvalidIntegrationVersionLength(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "12345678901234567890123456789012345678901234567890"))
	require.Error(t, err)
}

func TestInvalidIntegrationVersionCharacters(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "$"))
	require.Error(t, err)
}

//
// end of invalid NewClient Calls
//

func TestReleaseClient(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	// ensure latest id is not zero
	ctx := context.TODO()
	core, _ := NewExtismCore(ctx)
	config := NewDefaultConfig()
	config.SAToken = token
	config.IntegrationName = "name"
	config.IntegrationVersion = "version"
	Core.InitClient(core, config)
	latest, _ := Core.InitClient(core, config)

	// release memory
	Core.ReleaseClient(core, *latest)

	// check next initialization has id zero
	value, _ := Core.InitClient(core, config)
	assert.Equal(t, uint64(0), *value)
}
