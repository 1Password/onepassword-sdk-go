package onepassword

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// These tests were designed for CI/CD. If you want to run them locally you must make sure the following dependencies are in place:
// A valid (test) Service Account Token is set in the environment - export OP_SERVICE_ACCOUNT_TOKEN = ...
// Secret references and expected values are matching existing secrets in the test account.

func TestSecretRetrievalFromTestAccount(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory, err := NewClientFactory(context.TODO())
	if err != nil {
		panic(err)
	}

	client, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("Integration_Test_Go_SDK", DefaultIntegrationVersion),
	)
	if err != nil {
		panic(err)
	}

	secret, err := client.Secrets.Resolve("op://xw33qlvug6moegr3wkk5zkenoa/bckakdku7bgbnyxvqbkpehifki/foobar")
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "test_password", secret)
}
