package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// [developer-docs.sdk.go.sdk-import]-start
import 	"github.com/1password/onepassword-sdk-go"
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
	createSSHKeyItem(client)
	createAndReplaceDocumentItem(client)
	createAndAttachAndDeleteFileFieldItem(client)
	getAndUpdateItem(client, item.VaultID, item.ID)
	listVaultsAndItems(client, item.VaultID)
	generatePasswords()
	resolveSecretReference(client, item.VaultID, item.ID, "username")
	resolveTOTPSecretReference(client, item.VaultID, item.ID, "TOTP_onetimepassword")
	sharelink := generateItemSharing(client, item.VaultID, item.ID)
	fmt.Println(sharelink)
	deleteItem(client, item.VaultID, item.ID)
}

func listVaultsAndItems(client *onepassword.Client, vaultID string) {
	// [developer-docs.sdk.go.list-vaults]-start
	vaults, err := client.Vaults().ListAll(context.Background())
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
	items, err := client.Items().ListAll(context.Background(), vaultID)
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

// NOTE: just for the sake of archiving it. This is because the SDK
// NOTE: only works with active items, so archiving and then deleting
// NOTE: is not yet possible.
//
//lint:ignore U1000 NOTE: this is in a separate function to avoid creating a new item
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

	// [developer-docs.sdk.go.create-sshkey-item]-start
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
}

func createAndReplaceDocumentItem(client *onepassword.Client) {
	vaultID := os.Getenv("OP_VAULT_ID")

	// [developer-docs.sdk.go.create-document-item]-start
	fileContent, err := os.ReadFile("./example/file.txt")
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
	file2Content, err := os.ReadFile("./example/file2.txt")
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
}

func createAndAttachAndDeleteFileFieldItem(client *onepassword.Client) {
	vaultID := os.Getenv("OP_VAULT_ID")

	// [developer-docs.sdk.go.create-item-with-file-field]-start
	fileContent, err := os.ReadFile("./example/file.txt")
	if err != nil {
		panic(err)
	}
	sectionID := "extraDetails"
	itemParams := onepassword.ItemCreateParams{
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
	}

	// Create the file field item
	item, err := client.Items().Create(context.Background(), itemParams)
	if err != nil {
		panic(err)
	}
	// [developer-docs.sdk.go.create-item-with-file-field]-end

	// [developer-docs.sdk.go.attach-file-field-item]-start
	file2Content, err := os.ReadFile("./example/file2.txt")
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
}
