<p align="center">
  <a href="https://1password.com">
      <h1 align="center">1Password Go SDK</h1>
  </a>
</p>

<p align="center">
 <h4 align="center">Build integrations that programmatically access your secrets in 1Password.</h4>
</p>

<p align="center">
  <a href="https://developer.1password.com/docs/sdks/">Documentation</a> | <a href="https://github.com/1Password/onepassword-sdk-go/tree/main/example">Examples</a>
<br/>

---

## 🚀 Get started

To use the 1Password Go SDK in your project:

1. [Create a service account](https://my.1password.com/developer-tools/infrastructure-secrets/serviceaccount/) and give it the appropriate permissions in the vaults where the items you want to use with the SDK are saved.
2. Provision your service account token. We recommend provisioning your token from the environment. For example, to export your token to the `OP_SERVICE_ACCOUNT_TOKEN` environment variable:

   **macOS or Linux**

   ```bash
   export OP_SERVICE_ACCOUNT_TOKEN=<your-service-account-token>
   ```

   **Windows**

   ```powershell
   $Env:OP_SERVICE_ACCOUNT_TOKEN = "<your-service-account-token>"
   ```

3. Install the 1Password Go SDK in your project:

   ```bash
   go get github.com/1password/onepassword-sdk-go
   ```

4. Use the Go SDK in your project:

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
                // TODO: Set the following to your own integration name and version.
                onepassword.WithIntegrationInfo("My 1Password Integration", "v1.0.0"),
    )
    if err != nil {
	// handle err
    }
    secret, err := client.Secrets.Resolve(context.TODO(), "op://vault/item/field")
    if err != nil {
        // handle err
    }
    // do something with the secret
}
```

Make sure to use [secret reference URIs](https://developer.1password.com/docs/cli/secrets-reference-syntax/) with the syntax `op://vault/item/field` to securely load secrets from 1Password into your code.

## Supported functionality

1Password SDKs are in active development. We're keen to hear what you'd like to see next. Let us know by [upvoting](https://github.com/1Password/onepassword-sdk-go/issues) or [filing](https://github.com/1Password/onepassword-sdk-go/issues/new/choose) an issue.

### Item management
Operations:
- [x] [Retrieve secrets](https://developer.1password.com/docs/sdks/load-secrets)
- [x] [Retrieve items](https://developer.1password.com/docs/sdks/manage-items#get-an-item)
- [x] [Create items](https://developer.1password.com/docs/sdks/manage-items#create-an-item)
- [x] [Update items](https://developer.1password.com/docs/sdks/manage-items#edit-an-item)
- [x] [Delete items](https://developer.1password.com/docs/sdks/manage-items#delete-an-item)
- [x] [List items](https://developer.1password.com/docs/sdks/list-vaults-items/)
- [ ] Add & update tags on items 

Field types:
- [x] API Keys
- [x] Passwords
- [x] Concealed fields
- [x] Text fields
- [x] Notes
- [x] SSH private keys (partially supported: supported in resolving secret references, not yet supported in item create/get/update)
- [ ] SSH public keys, fingerprint and key type
- [x] One-time passwords 
- [x] URLs
- [x] Phone numbers
- [x] Credit card types
- [ ] Files attachments and Document items ([#108](https://github.com/1Password/onepassword-sdk-go/issues/108))

### Vault management
- [ ] Retrieve vaults
- [ ] Create vaults
- [ ] Update vaults
- [ ] Delete vaults
- [x] [List vaults](https://developer.1password.com/docs/sdks/list-vaults-items/)

### User & access management
- [ ] Provision users
- [ ] Retrieve users
- [ ] List users
- [ ] Suspend users
- [ ] Create groups
- [ ] Update group membership
- [ ] Update vault access & permissions

### Compliance & reporting
- [ ] Watchtower insights
- [ ] Travel mode
- [ ] Events ([#76](https://github.com/1Password/onepassword-sdk-go/issues/76)). For now, use [1Password Events Reporting API](https://developer.1password.com/docs/events-api/) directly.

### Authentication

- [x] [1Password Service Accounts](https://developer.1password.com/docs/service-accounts/get-started/)
- [ ] User authentication
- [ ] 1Password Connect. For now, use [1Password/connect-sdk-go](https://github.com/1Password/connect-sdk-go).

## 📖 Learn more

- [Load secrets with 1Password SDKs](https://developer.1password.com/docs/sdks/load-secrets)
- [Manage items with 1Password SDKs](https://developer.1password.com/docs/sdks/manage-items)
- [List vaults and items with 1Password SDKs](https://developer.1password.com/docs/sdks/list-vaults-items)
- [1Password SDK concepts](https://developer.1password.com/docs/sdks/concepts)
