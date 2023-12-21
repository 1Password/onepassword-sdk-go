package main

import (
	"context"
	"fmt"
	"os"

	onepassword "github.com/1password/1password-go-sdk/src"
)

func main() {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	client, err := onepassword.Client(
		onepassword.WithServiceAccountToken(token),
		onepassword.WithApp(onepassword.DefaultAppName, onepassword.DefaultAppVersion),
		onepassword.WithContext(context.Background()),
	)

	secret, err := client.Secrets.Resolve("op://xw33qlvug6moegr3wkk5zkenoa/bckakdku7bgbnyxvqbkpehifki/foobar")
	if err != nil {
		panic(err)
	}

	doSomethingSecret(*secret)
}

func doSomethingSecret(secret string) {
	fmt.Println(secret)
}
