package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// [developer-docs.sdk.go.sdk-import]-start
import "github.com/1password/onepassword-sdk-go"

// [developer-docs.sdk.go.sdk-import]-end

func main() {
	// [developer-docs.sdk.go.client-initialization]-start
	// Gets your service account token from the OP_SERVICE_ACCOUNT_TOKEN environment variable.
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	// Authenticates with your service account token and connects to 1Password.
	client, err := onepassword.NewClient(context.Background(),
		onepassword.WithServiceAccountToken(token),
		// TODO: Set the following to your own integration name and version.
		onepassword.WithIntegrationInfo("My 1Password Integration", "v1.0.0"),
	)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.client-initialization]-end

	item := createAndGetItem(client)
	getAndUpdateItem(client, item.VaultID, item.ID)
	listVaultsAndItems(client, item.VaultID)
	resolveSecretReference(client, item.VaultID, item.ID, "username")
	deleteItem(client, item.VaultID, item.ID)
}

func listVaultsAndItems(client *onepassword.Client, vaultID string) {
	// [developer-docs.sdk.go.list-vaults]-start
	vaults, err := client.Vaults.ListAll(context.Background())
	if err != nil {
		panic(err)
	}
	for {
		vault, err := vaults.Next()
		if errors.Is(err, onepassword.ErrorIteratorDone) {
			break
		} else if err != nil {
			panic(err)
		}

		fmt.Printf("%s %s\n", vault.ID, vault.Title)
	}
	// [developer-docs.sdk.go.list-vaults]-end

	// [developer-docs.sdk.go.list-items]-start
	items, err := client.Items.ListAll(context.Background(), vaultID)
	if err != nil {
		panic(err)
	}
	for {
		item, err := items.Next()
		if errors.Is(err, onepassword.ErrorIteratorDone) {
			break
		} else if err != nil {
			panic(err)
		}
		fmt.Printf("%s %s\n", item.ID, item.Title)
	}
	// [developer-docs.sdk.go.list-items]-end
}

func getAndUpdateItem(client *onepassword.Client, existingVaultID, existingItemID string) {
	// [developer-docs.sdk.go.update-item]-start
	// Retrieves the newly created item
	item, err := client.Items.Get(context.Background(), existingVaultID, existingItemID)
	if err != nil {
		panic(err)
	}

	// Finds the field named "Details" and edits its value
	for i := range item.Fields {
		if item.Fields[i].Title == "Details" {
			item.Fields[i].Value = "updated details"
		}
	}
	item.Title = "New Title"

	updatedItem, err := client.Items.Put(context.Background(), item)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.update-item]-end

	for _, f := range updatedItem.Fields {
		if f.Title == "Details" {
			fmt.Println(f.Value)
		}
	}
}

func resolveSecretReference(client *onepassword.Client, vaultID, itemID, fieldID string) {
	// [developer-docs.sdk.go.resolve-secret]-start
	// Retrieves a secret from 1Password.
	// Takes a secret reference as input and returns the secret to which it points.
	secret, err := client.Secrets.Resolve(context.Background(), fmt.Sprintf("op://%s/%s/%s", vaultID, itemID, fieldID))
	if err != nil {
		panic(err)
	}
	fmt.Println(secret)
	// [developer-docs.sdk.go.resolve-secret]-end
}

func createAndGetItem(client *onepassword.Client) onepassword.Item {
	// [developer-docs.sdk.go.create-item]-start
	sectionID := "extraDetails"
	itemParams := onepassword.ItemCreateParams{
		Title:    "Login created with the SDK",
		Category: onepassword.ItemCategoryLogin,
		VaultID:  "7turaasywpymt3jecxoxk5roli",
		Fields: []onepassword.ItemField{
			{
				ID:        "username",
				Title:     "username",
				Value:     "Wendy_Appleseed",
				FieldType: onepassword.ItemFieldTypeText,
			},
			{
				ID:        "password",
				Title:     "password",
				Value:     "my_weak_password123",
				FieldType: onepassword.ItemFieldTypeConcealed,
			},
			{
				ID:        "onetimepassword",
				Title:     "one-time password",
				Value:     "otpauth://totp/my-example-otp?secret=jncrjgbdjnrncbjsr&issuer=1Password",
				SectionID: &sectionID,
				FieldType: onepassword.ItemFieldTypeTOTP,
			},
			{
				ID:        "uniqueId",
				Title:     "Web address",
				Value:     "1password.com",
				FieldType: onepassword.ItemFieldTypeText,
				SectionID: &sectionID,
			},
		},
		Sections: []onepassword.ItemSection{
			{
				ID:    sectionID,
				Title: "Extra Details",
			},
		},
	}

	// Creates a new item based on the structure definition above
	createdItem, err := client.Items.Create(context.Background(), itemParams)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.create-item]-end

	// [developer-docs.sdk.go.get-item]-start
	// Retrieves the newly created item
	login, err := client.Items.Get(context.Background(), createdItem.VaultID, createdItem.ID)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.get-item]-end

	// [developer-docs.sdk.go.get-totp-item-crud]-start
	// Retrieve TOTP code from an item
	for _, f := range login.Fields {
		if f.FieldType == onepassword.ItemFieldTypeTOTP {
			OTPFieldDetails := f.Details.OTP()
			if OTPFieldDetails.ErrorMessage == nil {
				fmt.Println(*OTPFieldDetails.Code)
			} else {
				panic(*OTPFieldDetails.ErrorMessage)
			}
		}
	}
	// [developer-docs.sdk.go.get-totp-item-crud]-end

	return login
}

func deleteItem(client *onepassword.Client, vaultID string, itemID string) {
	// [developer-docs.sdk.go.delete-item]-start
	err := client.Items.Delete(context.Background(), vaultID, itemID)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.delete-item]-end
}
