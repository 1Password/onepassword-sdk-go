# Read 1Password Environments

This example shows how to read environment variables from a [1Password Environment](https://developer.1password.com/docs/sdks/environments/) using the Go SDK.

## Prerequisites

1. Clone the repository and follow the [get started](https://github.com/1Password/onepassword-sdk-go#-get-started) steps.
2. Get your Environment ID from the 1Password app:
   - Open the 1Password desktop app and unlock it.
   - Go to **Developer** > **View Environments**.
   - Select **View environment** next to the Environment you want to use.
   - Select **Manage environment** > **Copy environment ID**.

   **Tip:** To see the Environment ID in the 'manage environment' dropdown, enable **Show debugging tools** under **Application** > **Settings** > **Advanced**. You may need to log out and back in for this setting to take effect.

## Authentication

Use either a service account or the 1Password desktop app.

**Service account:**

```bash
export OP_SERVICE_ACCOUNT_TOKEN="<your token>"
```

**Desktop app:** Set your account name as shown in the desktop app:

```bash
export OP_ACCOUNT_NAME="<your account name>"
```

## How to run

Set the Environment ID, then run the example from the project root:

```bash
export OP_ENVIRONMENT_ID="<your environment id>"
go run example/environments/main.go
```

## Output

The program prints each variable in the environment, including its name, value, and whether its value is hidden by default in the 1Password app.

## Documentation

- [Read 1Password Environments (SDK guide)](https://developer.1password.com/docs/sdks/environments/)
