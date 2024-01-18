package main

import (
	"context"
	"os"

	onepassword "github.com/1password/1password-go-sdk"
)

// This is an example for retrieving a secret from 1Password and setting it as SECRET_ENV_VAR using the SDK client.

func main() {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	client, err := onepassword.Client(
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo(onepassword.DefaultIntegrationName, onepassword.DefaultIntegrationVersion),
		onepassword.WithContext(context.Background()),
	)
	if err != nil {
		panic(err)
	}

	secret, err := client.Secrets.Resolve("op://ljetebpbiql2tgwkqoa2vogrvi/qvncatly7yydv2lpj7f3gu5ntm/password")
	if err != nil {
		panic(err)
	}

	doSomethingSecret(*secret)
	println(*secret)
}

func doSomethingSecret(secret string) {
	err := os.Setenv("SECRET_ENV_VAR", secret)
	if err != nil {
		panic(err)
	}
}
