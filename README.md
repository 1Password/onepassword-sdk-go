<p align="center">
  <a href="https://1password.com">
      <h1 align="center">1Password Go SDK (beta)</h1>
  </a>
</p>

<p align="center">
 <h4 align="center"> ❗ The 1Password SDK project is in beta. Future iterations may bring backwards-incompatible changes.</h4>
</p>

<p align="center">
  <a href="https://developer.1password.com/docs/sdks/">Documentation</a> | <a href="https://github.com/1Password/onepassword-sdk-go/tree/main/example">Examples</a>
<br/>

---

The 1Password Go SDK offers programmatic access to your secrets in 1Password with Go. During the beta, you can retrieve, create, read, update, and delete items.

## 🔑 Authentication

1Password SDKs support authentication with [1Password Service Accounts](https://developer.1password.com/docs/service-accounts/get-started/). You can create service accounts if you're an owner or administrator on your team. Otherwise, ask your administrator for a service account token.

Before you get started, [create a service account](https://developer.1password.com/docs/service-accounts/get-started/#create-a-service-account) and give it the appropriate permissions in the vaults where the items you want to use with the SDK are saved.

## ❗ Limitations

1Password SDKs currently only support operations on text and concealed fields. If you update or delete an item that has information saved in other field types, that information will be permanently lost.

1Password SDKs don't yet support using secret references with query parameters, so you can't retrieve file attachments or SSH keys, or get more information about field metadata.

When managing items with 1Password SDKs, you must use [unique identifiers (IDs)](https://developer.1password.com/docs/sdks/concepts#unique-identifiers-ids) in place of vault, item, and field names.

## 🚀 Get started

To use the 1Password Go SDK in your project:

1. Provision your [service account](#authentication) token. We recommend provisioning your token from the environment. For example, to export your token to the `OP_SERVICE_ACCOUNT_TOKEN` environment variable:
    
    **Mac**
    
    ```bash
    export OP_SERVICE_ACCOUNT_TOKEN=<your-service-account-token>
    ```
    
    **Windows**
    
    ```powershell
    $Env:OP_SERVICE_ACCOUNT_TOKEN = "<your-service-account-token>"
    ```


2. Add the 1Password GitHub namespace to your [`GOPRIVATE` environment variable](https://pkg.go.dev/cmd/go#hdr-Configuration_for_downloading_non_public_code):

    **Mac**

    ```bash
    export GOPRIVATE=${GOPRIVATE},github.com/1password/*
    ```

    **Windows**
    
    ```powershell
    $Env:GOPRIVATE=${GOPRIVATE},github.com/1password/*
    ```

3. Install the 1Password Go SDK in your project:

    ```bash
    go get github.com/1password/onepassword-sdk-go
    ```

4. Import the Go SDK in your project:

    ```go
    import (
    "github.com/1password/onepassword-sdk-go"
    )
    ```

5. Initalize the Go SDK. Provision your service account token and use `integration_name` and `integration_version` to specify a name and version for your application.

    ```go
    func main() {
        // Gets your service account token from the OP_SERVICE_ACCOUNT_TOKEN environment variable.
        token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

        // Authenticates with your service account token and connects to 1Password.
        client, err := onepassword.NewClient(context.Background(),
            onepassword.WithServiceAccountToken(token),
            onepassword.WithIntegrationInfo("My 1Password Integration", "v1.0.0"),
        )
        if err != nil {
            panic(err)
        }
    }
    ```
 	

### Example

```go
import (
    "context"
    "os"

    "github.com/1password/onepassword-sdk-go"
)

func main() {
    token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	
    client, err := onepassword.NewClient(
                context.TODO(),
                onepassword.WithServiceAccountToken(token),
                onepassword.WithIntegrationInfo("My 1Password Integration", "v1.0.0"),
    )
    if err != nil {
	// handle err
    }
    secret, err := client.Secrets.Resolve("op://vault/item/field")
    if err != nil {
        // handle err
    }
    // do something with the secret
}
```

Make sure to use [secret reference URIs](https://developer.1password.com/docs/cli/secret-references/) with the syntax `op://vault/item/field` to securely load secrets from 1Password into your code.

## Learn more

- [Load secrets with 1Password SDKs](https://developer.1password.com/docs/sdks/load-secrets)
- [Manage items with 1Password SDKs](https://developer.1password.com/docs/sdks/manage-items)
- [1Password SDK concepts](https://developer.1password.com/docs/sdks/concepts)


Inside ```onepassword.WithIntegrationInfo(...)```, pass the name of your application and the version of your application as arguments.
