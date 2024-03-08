package integration_tests

import (
	"context"
	"os"
	"runtime"
	"sync"
	"testing"

	"github.com/1password/onepassword-sdk-go"
	"github.com/1password/onepassword-sdk-go/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests were designed for CI/CD. If you want to run them locally you must make sure the following dependencies are in place:
// A valid (test) Service Account Token is set in the environment - export OP_SERVICE_ACCOUNT_TOKEN = ...
// Secret references and expected values are matching existing secrets in the test account.

func TestSecretRetrievalFromTestAccount(t *testing.T) {
	t.Cleanup(func() {
		internal.ReleaseCore()
	})

	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	client, err := onepassword.NewClient(context.TODO(),
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("Integration_Test_Go_SDK", onepassword.DefaultIntegrationVersion),
	)
	require.NoError(t, err)

	secret, err := client.Secrets.Resolve(context.TODO(), "op://gowwbvgow7kxocrfmfvtwni6vi/6ydrn7ne6mwnqc2prsbqx4i4aq/password")
	require.NoError(t, err)

	assert.Equal(t, "test_password_42", secret)
}

func TestRetrivalWithMultipleClients(t *testing.T) {
	t.Cleanup(func() {
		internal.ReleaseCore()
	})
	TestSecretRetrievalFromTestAccount(t)
	TestSecretRetrievalFromTestAccount(t)
	TestSecretRetrievalFromTestAccount(t)

	// keep creating clients to check what happens
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	core, _ := internal.GetSharedCore()
	config := internal.NewDefaultConfig()
	config.SAToken = token
	config.IntegrationName = "name"
	config.IntegrationVersion = "version"

	ctx := context.TODO()
	value1, err1 := core.InitClient(ctx, config)
	require.NoError(t, err1)
	value2, err2 := core.InitClient(ctx, config)
	require.NoError(t, err2)
	value3, err3 := core.InitClient(ctx, config)
	require.NoError(t, err3)

	assert.Equal(t, uint64(3), *value1)
	assert.Equal(t, uint64(4), *value2)
	assert.Equal(t, uint64(5), *value3)
}

func TestInvalidInvoke(t *testing.T) {
	t.Cleanup(func() {
		internal.ReleaseCore()
	})
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	core, err := internal.GetSharedCore()
	require.NoError(t, err)

	config := internal.NewDefaultConfig()
	config.SAToken = token
	config.IntegrationName = "name"
	config.IntegrationVersion = "version"

	_, err = core.InitClient(context.TODO(), config)
	require.NoError(t, err)

	validClientID := uint64(0)
	validMethodName := "Resolve"
	validParams := "op://gowwbvgow7kxocrfmfvtwni6vi/6ydrn7ne6mwnqc2prsbqx4i4aq/password"
	invalidClientID := uint64(1)
	invalidMethodName := "InvalidName"
	invalidParams := ""

	// invalid client id
	invocation1 := internal.InvokeConfig{
		ClientID: invalidClientID,
		Invocation: internal.Invocation{
			MethodName:       validMethodName,
			SerializedParams: validParams,
		},
	}
	_, err1 := core.Invoke(context.TODO(), invocation1)
	assert.EqualError(t, err1, "internal error: invalid client id")

	// invalid method name
	invocation2 := internal.InvokeConfig{
		ClientID: validClientID,
		Invocation: internal.Invocation{
			MethodName:       invalidMethodName,
			SerializedParams: invalidParams,
		}}
	_, err2 := core.Invoke(context.TODO(), invocation2)
	assert.EqualError(t, err2, "unknown variant `InvalidName`, expected `Resolve` at line 1 column 48")

	// invalid serialized params
	invocation3 := internal.InvokeConfig{
		ClientID: validClientID,
		Invocation: internal.Invocation{
			MethodName:       validMethodName,
			SerializedParams: invalidParams,
		},
	}
	_, err3 := core.Invoke(context.TODO(), invocation3)
	assert.EqualError(t, err3, "error resolving secret reference: secret reference is not prefixed with \"op://\"")
}

func TestClientReleasedSuccessfully(t *testing.T) {
	TestSecretRetrievalFromTestAccount(t)
	runtime.GC()

	core, err := internal.GetSharedCore()
	require.NoError(t, err)

	invocation := internal.InvokeConfig{
		ClientID: 0, // this client id should be invalid because the client has been cleaned up by GC
		Invocation: internal.Invocation{
			MethodName:       "Resolve",
			SerializedParams: "SecretRef",
		},
	}
	_, err = core.Invoke(context.TODO(), invocation)
	assert.EqualError(t, err, "internal error: invalid client id")
}

func TestConcurrentCallsFromOneClient(t *testing.T) {
	t.Cleanup(func() {
		internal.ReleaseCore()
	})
	var wg sync.WaitGroup
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	client, err := onepassword.NewClient(context.TODO(),
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("Integration_Test_Go_SDK", onepassword.DefaultIntegrationVersion),
	)
	require.NoError(t, err)

	concurrentCalls := 10
	wg.Add(concurrentCalls)
	for i := 0; i < concurrentCalls; i++ {
		go func() {
			secret, err := client.Secrets.Resolve(context.TODO(), "op://gowwbvgow7kxocrfmfvtwni6vi/6ydrn7ne6mwnqc2prsbqx4i4aq/password")
			require.NoError(t, err)

			assert.Equal(t, "test_password_42", secret)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestConcurrentCallsFromMultipleClientsOnTheSameToken(t *testing.T) {
	t.Cleanup(func() {
		internal.ReleaseCore()
	})
	var wg sync.WaitGroup
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	concurrentClients := 5
	wg.Add(concurrentClients)
	for i := 0; i < concurrentClients; i++ {
		go func() {
			client, err := onepassword.NewClient(context.TODO(),
				onepassword.WithServiceAccountToken(token),
				onepassword.WithIntegrationInfo("Integration_Test_Go_SDK", onepassword.DefaultIntegrationVersion),
			)
			require.NoError(t, err)

			secret, err := client.Secrets.Resolve(context.TODO(), "op://gowwbvgow7kxocrfmfvtwni6vi/6ydrn7ne6mwnqc2prsbqx4i4aq/password")
			require.NoError(t, err)

			assert.Equal(t, "test_password_42", secret)
			wg.Done()
		}()
	}
	wg.Wait()
}
