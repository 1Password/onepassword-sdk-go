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
	client, err := onepassword.NewClient(context.Background(),
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("Integration_Test_Go_SDK", onepassword.DefaultIntegrationVersion),
	)
	require.NoError(t, err)

	secret, err := client.Secrets.Resolve(context.Background(), "op://gowwbvgow7kxocrfmfvtwni6vi/6ydrn7ne6mwnqc2prsbqx4i4aq/password")
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

	ctx := context.Background()
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

	_, err = core.InitClient(context.Background(), config)
	require.NoError(t, err)

	validClientID := uint64(0)
	validMethodName := "SecretsResolve"
	validParams := map[string]interface{}{"secret_reference": "op://gowwbvgow7kxocrfmfvtwni6vi/6ydrn7ne6mwnqc2prsbqx4i4aq/password"}
	invalidClientID := uint64(1)
	invalidMethodName := "InvalidName"
	invalidParams := map[string]interface{}{"secret_reference": ""}

	// invalid client id
	invocation1 := internal.InvokeConfig{
		Invocation: internal.Invocation{
			ClientID: &invalidClientID,
			Parameters: internal.Parameters{
				MethodName:       validMethodName,
				SerializedParams: validParams,
			},
		},
	}
	_, err1 := core.Invoke(context.Background(), invocation1)
	assert.EqualError(t, err1, "an internal error occurred, please contact 1Password at support@1password.com or https://developer.1password.com/joinslack: invalid client id")

	// invalid method name
	invocation2 := internal.InvokeConfig{
		Invocation: internal.Invocation{
			ClientID: &validClientID,
			Parameters: internal.Parameters{
				MethodName:       invalidMethodName,
				SerializedParams: invalidParams,
			},
		}}
	_, err2 := core.Invoke(context.Background(), invocation2)
	assert.NotNil(t, err2, "expected error when sending invocation that doesn't exist")

	// invalid serialized params
	invocation3 := internal.InvokeConfig{
		Invocation: internal.Invocation{
			ClientID: &validClientID,
			Parameters: internal.Parameters{
				MethodName:       validMethodName,
				SerializedParams: invalidParams,
			},

		},
	}
	_, err3 := core.Invoke(context.Background(), invocation3)
	assert.EqualError(t, err3, "error resolving secret reference: secret reference is not prefixed with \"op://\"")
}

func TestClientReleasedSuccessfully(t *testing.T) {
	TestSecretRetrievalFromTestAccount(t)
	runtime.GC()

	core, err := internal.GetSharedCore()
	require.NoError(t, err)
	clientID  := uint64(0)
	invocation := internal.InvokeConfig{
		Invocation: internal.Invocation{
			ClientID: &clientID, // this client id should be invalid because the client has been cleaned up by GC
			Parameters: internal.Parameters{
				MethodName:       "SecretsResolve",
				SerializedParams: map[string]interface{}{"secret_reference": "op://foo/bar/baz"},
			},
		},
	}
	_, err = core.Invoke(context.Background(), invocation)
	assert.EqualError(t, err, "an internal error occurred, please contact 1Password at support@1password.com or https://developer.1password.com/joinslack: invalid client id")
}

func TestConcurrentCallsFromOneClient(t *testing.T) {
	t.Cleanup(func() {
		internal.ReleaseCore()
	})
	var wg sync.WaitGroup
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	client, err := onepassword.NewClient(context.Background(),
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("Integration_Test_Go_SDK", onepassword.DefaultIntegrationVersion),
	)
	require.NoError(t, err)

	concurrentCalls := 10
	wg.Add(concurrentCalls)
	for i := 0; i < concurrentCalls; i++ {
		go func() {
			secret, err := client.Secrets.Resolve(context.Background(), "op://gowwbvgow7kxocrfmfvtwni6vi/6ydrn7ne6mwnqc2prsbqx4i4aq/password")
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
			client, err := onepassword.NewClient(context.Background(),
				onepassword.WithServiceAccountToken(token),
				onepassword.WithIntegrationInfo("Integration_Test_Go_SDK", onepassword.DefaultIntegrationVersion),
			)
			require.NoError(t, err)

			secret, err := client.Secrets.Resolve(context.Background(), "op://gowwbvgow7kxocrfmfvtwni6vi/6ydrn7ne6mwnqc2prsbqx4i4aq/password")
			require.NoError(t, err)

			assert.Equal(t, "test_password_42", secret)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestExpiredContextCancelsLongRunningOperation(t *testing.T) {
	c := context.Background()
	ctx, cancel := context.WithCancel(c)
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	var err error
	out := make(chan error)
	cancel()
	go func() {
		_, err = onepassword.NewClient(ctx,
			onepassword.WithServiceAccountToken(token),
			onepassword.WithIntegrationInfo("Integration_Test_Go_SDK", onepassword.DefaultIntegrationVersion),
		)
		out <- err
	}()

	err = <-out
	require.ErrorContains(t, err, `context canceled (recovered by wazero)`)
}
