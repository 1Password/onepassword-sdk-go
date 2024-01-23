package internal

import (
	"context"
	"runtime"
)

const (
	SDKSemverVersion      = "0010001" // v0.1.0
	SDKLanguage           = "Go"
	DefaultRequestLibrary = "net/http"
)

type Core interface {
	InitClient(config ClientConfig) (*uint64, error)
	Invoke(invokeConfig InvokeConfig) (*string, error)
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
	MethodName       string `json:"name"`
	SerializedParams string `json:"parameters"`
}

// SharedCore implements Core in such a way that all created client instances share the same core resources.
type SharedCore struct {
	sharedContext context.Context
}

// GetSharedCore initializes the shared core once and re-uses it on subsequent calls.
func GetSharedCore(ctx context.Context) *SharedCore {
	return &SharedCore{ctx}
}

func (ec SharedCore) InitClient(config ClientConfig) (*uint64, error) {
	return nil, nil
}
func (ec SharedCore) Invoke(invokeConfig InvokeConfig) (*string, error) {
	return nil, nil
}
func (ec SharedCore) ReleaseClient(clientID uint64) {}
