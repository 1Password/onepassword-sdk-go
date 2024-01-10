package onepassword

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"unicode"
)

const (
	SDKSemverVersion = "0010001" // v0.1.0

	DefaultIntegrationName    = "Unknown"
	DefaultIntegrationVersion = "Unknown"

	SDKLanguage           = "Go"
	DefaultRequestLibrary = "net/http"
)

// OpClient The client instance.
type OpClient struct {
	config  ClientConfig
	context context.Context

	Secrets SecretsAPI
}

type ClientConfig struct {
	SAToken               string `json:"saToken"`
	Language              string `json:"language"`
	SDKVersion            string `json:"sdkVersion"`
	IntegrationName       string `json:"integrationName"`
	IntegrationVersion    string `json:"integrationVersion"`
	RequestLibraryName    string `json:"requestLibraryName"`
	RequestLibraryVersion string `json:"requestLibraryVersion"`
	SystemOS              string `json:"os"`
	SystemOSVersion       string `json:"osVersion"`
	SystemArch            string `json:"arch"`
}

// NewClient returns a 1Password Go SDK client.
func NewClient(opts ...ClientOption) (*OpClient, error) {
	client := OpClient{
		context: context.Background(),
	}

	for _, opt := range opts {
		err := opt(&client)
		if err != nil {
			return nil, err
		}
	}

	if len(client.config.SAToken) == 0 {
		return nil, errors.New("cannot create a client without specifying a Service Account Token")
	}

	if len(client.config.IntegrationVersion) == 0 || len(client.config.IntegrationName) == 0 {
		return nil, errors.New("cannot create a client without defining an app name and version. If you don't want to specify any, use the provided constants: 'DefaultIntegrationName', 'DefaultIntegrationVersion'")
	}

	const defaultOSVersion = "0.0.0"

	client.config.Language = SDKLanguage
	client.config.SDKVersion = SDKSemverVersion
	client.config.RequestLibraryName = DefaultRequestLibrary
	client.config.RequestLibraryVersion = runtime.Version()
	client.config.SystemOS = runtime.GOOS
	client.config.SystemArch = runtime.GOARCH
	// TODO: add logic for determining this for all systems in a different PR.
	client.config.SystemOSVersion = defaultOSVersion

	clientID, err := InitClient(client.context, client.config)
	if err != nil {
		return nil, fmt.Errorf("error initializing client: %w", err)
	}

	client.Secrets = SecretsSource{
		clientID: *clientID,
	}
	runtime.SetFinalizer(&client, func(f *OpClient) {
		ReleaseClient(*clientID)
	})
	return &client, nil
}

type ClientOption func(config *OpClient) error

func WithServiceAccountToken(token string) ClientOption {
	return func(c *OpClient) error {
		c.config.SAToken = token
		return nil
	}
}

func WithIntegrationInfo(name string, version string) ClientOption {
	const (
		integrationNameMaxLen    = 40
		integrationVersionMaxLen = 20
		allowedSymbols           = "_- .,"
	)
	return func(c *OpClient) error {
		if len(name) > integrationNameMaxLen {
			return fmt.Errorf("integration name can't be longer than 40 characters")
		}

		if len(version) > integrationVersionMaxLen {
			return fmt.Errorf("integration version can't be longer than 20 characters")
		}

		for _, r := range name {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune(allowedSymbols, r) {
				return fmt.Errorf("integration name can only contain digits, letters and allowed symbols")
			}
		}

		for _, r := range version {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune(allowedSymbols, r) {
				return fmt.Errorf("integration version can only contain digits, letters and allowed symbols")
			}
		}

		c.config.IntegrationName = name
		c.config.IntegrationVersion = version
		return nil
	}
}

func WithContext(ctx context.Context) ClientOption {
	return func(c *OpClient) error {
		c.context = ctx
		return nil
	}
}
