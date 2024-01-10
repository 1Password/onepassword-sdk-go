package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	onepassword "github.com/1password/1password-go-sdk"
)

// This is an example for retrieving a secret from 1Password and setting it as SECRET_ENV_VAR using the SDK client.

func main() {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Print("Before creating client\n")
	fmt.Printf("Alloc = %v bytes\n", m.Alloc)
	fmt.Printf("TotalAlloc = %v bytes\n", m.TotalAlloc)
	client, err := onepassword.Client(
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo(onepassword.DefaultIntegrationName, onepassword.DefaultIntegrationVersion),
		onepassword.WithContext(context.Background()),
	)
	if err != nil {
		panic(err)
	}
	runtime.ReadMemStats(&m)
	fmt.Print("After creating client\n")
	fmt.Printf("Alloc = %v bytes\n", m.Alloc)
	fmt.Printf("TotalAlloc = %v bytes\n", m.TotalAlloc)

	secret, err := client.Secrets.Resolve("op://xw33qlvug6moegr3wkk5zkenoa/bckakdku7bgbnyxvqbkpehifki/foobar")
	if err != nil {
		panic(err)
	}
	runtime.ReadMemStats(&m)
	fmt.Print("After resolving secret ref\n")
	fmt.Printf("Alloc = %v bytes\n", m.Alloc)
	fmt.Printf("TotalAlloc = %v bytes\n", m.TotalAlloc)

	print("Secret: " + *secret)
}
