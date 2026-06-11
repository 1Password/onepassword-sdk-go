// This example demonstrates reading environment variables from a 1Password Environment
// using the 1Password Go SDK. See: https://1password.dev/sdks/environments/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/1password/onepassword-sdk-go"
)

func main() {
	environmentID := os.Getenv("OP_ENVIRONMENT_ID")
	if environmentID == "" {
		fmt.Fprintln(os.Stderr, "OP_ENVIRONMENT_ID is required. You can find your Environment ID from the 1Password app:")
		fmt.Fprintln(os.Stderr, `Navigate to: Developer > View Environments > Select "View Environment" for the desired environment > Click the vertical ellipsis button > Select "Copy environment ID"`)
		os.Exit(1)
	}

	client, err := newClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Read variables from the 1Password Environment
	response, err := client.Environments().GetVariables(context.Background(), environmentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetVariables failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Environment has %d variable(s):\n", len(response.Variables))
	for _, v := range response.Variables {
		fmt.Printf("  %s = %s (masked: %t)\n", v.Name, v.Value, v.Masked)
	}
}

func newClient() (*onepassword.Client, error) {
	ctx := context.Background()

	if token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN"); token != "" {
		return onepassword.NewClient(ctx,
			onepassword.WithServiceAccountToken(token),
			onepassword.WithIntegrationInfo("Environments Example", "1.0.0"),
		)
	}

	accountName := os.Getenv("OP_ACCOUNT_NAME")
	if accountName == "" {
		accountName = "YourAccountNameAsShownInTheDesktopApp"
	}
	return onepassword.NewClient(ctx,
		onepassword.WithDesktopAppIntegration(accountName),
		onepassword.WithIntegrationInfo("Environments Example", "1.0.0"),
	)
}
