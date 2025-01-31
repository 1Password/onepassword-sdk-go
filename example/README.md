# Examples

This folder contains a code snippet that demonstrates how to use the 1Password Go SDK to retrieve a secret from 1Password and export it as an environment variable. 

## Prerequisites

1. Clone the repository and follow the steps to [get started](https://github.com/1Password/onepassword-sdk-go/blob/main/README.md#get-started).
2. Make sure to export a valid service account token. For example:
	```bash
	export OP_SERVICE_ACCOUNT_TOKEN="<your token>"
	```
3. Set a vault ID that your service account has Read, Write and Share access as the `OP_VAULT_ID` environment variable:
    ```bash
    export OP_VAULT_ID="<your vault uuid>"
    ```
