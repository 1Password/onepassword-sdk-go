package main

import (
	"context"
	"fmt"
	"os"

	onepassword "github.com/1password/onepassword-sdk-go"
)

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

	item := createAndGetItem(client)
	updateItem(client, item.VaultID, item.ID)
	getAndUpdateItem(client, item.VaultID, item.ID)
	resolveSecretReference(client, item.VaultID, item.ID, "username")
}

func getAndUpdateItem(client *onepassword.Client, existingVaultID, existingItemID string) {
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

	updatedItem, err := client.Items.Update(context.Background(), item)
	if err != nil {
		panic(err)
	}

	for _, f := range updatedItem.Fields {
		if f.Title == "Details" {
			doSomethingSecret(f.Value)
		}
	}
}

func updateItem(client *onepassword.Client, existingVaultID, existingID string) {
	sectionID := "extraDetails"
	newItem := onepassword.Item{
		ID:       existingID,
		Title:    "My Login SDK 1234",
		Category: onepassword.ItemCategoryLogin,
		VaultID:  existingVaultID,
		Fields: []onepassword.ItemField{
			{
				ID:        "username",
				Value:     "Wendy_Appleseed1234",
				FieldType: onepassword.ItemFieldTypeText,
			},
			{
				ID:        "password",
				Value:     "my_weak_password1234",
				FieldType: onepassword.ItemFieldTypeConcealed,
			},
			{
				Title:     "Details",
				ID:        "myDetailsID",
				Value:     "Test Item",
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

	_, err := client.Items.Update(context.Background(), newItem)
	if err != nil {
		panic(err)
	}
}

func resolveSecretReference(client *onepassword.Client, vaultID, itemID, fieldID string) {
	// Retrieves a secret from 1Password.
	// Takes a secret reference as input and returns the secret to which it points.
	secret, err := client.Secrets.Resolve(context.Background(), fmt.Sprintf("op://%s/%s/%s", vaultID, itemID, fieldID))
	if err != nil {
		panic(err)
	}

	doSomethingSecret(secret)
}

func createAndGetItem(client *onepassword.Client) onepassword.Item {
	sectionID := "extraDetails"
	item := onepassword.Item{
		Title:    "Login created with the SDK",
		Category: onepassword.ItemCategoryLogin,
		VaultID:  "xw33qlvug6moegr3wkk5zkenoa",
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
	createdItem, err := client.Items.Create(context.Background(), item)
	if err != nil {
		panic(err)
	}

	// Retrieves the newly created item
	login, err := client.Items.Get(context.Background(), createdItem.VaultID, createdItem.ID)
	if err != nil {
		panic(err)
	}

	return login
}

// Exports the secret to the SECRET_ENV_VAR environment variable.
func doSomethingSecret(secret string) {
	err := os.Setenv("SECRET_ENV_VAR", secret)
	if err != nil {
		panic(err)
	}
}
