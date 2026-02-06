package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/1password/onepassword-sdk-go"
)

// [developer-docs.sdk.go.sdk-import]-start

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

	groupID := os.Getenv("OP_GROUP_ID")
	if groupID == "" {
		panic("OP_GROUP_ID is required")
	}

	item := createAndGetItem(client)
	createSSHKeyItem(client)
	createAndReplaceDocumentItem(client)
	createAndAttachAndDeleteFileFieldItem(client)
	getAndUpdateItem(client, item.VaultID, item.ID)
	listVaultsAndItems(client, item.VaultID)
	showcaseVaultOperations(client)
	showcaseBatchItemOperations(client, item.VaultID)
	showcaseGroupPermissionOperations(client, item.VaultID, groupID)
	generatePasswords()
	resolveSecretReference(client, item.VaultID, item.ID, "username")
	resolveBulkSecretReferences(client, item.VaultID, item.ID, "username", "password")
	resolveTOTPSecretReference(client, item.VaultID, item.ID, "TOTP_onetimepassword")
	sharelink := generateItemSharing(client, item.VaultID, item.ID)
	fmt.Println(sharelink)
	archiveItem(client, item.VaultID, item.ID)
	deleteItem(client, item.VaultID, item.ID)
}

func showcaseVaultOperations(client *onepassword.Client) {

	// [developer-docs.sdk.go.create-vault]-start
	description := "This vault was created with the Go SDK."
	// Create a vault with a description
	create_params := onepassword.VaultCreateParams{
		Title:       "Go SDK Vault",
		Description: &description,
	}

	created_vault, err := client.Vaults().Create(context.Background(), create_params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created vault with description: %v\n", created_vault)
	// [developer-docs.sdk.go.create-vault]-start

	// [developer-docs.sdk.go.get-vault-overview]-start
	// Get vault overview
	vaultOverview, err := client.Vaults().GetOverview(context.Background(), created_vault.ID)
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
	updated_vault, err := client.Vaults().Update(context.Background(), created_vault.ID, updateParams)
	if err != nil {
		panic(err)
	}
	fmt.Println("Updated Vault: ", updated_vault.Title)
	// [developer-docs.sdk.go.update-vault]-end

	// [developer-docs.sdk.go.delete-vault]-start
	// Delete vault
	client.Vaults().Delete(context.Background(), created_vault.ID)
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

func showcaseGroupPermissionOperations(client *onepassword.Client, vaultID string, groupID string) {
	// Grant group permissions to a vault.
	groupAccess := onepassword.GroupAccess{
		GroupID:     groupID,
		Permissions: onepassword.ReadItems,
	}
	err := client.Vaults().GrantGroupPermissions(context.Background(), vaultID, []onepassword.GroupAccess{groupAccess})
	if err != nil {
		panic(err)
	}
	fmt.Println("Granted group permissions to vault.")

	// update group permissions for vaults.
	groupVaultAccess := onepassword.GroupVaultAccess{
		GroupID:     groupID,
		VaultID:     vaultID,
		Permissions: onepassword.ReadItems | onepassword.CreateItems | onepassword.UpdateItems,
	}
	err = client.Vaults().UpdateGroupPermissions(context.Background(), []onepassword.GroupVaultAccess{groupVaultAccess})
	if err != nil {
		panic(err)
	}

	// Revoke group permissions from a vault.
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

func listVaultsAndItems(client *onepassword.Client, vaultID string) {
	// [developer-docs.sdk.go.list-vaults]-start
	vaults, err := client.Vaults().List(context.Background())
	if err != nil {
		panic(err)
	}
	for _, vault := range vaults {
		fmt.Printf("%+v\n", vault)
	}
	// [developer-docs.sdk.go.list-vaults]-end

	// [developer-docs.sdk.go.list-items]-start
	overviews, err := client.Items().List(context.Background(), vaultID)
	if err != nil {
		panic(err)
	}
	for _, overview := range overviews {
		fmt.Printf("%s %s\n", overview.ID, overview.Title)
	}
	// [developer-docs.sdk.go.list-items]-end

	// [developer-docs.sdk.go.use-item-filters]-start
	archivedOverviews, err := client.Items().List(context.Background(), vaultID,
		onepassword.NewItemListFilterTypeVariantByState(
			&onepassword.ItemListFilterByStateInner{
				Active:   false,
				Archived: true,
			},
		),
	)
	if err != nil {
		panic(err)
	}
	for _, overview := range archivedOverviews {
		fmt.Printf("%s %s\n", overview.ID, overview.Title)
	}
	// [developer-docs.sdk.go.use-item-filters]-end

}

func getAndUpdateItem(client *onepassword.Client, existingVaultID, existingItemID string) {
	// [developer-docs.sdk.go.update-item]-start
	// Retrieves the newly created item
	item, err := client.Items().Get(context.Background(), existingVaultID, existingItemID)
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
	item.Websites = append(item.Websites, onepassword.Website{
		URL:              "2password.com",
		Label:            "my second custom website",
		AutofillBehavior: onepassword.AutofillBehaviorNever,
	})

	updatedItem, err := client.Items().Put(context.Background(), item)
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
	// [developer-docs.sdk.go.validate-secret-reference]-start
	// Validate your secret reference
	err := onepassword.Secrets.ValidateSecretReference(context.Background(), fmt.Sprintf("op://%s/%s/%s", vaultID, itemID, fieldID))
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.validate-secret-reference]-end

	// [developer-docs.sdk.go.resolve-secret]-start
	// Retrieves a secret from 1Password.
	// Takes a secret reference as input and returns the secret to which it points.
	secret, err := client.Secrets().Resolve(context.Background(), fmt.Sprintf("op://%s/%s/%s", vaultID, itemID, fieldID))
	if err != nil {
		panic(err)
	}
	fmt.Println(secret)
	// [developer-docs.sdk.go.resolve-secret]-end
}

func resolveBulkSecretReferences(client *onepassword.Client, vaultID, itemID, fieldID, fieldID2 string) {
	// [developer-docs.sdk.go.resolve-bulk-secret]-start
	// Retrieves multiple secrets from 1Password.
	// Takes multiple secret references as input and returns the secret to which it points.
	secret, _ := client.Secrets().ResolveAll(
		context.Background(),
		[]string{
			fmt.Sprintf("op://%s/%s/%s", vaultID, itemID, fieldID),
			fmt.Sprintf("op://%s/%s/%s", vaultID, itemID, fieldID2),
		},
	)
	for _, s := range secret.IndividualResponses {
		if s.Error != nil {
			panic(string(s.Error.Type))
		}
		fmt.Println(s.Content.Secret)
	}
	// [developer-docs.sdk.go.resolve-bulk-secret]-end
}

func resolveTOTPSecretReference(client *onepassword.Client, vaultID, itemID, fieldID string) {
	// [developer-docs.sdk.go.resolve-totp-code]-start
	// Retrieves a TOTP code from 1Password.
	code, err := client.Secrets().Resolve(context.Background(), fmt.Sprintf("op://%s/%s/%s?attribute=totp", vaultID, itemID, fieldID))
	if err != nil {
		panic(err)
	}
	fmt.Println(code)
	// [developer-docs.sdk.go.resolve-totp-code]-end
}

func createAndGetItem(client *onepassword.Client) onepassword.Item {
	vaultID := os.Getenv("OP_VAULT_ID")

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

func deleteItem(client *onepassword.Client, vaultID string, itemID string) {
	// [developer-docs.sdk.go.delete-item]-start
	// Delete a item from your vault.
	err := client.Items().Delete(context.Background(), vaultID, itemID)
	if err != nil {
		panic(err)
	}

	// [developer-docs.sdk.go.delete-item]-end
}

func generatePasswords() {
	// [developer-docs.sdk.go.generate-pin-password]-start
	pinPassword, err := onepassword.Secrets.GeneratePassword(context.Background(), onepassword.NewPasswordRecipeTypeVariantPin(&onepassword.PasswordRecipePinInner{Length: 10}))
	if err != nil {
		panic(err)
	}
	fmt.Println(pinPassword.Password)
	// [developer-docs.sdk.go.generate-pin-password]-end

	// [developer-docs.sdk.go.generate-random-password]-start
	randomPassword, err := onepassword.Secrets.GeneratePassword(context.Background(), onepassword.NewPasswordRecipeTypeVariantRandom(&onepassword.PasswordRecipeRandomInner{
		IncludeDigits:  true,
		IncludeSymbols: true,
		Length:         10,
	}))
	if err != nil {
		panic(err)
	}
	fmt.Println(randomPassword.Password)
	// [developer-docs.sdk.go.generate-random-password]-end

	// [developer-docs.sdk.go.generate-memorable-password]-start
	memorablePassword, err := onepassword.Secrets.GeneratePassword(context.Background(), onepassword.NewPasswordRecipeTypeVariantMemorable(&onepassword.PasswordRecipeMemorableInner{
		SeparatorType: onepassword.SeparatorTypeCommas,
		WordListType:  onepassword.WordListTypeFullWords,
		Capitalize:    true,
		WordCount:     10,
	}))
	if err != nil {
		panic(err)
	}
	fmt.Println(memorablePassword.Password)
	// [developer-docs.sdk.go.generate-memorable-password]-end
}

func archiveItem(client *onepassword.Client, vaultID string, itemID string) {
	// [developer-docs.sdk.go.archive-item]-start
	// Archive a item from your vault.
	err := client.Items().Archive(context.Background(), vaultID, itemID)

	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.archive-item]-end
}

func generateItemSharing(client *onepassword.Client, vaultID string, itemID string) string {
	// [developer-docs.sdk.go.item-share-get-item]-start
	item, err := client.Items().Get(context.Background(), vaultID, itemID)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.item-share-get-item]-end

	// [developer-docs.sdk.go.item-share-get-account-policy]-start
	accountPolicy, err := client.Items().Shares().GetAccountPolicy(context.Background(), item.VaultID, item.ID)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.item-share-get-account-policy]-end

	// [developer-docs.sdk.go.item-share-validate-recipients]-start
	recipients, err := client.Items().Shares().ValidateRecipients(context.Background(), accountPolicy, []string{"helloworld@agilebits.com"})
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.item-share-validate-recipients]-end

	// [developer-docs.sdk.go.item-share-create-share]-start
	shareLink, err := client.Items().Shares().Create(context.Background(), item, accountPolicy, onepassword.ItemShareParams{
		Recipients:  recipients,
		ExpireAfter: &accountPolicy.DefaultExpiry,
		OneTimeOnly: false,
	})
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.item-share-create-share]-end

	return shareLink
}

