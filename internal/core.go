package internal

import (
	"context"
	_ "embed"
	"runtime"
)

const (
	SDKLanguage           = "Go"
	DefaultRequestLibrary = "net/http"
)

//go:embed release/version-build
var SDKSemverVersion string

type Core interface {
	InitClient(ctx context.Context, config ClientConfig) (*uint64, error)
	Invoke(ctx context.Context, invokeConfig InvokeConfig) (*string, error)
	ReleaseClient(clientID uint64)
}

// ClientConfig contains information required for creating a client.
type ClientConfig struct {
	SAToken               string `json:"serviceAccountToken"`
	Language              string `json:"programmingLanguage"`
	SDKVersion            string `json:"sdkVersion"`
	IntegrationName       string `json:"integrationName"`
	IntegrationVersion    string `json:"integrationVersion"`
	RequestLibraryName    string `json:"requestLibraryName"`
	RequestLibraryVersion string `json:"requestLibraryVersion"`
	SystemOS              string `json:"os"`
	SystemOSVersion       string `json:"osVersion"`
	SystemArch            string `json:"architecture"`
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

// InvokeConfig specifies over the FFI on which client the specified method should be invoked on.
type InvokeConfig struct {
	ClientID   uint64     `json:"clientId"`
	Invocation Invocation `json:"invocation"`
}

// Invocation holds the information required for invoking SDK functionality.
type Invocation struct {
	MethodName       string                 `json:"name"`
	SerializedParams map[string]interface{} `json:"parameters"`
}

// InnerClient represents the sdk-core client on which calls will be made.
type InnerClient struct {
	ID   uint64
	Core Core
}
