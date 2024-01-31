# 1Password Go SDK

> ❗ This project is still in its early, pre-alpha stages of development. Its stability is not yet fully assessed, and future iterations may bring backwards incompatible changes. Proceed with caution.

The 1Password Go SDK offers programmatic read access to your secrets in 1Password in an interface native to Go. The SDK currently supports authentication with [1Password Service Accounts](https://developer.1password.com/docs/service-accounts/).

## Get started

To use the 1Password Go SDK in your project:

1. [Create a 1Password Service Account](https://developer.1password.com/docs/service-accounts/get-started/#create-a-service-account). Make sure to grant the service account access to the vaults where the secrets your project needs access to are stored.
2. Export your service account token to the `OP_SERVICE_ACCOUNT_TOKEN` environment variable:

```bash
export OP_SERVICE_ACCOUNT_TOKEN=<your-service-account-token>
```

3. Add the 1Password namespace on GitHub to your [`GOPRIVATE` environment variable](https://pkg.go.dev/cmd/go#hdr-Configuration_for_downloading_non_public_code):

```bash
export GOPRIVATE=${GOPRIVATE},github.com/1password/*
```

4. Redirect the default traffic of `go get` from HTTPS to SSH. To do this, add the following snippet to your `~/.gitconfig` file:

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
import (
    "context"
    "os"

    onepassword "github.com/1password/1password-go-sdk"
)

func main() {
    token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	
    client, err := onepassword.NewClient(
        context.TODO()
        onepassword.WithServiceAccountToken(token),
        onepassword.WithIntegrationInfo("<your app name>", "<your app version>"), 
    )
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

Make sure to use [secret reference URIs](https://developer.1password.com/docs/cli/secret-references/) with the syntax `op://vault/item/field` to securely load secrets from 1Password into your code.

