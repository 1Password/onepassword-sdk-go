package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/1password/onepassword-sdk-go"
)

// [developer-docs.sdk.go.sdk-import]-start

// [developer-docs.sdk.go.sdk-import]-end

func main() {
	// [developer-docs.sdk.go.client-initialization]-start
	// Connects to the 1Password Desktop app.
	client, err := onepassword.NewClient(context.Background(),
		onepassword.WithDesktopAppIntegration("YourAccountNameAsShownInTheDesktopApp"),
		// TODO: Set the following to your own integration name and version.
		onepassword.WithIntegrationInfo("My 1Password Integration", "v1.0.0"),
	)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.client-initialization]-end

	vaultID := os.Getenv("OP_VAULT_ID")
	if vaultID == "" {
		panic("OP_VAULT_ID is required")
	}

	createAndGetItem(client, vaultID)
	showcaseVaultOperations(client, vaultID)
	showcaseBatchItemOperations(client, vaultID)

	groupID := os.Getenv("OP_GROUP_ID")
	if groupID == "" {
		panic("OP_GROUP_ID is required")
	}

	// [developer-docs.sdk.go.get-group]-start
	// Get a group
	group, err := client.Groups().Get(context.Background(), groupID, onepassword.GroupGetParams{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Group details: %v\n", group)
	// [developer-docs.sdk.go.get-group]-end
}

func createAndGetItem(client *onepassword.Client, vaultID string) onepassword.Item {
	// [developer-docs.sdk.go.create-item]-start
	sectionID := "extraDetails"
	itemParams := onepassword.ItemCreateParams{
		Title:    "Login created with the SDK",
		Category: onepassword.ItemCategoryLogin,
		VaultID:  vaultID,
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
		},
		Sections: []onepassword.ItemSection{
			{
				ID:    sectionID,
				Title: "Extra Details",
			},
		},
		Tags: []string{"test tag1", "test tag 2"},
		Websites: []onepassword.Website{
			{
				URL:              "1password.com",
				AutofillBehavior: onepassword.AutofillBehaviorAnywhereOnWebsite,
				Label:            "my custom website",
			},
		},
	}

	// Creates a new item based on the structure definition above
	createdItem, err := client.Items().Create(context.Background(), itemParams)
	if err != nil {
		panic(err)
	}
	p, err := json.Marshal(createdItem)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(p))
	// [developer-docs.sdk.go.create-item]-end

	// [developer-docs.sdk.go.get-item]-start
	// Retrieves the newly created item
	login, err := client.Items().Get(context.Background(), createdItem.VaultID, createdItem.ID)
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

func showcaseVaultOperations(client *onepassword.Client, vaultID string) {
	// [developer-docs.sdk.go.list-vaults]-start
	// List vaults
	vaults, err := client.Vaults().List(context.Background())
	if err != nil {
		panic(err)
	}
	for _, vault := range vaults {
		fmt.Println("VAULT ID: ", vault.ID)
	}
	// [developer-docs.sdk.go.list-vaults]-end

	// [developer-docs.sdk.go.get-vault-overview]-start
	// Get vault overview
	vaultOverview, err := client.Vaults().GetOverview(context.Background(), vaultID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Vault overview: %v\n", vaultOverview)
	// [developer-docs.sdk.go.get-vault-overview]-end

	// [developer-docs.sdk.go.get-vault-details]-start
	// Get vault details
	vault, err := client.Vaults().Get(context.Background(), vaultOverview.ID, onepassword.VaultGetParams{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Vault details: %v\n", vault)
	// [developer-docs.sdk.go.get-vault-details]-end
}

func showcaseBatchItemOperations(client *onepassword.Client, vaultID string) {
	// [developer-docs.sdk.go.batch-create-items]-start
	sectionID := "extraDetails"
	var itemsToCreate []onepassword.ItemCreateParams
	for i := 1; i <= 3; i++ {
		itemsToCreate = append(itemsToCreate, onepassword.ItemCreateParams{
			Title:    fmt.Sprintf("Login %d created with the SDK", i),
			Category: onepassword.ItemCategoryLogin,
			VaultID:  vaultID,
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
			},
			Sections: []onepassword.ItemSection{
				{
					ID:    sectionID,
					Title: "Extra Details",
				},
			},
			Tags: []string{"test tag1", "test tag 2"},
			Websites: []onepassword.Website{
				{
					URL:              "1password.com",
					AutofillBehavior: onepassword.AutofillBehaviorAnywhereOnWebsite,
					Label:            "my custom website",
				},
			},
		})
	}

	// Create all items in the same vault in a single batch
	batchCreateResponse, err := client.Items().CreateAll(context.Background(), vaultID, itemsToCreate)
	if err != nil {
		panic(err)
	}

	var itemIDs []string
	for _, res := range batchCreateResponse.IndividualResponses {
		if res.Content != nil {
			fmt.Printf("Created Item %q (%s)\n", res.Content.Title, res.Content.ID)
			itemIDs = append(itemIDs, res.Content.ID)
		} else if res.Error != nil {
			fmt.Printf("[Batch create] Something went wrong: %s\n", res.Error)
		}
	}
	// [developer-docs.sdk.go.batch-create-items]-end

	// [developer-docs.sdk.go.batch-get-items]-start
	// Get multiple items form the same vault in a single batch
	batchGetResponse, err := client.Items().GetAll(context.Background(), vaultID, itemIDs)
	if err != nil {
		panic(err)
	}
	for _, res := range batchGetResponse.IndividualResponses {
		if res.Content != nil {
			fmt.Printf("Obtained Item %q (%s)\n", res.Content.Title, res.Content.ID)
		} else if res.Error != nil {
			fmt.Printf("[Batch get] Something went wrong: %s\n", res.Error)
		}
	}
	// [developer-docs.sdk.go.batch-get-items]-end

	// [developer-docs.sdk.go.batch-delete-items]-start
	// Delete multiple items from the same vault in a single batch
	batchDeleteResponse, err := client.Items().DeleteAll(context.Background(), vaultID, itemIDs)
	if err != nil {
		panic(err)
	}
	for id, res := range batchDeleteResponse.IndividualResponses {
		if res.Error != nil {
			fmt.Printf("[Batch delete] Something went wrong: %s\n", res.Error)
		} else {
			fmt.Printf("Deleted item %s\n", id)
		}
	}
	// [developer-docs.sdk.go.batch-delete-items]-end
}