func createSSHKeyItem(client *onepassword.Client) {
	// [developer-docs.sdk.go.create-sshkey-item]-start
	// Generate the RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	// Encode the data into PEM format
	sshKeyPEMBytes := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}))

	vaultID := os.Getenv("OP_VAULT_ID")

	sectionID := "extraDetails"
	itemParams := onepassword.ItemCreateParams{
		Title:    "SSH Key Item Created With Go SDK",
		Category: onepassword.ItemCategorySSHKey,
		VaultID:  vaultID,
		Fields: []onepassword.ItemField{
			{
				ID:        "private_key",
				Title:     "private key",
				Value:     sshKeyPEMBytes,
				FieldType: onepassword.ItemFieldTypeSSHKey,
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
	createdItem, err := client.Items().Create(context.Background(), itemParams)
	if err != nil {
		panic(err)
	}

	// Fetch all SSH key attributes
	fmt.Println(createdItem.Fields[0].Value)
	if sshAttributes := createdItem.Fields[0].Details.SSHKey(); sshAttributes != nil {
		fmt.Println(createdItem.Fields[0].Details.SSHKey().PublicKey)
		fmt.Println(createdItem.Fields[0].Details.SSHKey().Fingerprint)
		fmt.Println(createdItem.Fields[0].Details.SSHKey().KeyType)
	}
	// [developer-docs.sdk.go.create-sshkey-item]-end
	err = client.Items().Delete(context.Background(), createdItem.VaultID, createdItem.ID)
	if err != nil {
		panic(err)
	}
}

func createAndReplaceDocumentItem(client *onepassword.Client) {
	vaultID := os.Getenv("OP_VAULT_ID")

	// [developer-docs.sdk.go.create-document-item]-start
	fileContent, err := os.ReadFile("./example/service_account/file.txt")
	if err != nil {
		panic(err)
	}
	// Create the document item
	documentItem, err := client.Items().Create(context.Background(), onepassword.ItemCreateParams{
		Title:    "Document Item Created With Go SDK",
		Category: onepassword.ItemCategoryDocument,
		VaultID:  vaultID,
		Document: &onepassword.DocumentCreateParams{
			Name:    "file.txt",
			Content: fileContent,
		},
	})
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.create-document-item]-end

	// [developer-docs.sdk.go.replace-document-item]-start
	// Replace the document item
	file2Content, err := os.ReadFile("./example/service_account/file2.txt")
	if err != nil {
		panic(err)
	}
	replacedDocItem, err := client.Items().Files().ReplaceDocument(context.Background(), documentItem, onepassword.DocumentCreateParams{
		Name:    "file2.txt",
		Content: file2Content,
	})
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.replace-document-item]-end

	// [developer-docs.sdk.go.read-document-item]-start
	// Read the document item
	content, err := client.Items().Files().Read(context.Background(), replacedDocItem.VaultID, replacedDocItem.ID, *replacedDocItem.Document)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.read-document-item]-end
	fmt.Println(string(content))

	err = client.Items().Delete(context.Background(), replacedDocItem.VaultID, replacedDocItem.ID)
	if err != nil {
		panic(err)
	}
}

