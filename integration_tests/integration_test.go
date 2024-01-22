package integration_tests

import (
	"context"
	"os"
	"testing"

	onepassword "github.com/1password/1password-go-sdk"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests were designed for CI/CD. If you want to run them locally you must make sure the following dependencies are in place:
// A valid (test) Service Account Token is set in the environment - export OP_SERVICE_ACCOUNT_TOKEN = ...
// Secret references and expected values are matching existing secrets in the test account.

func TestSecretRetrievalFromTestAccount(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	client, err := onepassword.NewClient(context.TODO(),
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("Integration_Test_Go_SDK", onepassword.DefaultIntegrationVersion),
	)
	require.NoError(t, err)

	secret, err := client.Secrets.Resolve("op://tfctuk7dxnrwjwqqhwatuhy3gi/dqtyg7dswx5kvpcxwv32psdbse/password")
	require.NoError(t, err)

	assert.Equal(t, "test_password", *secret)
}

func TestInitClientIncrement(t *testing.T) {

	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	ctx := context.TODO()
	core, _ := onepassword.NewExtismCore(ctx)
	config := onepassword.NewDefaultConfig()
	config.SAToken = token
	config.IntegrationName = "name"
	config.IntegrationVersion = "version"

	value1, err1 := core.InitClient(config)
	require.NoError(t, err1)
	value2, err2 := core.InitClient(config)
	require.NoError(t, err2)
	value3, err3 := core.InitClient(config)
	require.NoError(t, err3)

	assert.Equal(t, uint64(0), *value1)
	assert.Equal(t, uint64(1), *value2)
	assert.Equal(t, uint64(2), *value3)
}

func TestInvalidInvoke(t *testing.T) {

	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	ctx := context.TODO()
	core, _ := onepassword.NewExtismCore(ctx)
	config := onepassword.NewDefaultConfig()
	config.SAToken = token
	config.IntegrationName = "name"
	config.IntegrationVersion = "version"

	value, _ := core.InitClient(config)

	client, err := onepassword.NewClient(context.TODO(),
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("Integration_Test_Go_SDK", onepassword.DefaultIntegrationVersion),
	)
	require.NoError(t, err)

	secret, err := client.Secrets.Resolve("op://tfctuk7dxnrwjwqqhwatuhy3gi/dqtyg7dswx5kvpcxwv32psdbse/password")
	require.NoError(t, err)

	validClientID := *value
	validMethodName := "Resolve"
	validParams := *secret
	invalidClientID := -1
	invalidMethodName := ""
	invalidParams := ""

	// invalid client id
	invocation1 := onepassword.Invocation{ClientID: uint64(invalidClientID), MethodName: validMethodName, SerializedParams: validParams}
	_, err1 := core.Invoke(invocation1)

	assert.Equal(t, "wrong method", err1.Error())

	// invalid method name
	invocation2 := onepassword.Invocation{ClientID: uint64(validClientID), MethodName: invalidMethodName, SerializedParams: validParams}
	_, err2 := core.Invoke(invocation2)

	assert.Equal(t, "wrong method", err2.Error())

	// invalid serialized params
	invocation3 := onepassword.Invocation{ClientID: uint64(validClientID), MethodName: validMethodName, SerializedParams: invalidParams}
	_, err3 := core.Invoke(invocation3)

	assert.Equal(t, "wrong method", err3.Error())
}
