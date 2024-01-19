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

	assert.Equal(t, 0, value1)
	assert.Equal(t, 1, value2)
	assert.Equal(t, 2, value3)
}