func createAndAttachAndDeleteFileFieldItem(client *onepassword.Client) {
	vaultID := os.Getenv("OP_VAULT_ID")
	sectionID := "extraDetails"

	// [developer-docs.sdk.go.create-item-with-file-field]-start
	fileContent, err := os.ReadFile("./example/service_account/file.txt")
	if err != nil {
		panic(err)
	}
	// Create the File Field item
	item, err := client.Items().Create(context.Background(), onepassword.ItemCreateParams{
		Title:    "Login with File Field created with SDK",
		Category: onepassword.ItemCategoryLogin,
		VaultID:  vaultID,
		Sections: []onepassword.ItemSection{
			{
				ID: sectionID,
			},
		},
		Files: []onepassword.FileCreateParams{
			{
				Name:      "file.txt",
				Content:   fileContent,
				SectionID: sectionID,
				FieldID:   "file_field",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.create-item-with-file-field]-end

	// [developer-docs.sdk.go.read-file-field]-start
	// Read the file field from an item
	retrievedFileContent, err := client.Items().Files().Read(context.Background(), item.VaultID, item.ID, item.Files[0].Attributes)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.read-file-field]-end
	fmt.Println(string(retrievedFileContent))

	// [developer-docs.sdk.go.attach-file-field-item]-start
	file2Content, err := os.ReadFile("./example/service_account/file2.txt")
	if err != nil {
		panic(err)
	}

	// Attach a file field to an item
	newItem, err := client.Items().Files().Attach(context.Background(), item, onepassword.FileCreateParams{
		Name:      "file2.txt",
		Content:   file2Content,
		SectionID: sectionID,
		FieldID:   "new_file_field",
	})
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.attach-file-field-item]-end

	// [developer-docs.sdk.go.delete-file-field-item]-start
	// Delete a file field from an item
	updatedItemWithDeletedFile, err := client.Items().Files().Delete(context.Background(), newItem, newItem.Files[0].SectionID, newItem.Files[0].FieldID)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.delete-file-field-item]-end
	fmt.Println(len(updatedItemWithDeletedFile.Files))

	err = client.Items().Delete(context.Background(), updatedItemWithDeletedFile.VaultID, updatedItemWithDeletedFile.ID)
	if err != nil {
		panic(err)
	}
}

