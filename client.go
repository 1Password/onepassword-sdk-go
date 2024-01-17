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

// `OpClient` represents an instance of the 1Password Go SDK client.
type OpClient struct {
	config  ClientConfig
	Secrets SecretsAPI
}

// ClientConfig contains information about the current client.
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

func NewDefaultConfig() ClientConfig {
	// TODO: add logic for determining this for all systems in a different PR.
	const defaultOSVersion = "0.0.0"
	return ClientConfig{
		Language:              SDKLanguage,
		SDKVersion:            SDKSemverVersion,
		RequestLibraryName:    DefaultRequestLibrary,
		RequestLibraryVersion: runtime.Version(),
		SystemOS:              runtime.GOOS,
		SystemArch:            runtime.GOARCH,
		SystemOSVersion:       defaultOSVersion,
	}
}

// The `ClientFactory` creates 1Password Go SDK clients based on the same instance of a WASM module.
type ClientFactory struct {
	core Core
}

func NewClientFactory(ctx context.Context) (*ClientFactory, error) {
	extismCore, err := NewExtismCore(ctx)
	if err != nil {
		return nil, err
	}
	return &ClientFactory{core: extismCore}, nil
}

// NewClient returns a 1Password Go SDK client using the provided ClientOption list.
func (cf ClientFactory) NewClient(opts ...ClientOption) (*OpClient, error) {
	client := OpClient{
		config: NewDefaultConfig(),
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

	clientID, err := cf.core.InitClient(client.config)
	if err != nil {
		return nil, fmt.Errorf("error initializing client: %w", err)
	}

	client.Secrets = NewSecretsSource(*clientID, cf.core)

	runtime.SetFinalizer(&client, func(f *OpClient) {
		cf.core.ReleaseClient(*clientID)
	})
	return &client, nil
}

type ClientOption func(config *OpClient) error

// `WithServiceAccountToken` specifies the [1Password Service Account](https://developer.1password.com/docs/service-accounts) token to use to authenticate the SDK client.
func WithServiceAccountToken(token string) ClientOption {
	return func(c *OpClient) error {
		c.config.SAToken = token
		return nil
	}
}

// `WithIntegrationInfo` specifies the name and version of the integration built using the 1Password Go SDK. If you don't know which name and version to use, use `DefaultIntegrationName` and `DefaultIntegrationVersion`, respectively.
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
