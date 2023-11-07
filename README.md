# 1Password Go SDK

> ❗ This project is still in its early, pre-alpha stages of development. Its stability is not yet fully assessed, and future iterations may bring backwards incompatible changes. Proceed with caution.

The 1Password Go SDK offers programmatic read-access to your secrets in 1Password in an interface native to Go. To use it in your project:

1. [Create a 1Password Service Account](https://developer.1password.com/docs/service-accounts/get-started/#create-a-service-account).

2. Add your service account token to your environment:

```bash
export OP_SERVICE_ACCOUNT_TOKEN=<your-sa-token>
```

3. Add the 1Password namespace on GitHub to your `GOPRIVATE` environment variable:

```bash
export GOPRIVATE=${GOPRIVATE},github.com/1password/*
```

4. Redirect the default traffic of `go get` from HTTPS to SSH. For this, add the following snippet to your `~/.gitconfig` file:

```
[url "ssh://git@github.com/"]
	insteadOf = https://github.com/
```

5. In your project, download the 1Password Go SDK:

```bash
# TODO - use the latest commit SHA

go get github.com/1password/1password-go-sdk@235ebb0693d82e2f9f886f7eb9d672d4ba3b1e8e
```

6. Import the SDK in any of your project’s files:

```go
import onepassword "github.com/1password/1password-go-sdk"
```

7. Create a client and read secrets with it:

```go
client, err := onepassword.NewServiceAccountClientFromEnv()
if err != nil {
	// handle err
}
secret, err := client.Resolve("op://path/to/your/secret")
if err != nil {
	// handle err
}
// do something with the secret
```
