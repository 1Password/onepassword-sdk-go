package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

)

// [developer-docs.sdk.go.sdk-import]-start
import 	"github.com/1password/onepassword-sdk-go"
// [developer-docs.sdk.go.sdk-import]-end

func main() {
	// [developer-docs.sdk.go.client-initialization]-start
	// Connect to the 1Password desktop app
	client, err := onepassword.NewClient(context.Background(),
		onepassword.WithDesktopAppIntegration("YourAccountNameAsShownInTheDesktopApp"),
		// TODO: Set to your own integration name and version
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
	showcaseVaultOperations(client)
	showcaseBatchItemOperations(client, vaultID)

	groupID := os.Getenv("OP_GROUP_ID")
	if groupID == "" {
		panic("OP_GROUP_ID is required")
	}

	showcaseGroupPermissionOperations(client, vaultID, groupID)
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

	// Create a new item based on the structure definition above
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
	// Get the newly-created item
	login, err := client.Items().Get(context.Background(), createdItem.VaultID, createdItem.ID)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.get-item]-end

	// [developer-docs.sdk.go.get-totp-item-crud]-start
	// Get a one-time password code from an item
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

func showcaseVaultOperations(client *onepassword.Client) {

	// [developer-docs.sdk.go.create-vault]-start
	description := "This vault was created with the Go SDK."
	// Create a vault
	createParams := onepassword.VaultCreateParams{
		Title:       "Go SDK Vault",
		Description: &description,
	}

	createdVault, err := client.Vaults().Create(context.Background(), createParams)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created vault with description: %v\n", createdVault)
	// [developer-docs.sdk.go.create-vault]-end

	// [developer-docs.sdk.go.get-vault-overview]-start
	// Get a vault overview
	vaultOverview, err := client.Vaults().GetOverview(context.Background(), createdVault.ID)
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

	// [developer-docs.sdk.go.update-vault]-start
	updateParams := onepassword.VaultUpdateParams{
		Title:       nil,
		Description: nil,
	}

	name := "Go SDK Updated Vault"
	description = "Updated description from Go SDK"
	updateParams.Title = &name
	updateParams.Description = &description

	// Update the vault
	updatedVault, err := client.Vaults().Update(context.Background(), createdVault.ID, updateParams)
	if err != nil {
		panic(err)
	}
	fmt.Println("Updated Vault: ", updatedVault.Title)
	// [developer-docs.sdk.go.update-vault]-end

	// [developer-docs.sdk.go.delete-vault]-start
	// Delete a vault
	err = client.Vaults().Delete(context.Background(), createdVault.ID)
	if err != nil {
		panic(err)
	}
	fmt.Println("Deleted vault.")

	// [developer-docs.sdk.go.delete-vault]-end

	// [developer-docs.sdk.go.list-vaults]-start
	// List vaults
	vaults, err := client.Vaults().List(context.Background())
	if err != nil {
		panic(err)
	}
	for _, vault := range vaults {
		fmt.Println("VAULT ID: ", vault.ID, "VAULT NAME: ", vault.Title)
	}
	// [developer-docs.sdk.go.list-vaults]-end
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

	// Batch create all items in the same vault
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
	// Get multiple items from the same vault
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
	// Delete multiple items from the same vault
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

func showcaseGroupPermissionOperations(client *onepassword.Client, vaultID string, groupID string) {
	// Grant group permissions in a vault
	groupAccess := onepassword.GroupAccess{
		GroupID:     groupID,
		Permissions: onepassword.ReadItems,
	}
	err := client.Vaults().GrantGroupPermissions(context.Background(), vaultID, []onepassword.GroupAccess{groupAccess})
	if err != nil {
		panic(err)
	}
	fmt.Println("Granted group permissions to vault.")

	// Update group permissions in a vault
	groupVaultAccess := onepassword.GroupVaultAccess{
		GroupID:     groupID,
		VaultID:     vaultID,
		Permissions: onepassword.ReadItems | onepassword.CreateItems | onepassword.UpdateItems,
	}
	err = client.Vaults().UpdateGroupPermissions(context.Background(), []onepassword.GroupVaultAccess{groupVaultAccess})
	if err != nil {
		panic(err)
	}

	// Revoke a group's permissions in a vault
	err = client.Vaults().RevokeGroupPermissions(context.Background(), vaultID, groupID)
	if err != nil {
		panic(err)
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