//lint:ignore U1000 NOTE: this is just to showcase how to instantiate custom ItemFields
func generateSpecialItemFields() []onepassword.ItemField {
	sectionID := "extraDetails"

	// [developer-docs.sdk.go.address-field-type]-start
	address := onepassword.NewItemFieldDetailsTypeVariantAddress(&onepassword.AddressFieldDetails{
		Street:  "123 Main St",
		City:    "Anytown",
		State:   "CA",
		Zip:     "12345",
		Country: "USA",
	})
	addressField := onepassword.ItemField{
		ID:        "address",
		Title:     "Address",
		SectionID: &sectionID,
		FieldType: onepassword.ItemFieldTypeAddress,
		Value:     "",
		Details:   &address,
	}
	// [developer-docs.sdk.go.address-field-type]-end
	return []onepassword.ItemField{
		addressField,
		// [developer-docs.sdk.go.date-field-type]-start
		{
			ID:        "date",
			Title:     "Date",
			SectionID: &sectionID,
			FieldType: onepassword.ItemFieldTypeDate,
			Value:     "1998-03-15",
		},
		// [developer-docs.sdk.go.date-field-type]-end
		// [developer-docs.sdk.go.month-year-field-type]-start
		{
			ID:        "month_year",
			Title:     "Month Year",
			SectionID: &sectionID,
			FieldType: onepassword.ItemFieldTypeMonthYear,
			Value:     "03/1998",
		},
		// [developer-docs.sdk.go.month-year-field-type]-end
		// [developer-docs.sdk.go.reference-field-type]-start
		{
			ID:        "reference",
			Title:     "Reference",
			FieldType: onepassword.ItemFieldTypeReference,
			SectionID: &sectionID,
			Value:     "f43hnkatjllm5fsfsmgaqdhv7a",
		},
		// [developer-docs.sdk.go.reference-field-type]-end
		// [developer-docs.sdk.go.totp-field-type]-start
		{
			ID:        "onetimepassword",
			Title:     "One-Time Password URL",
			SectionID: &sectionID,
			FieldType: onepassword.ItemFieldTypeTOTP,
			Value:     "otpauth://totp/my-example-otp?secret=jncrjgbdjnrncbjsr&issuer=1Password",
		},
		// [developer-docs.sdk.go.totp-field-type]-end
	}
}
