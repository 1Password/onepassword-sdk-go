package main

import (
	"context"
	"os"
	"runtime"

	onepassword "github.com/1password/1password-go-sdk"
)

// This is an example for retrieving a secret from 1Password and setting it as SECRET_ENV_VAR using the SDK client.

func main() {
	runtime.MemProfileRate = 1
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	var m runtime.MemStats
	clients := make([]*onepassword.OpClient, 10)
	runtime.GC()
	runtime.ReadMemStats(&m)
	println(m.Alloc)

	for i := 0; i < 10; i++ {
		client, err := onepassword.Client(
			onepassword.WithServiceAccountToken(token),
			onepassword.WithIntegrationInfo(onepassword.DefaultIntegrationName, onepassword.DefaultIntegrationVersion),
			onepassword.WithContext(context.Background()),
		)
		if err != nil {
			panic(err)
		}
		clients[i] = client
		runtime.GC()
		runtime.ReadMemStats(&m)
		println(m.Alloc)
	}

	for i := 0; i < 10; i++ {
		secret, err := clients[i].Secrets.Resolve("op://xw33qlvug6moegr3wkk5zkenoa/bckakdku7bgbnyxvqbkpehifki/password")
		if err != nil {
			panic(err)
		}
		println("Secret: " + *secret)

	}
}
