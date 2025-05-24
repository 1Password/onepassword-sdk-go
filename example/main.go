package main

import (
	"context"
	"fmt"
	"github.com/1password/onepassword-sdk-go"
)

// [developer-docs.sdk.go.sdk-import]-end

func main() {
	// Authenticates with your service account token and connects to 1Password.
	client, err := onepassword.NewClient(context.Background(),
		// TODO: Set the following to your own integration name and version.
		onepassword.WithIntegrationInfo("Go SDK Auth Prompt", "v1.0.0"),
	)
	if err != nil {
		panic(err)
	}

	secret,err := client.Secrets().Resolve(context.Background(), "op://vault/item/field")
	if err != nil {
		panic(err)
	}
	fmt.Println("secret:", secret)

}
