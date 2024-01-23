package onepassword

import (
	"context"
	"fmt"
	"runtime"
)

const (
	SDKSemverVersion = "0010001" // v0.1.0

	DefaultIntegrationName    = "Unknown"
	DefaultIntegrationVersion = "Unknown"

	SDKLanguage           = "Go"
	DefaultRequestLibrary = "net/http"
)

// Client represents an instance of the 1Password Go SDK client.
type Client struct {
	config  ClientConfig
	Secrets SecretsAPI
}

// NewClient returns a 1Password Go SDK client using the provided ClientOption list.
func NewClient(sharedContext context.Context, opts ...ClientOption) (*Client, error) {
	core := GetSharedCore(sharedContext)
	return createClient(core, opts...)
}

func createClient(core Core, opts ...ClientOption) (*Client, error) {
	client := Client{
		config: NewDefaultConfig(),
	}

	for _, opt := range opts {
		err := opt(&client)
		if err != nil {
			return nil, err
		}
	}

	clientID, err := core.InitClient(client.config)
	if err != nil {
		return nil, fmt.Errorf("error initializing client: %w", err)
	}

	inner := InnerClient{
		id:   *clientID,
		core: core,
	}

	client.Secrets = NewSecretsSource(inner)

	runtime.SetFinalizer(&client, func(f *Client) {
		core.ReleaseClient(*clientID)
	})
	return &client, nil
}

// InnerClient represents the sdk-core client on which calls will be made.
type InnerClient struct {
	id   uint64
	core Core
}

type ClientOption func(client *Client) error

// WithServiceAccountToken specifies the [1Password Service Account](https://developer.1password.com/docs/service-accounts) token to use to authenticate the SDK client.
func WithServiceAccountToken(token string) ClientOption {
	return func(c *Client) error {
		c.config.SAToken = token
		return nil
	}
}

// WithIntegrationInfo specifies the name and version of the integration built using the 1Password Go SDK. If you don't know which name and version to use, use `DefaultIntegrationName` and `DefaultIntegrationVersion`, respectively.
func WithIntegrationInfo(name string, version string) ClientOption {
	return func(c *Client) error {
		c.config.IntegrationName = name
		c.config.IntegrationVersion = version
		return nil
	}
}
