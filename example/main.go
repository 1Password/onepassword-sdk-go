package main

import (
	"context"
	"os"

	onepassword "github.com/1password/onepassword-sdk-go"
)

func main() {
	// Gets your service account token from the OP_SERVICE_ACCOUNT_TOKEN environment variable.
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	// Authenticates with your service account token and connects to 1Password.
	client, err := onepassword.NewClient(context.Background(),
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("My 1Password Integration", "v1.0.0"),
	)
	if err != nil {
		panic(err)
	}

	// Retrieves a secret from 1Password.
	// Takes a secret reference as input and returns the secret to which it points.
	secret, err := client.Secrets.Resolve(context.Background(), "op://vault/item/field")
	if err != nil {
		panic(err)
	}

	doSomethingSecret(secret)
}

// Exports the secret to the SECRET_ENV_VAR environment variable.
func doSomethingSecret(secret string) {
	err := os.Setenv("SECRET_ENV_VAR", secret)
	if err != nil {
		panic(err)
	}
}
