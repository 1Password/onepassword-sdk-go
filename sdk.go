package sdk

import (
	"fmt"
	"os"
	"runtime"

	core "github.com/1password/1password-sdk-core"
)

// opClient The client instance.
type opClient struct {
	id uint64
}

// NewServiceAccountClient constructor for `opClient`.
func NewServiceAccountClient(saToken string) (*opClient, error) {
	clientId, err := core.InitClient(saToken)
	if err != nil {
		return nil, err
	}
	client := &opClient{id: *clientId}

	runtime.SetFinalizer(client, func(c *opClient) {
		core.ReleaseClient(c.id)
	})

	return client, nil
}

// NewServiceAccountClientFromEnv constructor for `opClient` from the environment.
func NewServiceAccountClientFromEnv() (*opClient, error) {
	const tokenEnvVar = "OP_SERVICE_ACCOUNT_TOKEN"
	token, ok := os.LookupEnv(tokenEnvVar)
	if !ok {
		return nil, fmt.Errorf("no variable %s was found in the enviroment", tokenEnvVar)
	}

	return NewServiceAccountClient(token)
}

// Resolve returns a secret pointed to by the given secret reference
func (c *opClient) Resolve(secretReference string) ([]byte, error) {
	response, err := core.Invoke(c.id, "Resolve", secretReference)
	if err != nil {
		return nil, err
	}
	secret := []byte(*response)
	return secret, nil
}
