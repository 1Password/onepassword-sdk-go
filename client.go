package _password_go_sdk

import (
	"context"
	"errors"
	"fmt"
	"runtime"
)

const (
	DefaultAppName    = "Unknown"
	DefaultAppVersion = "Unknown"

	SDKLanguage = "Go"

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
	AppName               string `json:"appName"`
	AppVersion            string `json:"appVersion"`
	RequestLibraryName    string `json:"requestLibraryName"`
	RequestLibraryVersion string `json:"requestLibraryVersion"`
	SystemOS              string `json:"os"`
	SystemArch            string `json:"arch"`
}

func Client(opts ...ClientOption) (*OpClient, error) {
	client := OpClient{
		context: context.Background(),
	}

	for _, opt := range opts {
		opt(&client)
	}

	if len(client.config.SAToken) == 0 {
		return nil, errors.New("cannot create a client without specifying a Service Account Token")
	}

	if len(client.config.AppName) == 0 || len(client.config.AppVersion) == 0 {
		return nil, errors.New("cannot create a client without defining an app name and version. If you don't want to specify any, use the provided constants: 'DefaultAppName', 'DefaultAppVersion'")
	}

	client.config.Language = SDKLanguage
	client.config.RequestLibraryName = DefaultRequestLibrary
	client.config.RequestLibraryVersion = runtime.Version()
	client.config.SystemOS = runtime.GOOS
	client.config.SystemArch = runtime.GOARCH

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

type ClientOption func(config *OpClient)

func WithServiceAccountToken(token string) ClientOption {
	return func(c *OpClient) {
		c.config.SAToken = token
	}
}

func WithApp(name string, version string) ClientOption {
	return func(c *OpClient) {
		c.config.AppName = name
		c.config.AppVersion = version
	}
}

func WithContext(ctx context.Context) ClientOption {
	return func(c *OpClient) {
		c.context = ctx
	}
}
