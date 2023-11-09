# 1Password Go SDK

> ❗ This project is still in its early, pre-alpha stages of development. Its stability is not yet fully assessed, and future iterations may bring backwards incompatible changes. Proceed with caution.

The 1Password Go SDK offers programmatic read-access to your secrets in 1Password in an interface native to Go. To use it in your project:

1. [Create a 1Password Service Account](https://developer.1password.com/docs/service-accounts/get-started/#create-a-service-account).

2. Add the 1Password namespace on GitHub to your `GOPRIVATE` environment variable:

```bash
export GOPRIVATE=${GOPRIVATE},github.com/1password/*
```

3. Redirect the default traffic of `go get` from HTTPS to SSH. For this, add the following snippet to your `~/.gitconfig` file:

```
[url "ssh://git@github.com/"]
	insteadOf = https://github.com/
```

4. In your project, download the 1Password Go SDK:

```bash
go get github.com/1password/1password-go-sdk
```

5. Use the SDK in your project:

```go
import onepassword "github.com/1password/1password-go-sdk"

func main() {
	client, err := onepassword.NewServiceAccountClient("<your-service-account-token>")
	if err != nil {
		// handle err
	}
	secret, err := client.Resolve("op://vault/item/field")
	if err != nil {
		// handle err
	}
	// do something with the secret
}
```
For passing the service account token as an environment variable (`OP_SERVICE_ACCOUNT_TOKEN`), you can also leverage the `onepassword.NewServiceAccountClientFromEnv()` function.
