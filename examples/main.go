package main

import (
	"context"
	"os"
	"strconv"
	"time"

	onepassword "github.com/1password/1password-go-sdk"
)

// This is an example for retrieving a secret from 1Password and setting it as SECRET_ENV_VAR using the SDK client.

func main() {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	println("Initializing client")
	beforeInit := time.Now().UnixNano()
	client, err := onepassword.Client(
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo(onepassword.DefaultIntegrationName, onepassword.DefaultIntegrationVersion),
		onepassword.WithContext(context.Background()),
	)
	afterInit := time.Now().UnixNano()

	println("Total time to initialize client: " + strconv.FormatInt(afterInit-beforeInit, 10))
	println("Done initializing client")

	if err != nil {
		panic("A " + err.Error())
	}
	for i := 0; i < 5; i++ {
		println("Making invocation #" + strconv.Itoa(i))
		beforeInvocation := time.Now().UnixNano()
		secret, err := client.Secrets.Resolve("op://xw33qlvug6moegr3wkk5zkenoa/bckakdku7bgbnyxvqbkpehifki/password")
		if err != nil {
			panic("B " + err.Error())
		}
		afterInvocation := time.Now().UnixNano()
		println("Secret: " + *secret)
		println("Finished invocation #" + strconv.Itoa(i))
		println("Total time for invocation: " + strconv.FormatInt(afterInvocation-beforeInvocation, 10))
	}

}
