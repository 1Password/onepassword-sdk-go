# Examples
This folder contains a code snippet demonstrating how to use the 1Password Go SDK for performing various operations on 1Password vaults and items. Specifically, the example showcases how to:

- Authenticate with the 1Password API using a service account token.
- List available vaults and items within those vaults.
- Retrieve a specific secret and resolve a one-time password (TOTP).
- Create a new item in a vault with multiple fields and tags.
- Update an existing item by modifying its fields and adding a new website.
- Generate different types of passwords (PIN, memorable, and random).
- Share an item with valid recipients and create a shareable link.
- Archive or delete items from the vault.
- Create and manage SSH key items.
- Create and manage Document items, including replacing and reading documents.
- Read files stored in file fields and Document items.
- Create and manage items that contain file fields, by attaching and deleting files.

## Prerequisites

1. Clone the repository and follow the steps to [get started](https://github.com/1Password/onepassword-sdk-go/blob/main/README.md#get-started).
2. Export a valid service account token as an environment variable:
    ```bash
    export OP_SERVICE_ACCOUNT_TOKEN="<your token>"
    ```
3. Set a vault ID that your service account has Read, Write, and Share access to as the `OP_VAULT_ID` environment variable:
    ```bash
    export OP_VAULT_ID="<your vault uuid>"
    ```

## How to run

To run the example file, navigate to the project root directory and run: 
```bash
go run example/main.go
```

## Terminal output

When running the example, the terminal will display:

- A list of vaults and items.
- Retrieved secrets and TOTP codes.
- Details of newly created and updated items.
- Generated passwords (PIN, memorable, random).
- A shareable link for shared items.
- SSH key attributes like public key and fingerprint.
- Document content after replacing the file.
- A list of file field items and file deletions.

These outputs show the results of vault and item operations, password generation, item sharing, and management of SSH and document items.
