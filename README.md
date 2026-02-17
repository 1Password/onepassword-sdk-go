<p align="center">
  <a href="https://1password.com">
      <h1 align="center">1Password Go SDK</h1>
  </a>
</p>

<p align="center">
 <h4 align="center">Build integrations that programmatically interact with 1Password.</h4>
</p>

<p align="center">
  <a href="https://developer.1password.com/docs/sdks/">Documentation</a> | <a href="https://github.com/1Password/onepassword-sdk-go/tree/main/example">Examples</a>
<br/>

---

## 🚀 Get started

You can choose between two [authentication methods](https://developer.1password.com/docs/sdks/concepts#authentication) for the 1Password Go SDK: local authorization prompts from the [1Password desktop app](#option-1-1password-desktop-app) or automated authentication with a [1Password Service Account](#option-2-1password-service-account).

### Option 1: 1Password desktop app

[1Password desktop app authentication](https://developer.1password.com/docs/sdks/concepts#1password-desktop-app) is best for local integrations that require minimal setup from end users and sensitive workflows that require human-in-the-loop approval. To set up the SDK to authenticate with the 1Password app:

1. Install the [1Password desktop app](https://1password.com/downloads/) and sign in to your account in the app.
2. Select your account or collection at the top of the sidebar, then navigate to **Settings** > [Developer](onepassword://settings/developers).
3. Under Integrate with the 1Password SDKs, select **Integrate with other apps**.
4. If you want to authenticate with biometrics, navigate to **Settings** > [Security](onepassword://settings/security), then turn on the option to unlock using [Touch ID](https://support.1password.com/touch-id-mac/),  [Windows Hello](https://support.1password.com/windows-hello/), or [system authentication](https://support.1password.com/system-authentication-linux/).
3. Install the 1Password Go SDK in your project:

   ```bash
   go get github.com/1password/onepassword-sdk-go
   ```

4. To use the Go SDK in your project, replace `your-account-name` in the code below with the name of your 1Password account as it appears at the top left sidebar of the 1Password desktop app.

```go
import (
    "context"
    "os"

    "github.com/1password/onepassword-sdk-go"
)

func main() {
    
    client, err := onepassword.NewClient(
                context.TODO(),
                onepassword.WithDesktopAppIntegration("your-account-name"),
                // TODO: Set the following to your own integration name and version.
                onepassword.WithIntegrationInfo("My 1Password Integration", "v1.0.0"),
    )
    if err != nil {
	// handle err
    }
    secret, err := client.Secrets().Resolve(context.TODO(), "op://vault/item/field")
    if err != nil {
        // handle err
    }
    // do something with the secret
}
```

Make sure to use [secret reference URIs](https://developer.1password.com/docs/cli/secret-reference-syntax/) with the syntax `op://vault/item/field` to securely load secrets from 1Password into your code.

### Option 2: 1Password Service Account

[Service account authentication](https://developer.1password.com/docs/sdks/concepts#service-account) is best for automated access and limiting your integration to least privilege access. To set up the SDK to authenticate with a service account token:

1. [Create a service account](https://my.1password.com/developer-tools/infrastructure-secrets/serviceaccount/?source=github-sdk) and give it the appropriate permissions in the vaults where the items you want to use with the SDK are saved.
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
    secret, err := client.Secrets().Resolve(context.TODO(), "op://vault/item/field")
    if err != nil {
        // handle err
    }
    // do something with the secret
}
```

Make sure to use [secret reference URIs](https://developer.1password.com/docs/cli/secret-reference-syntax/) with the syntax `op://vault/item/field` to securely load secrets from 1Password into your code.

## Supported functionality

1Password SDKs are in active development. We're keen to hear what you'd like to see next. Let us know by [upvoting](https://github.com/1Password/onepassword-sdk-go/issues) or [filing](https://github.com/1Password/onepassword-sdk-go/issues/new/choose) an issue.

### Item management

**Operations:**

- [x] [Retrieve secrets](https://developer.1password.com/docs/sdks/load-secrets)
- [x] [Retrieve items](https://developer.1password.com/docs/sdks/manage-items#get-an-item)
- [x] [Create items](https://developer.1password.com/docs/sdks/manage-items#create-an-item)
- [x] [Update items](https://developer.1password.com/docs/sdks/manage-items#update-an-item)
- [x] [Delete items](https://developer.1password.com/docs/sdks/manage-items#delete-an-item)
- [x] [Archive items](https://developer.1password.com/docs/sdks/manage-items/#archive-an-item)
- [x] [List items](https://developer.1password.com/docs/sdks/list-vaults-items/)
- [x] [Share items](https://developer.1password.com/docs/sdks/share-items)
- [x] [Generate PIN, random and memorable passwords](https://developer.1password.com/docs/sdks/manage-items#generate-a-password)

**Field types:**

- [x] API Keys
- [x] Passwords
- [x] Concealed fields
- [x] Text fields
- [x] Notes
- [x] SSH private keys, public keys, fingerprint and key type
- [x] One-time passwords
- [x] URLs
- [x] Websites (used to suggest and autofill logins)
- [x] Phone numbers
- [x] Credit card types
- [x] Credit card numbers
- [x] Emails
- [x] References to other items
- [x] Address
- [x] Date
- [x] MM/YY
- [x] File attachments and Document items
- [x] Menu
- [ ] Passkeys

### Vault management

- [x] [Retrieve vaults](https://developer.1password.com/docs/sdks/vaults#get-a-vault-overview)
- [x] [Create vaults](https://developer.1password.com/docs/sdks/vaults#create-a-vault)
- [x] [Update vaults](https://developer.1password.com/docs/sdks/vaults#update-a-vault)
- [x] [Delete vaults](https://developer.1password.com/docs/sdks/vaults#delete-a-vault)
- [x] [List vaults](https://developer.1password.com/docs/sdks/list-vaults-items#list-vaults)
- [x] [Manage group vault permissions](https://developer.1password.com/docs/sdks/vault-permissions)
- [ ] Manage user vault permissions

### User & access management

- [ ] Provision users
- [ ] Retrieve users
- [ ] List users
- [ ] Suspend users
- [x] [Retrieve groups](https://developer.1password.com/docs/sdks/groups/)
- [ ] List groups
- [ ] Create groups
- [ ] Update group membership

### Compliance & reporting
- [ ] Watchtower insights
- [ ] Travel mode
- [ ] Events ([#76](https://github.com/1Password/onepassword-sdk-go/issues/76)). For now, use [1Password Events Reporting API](https://developer.1password.com/docs/events-api/) directly.

### Authentication

- [x] [1Password Service Accounts](https://developer.1password.com/docs/sdks/concepts#1password-service-account)
- [x] [User authentication](https://developer.1password.com/docs/sdks/concepts#1password-desktop-app)
- [ ] 1Password Connect. For now, use [1Password/connect-sdk-go](https://github.com/1Password/connect-sdk-go).

### Cross compilation

If you want to cross compile the Go SDK while using the 1Password desktop app to authenticate, you'll need to enable CGO.

CGO can be enabled either by default if your system permits that, or deliberately if your system doesn't permit that by:

- Setting the `CGO_ENABLED` env var to 1.
- Downloading and selecting the right cross-compiler depending on your desired target and host system.

## 📖 Learn more

- [Load secrets](https://developer.1password.com/docs/sdks/load-secrets)
- [Read 1Password Environments (beta)](https://developer.1password.com/docs/sdks/environments)
- [Manage items](https://developer.1password.com/docs/sdks/manage-items)
- [Manage files](https://developer.1password.com/docs/sdks/files)
- [Share items](https://developer.1password.com/docs/sdks/share-items)
- [List vaults and items](https://developer.1password.com/docs/sdks/list-vaults-items)
- [Manage vaults](https://developer.1password.com/docs/sdks/vaults)
- [Manage vault permissions](https://developer.1password.com/docs/sdks/vault-permissions)
- [Manage groups](https://developer.1password.com/docs/sdks/groups)
- [1Password SDK concepts](https://developer.1password.com/docs/sdks/concepts)
