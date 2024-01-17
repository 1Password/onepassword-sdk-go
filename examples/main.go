package main

import (
	"context"
	"os"

	onepassword "github.com/1password/1password-go-sdk"
)

// This is an example of how to retrieve a secret from 1Password and set it as SECRET_ENV_VAR using the SDK client.

func main() {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory, err := onepassword.NewClientFactory(context.TODO())
	if err != nil {
		panic(err)
	}

	client, err := clientFactory.NewClient(
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo(onepassword.DefaultIntegrationName, onepassword.DefaultIntegrationVersion),
	)
	if err != nil {
		panic(err)
	}

	secret, err := client.Secrets.Resolve("op://xw33qlvug6moegr3wkk5zkenoa/bckakdku7bgbnyxvqbkpehifki/foobar")
	if err != nil {
		panic(err)
	}

	doSomethingSecret(*secret)
}

func doSomethingSecret(secret string) {
	err := os.Setenv("SECRET_ENV_VAR", secret)
	if err != nil {
		panic(err)
	}
}
