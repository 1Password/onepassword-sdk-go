package main

import (
	"context"
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

	resolveSecretReference(client)
	createAndGetItem(client)
}

func resolveSecretReference(client *onepassword.Client) {
	// Retrieves a secret from 1Password.
	// Takes a secret reference as input and returns the secret to which it points.
	secret, err := client.Secrets.Resolve(context.Background(), "op://vault/item/field")
	if err != nil {
		panic(err)
	}

	doSomethingSecret(secret)
}

func createAndGetItem(client *onepassword.Client) {
	sectionID := "extra_details"
	item := onepassword.Item{
		ID:       "",
		Title:    "My Login",
		Category: onepassword.ItemCategoryLogin,
		VaultID:  "qw33qlyug6moear3wkk9zkemui",
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
				ID:        "unique_id",
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

	if len(login.Fields) > 0 {
		doSomethingSecret(login.Fields[0].Value)
	}
}

// Exports the secret to the SECRET_ENV_VAR environment variable.
func doSomethingSecret(secret string) {
	err := os.Setenv("SECRET_ENV_VAR", secret)
	if err != nil {
		panic(err)
	}
}
