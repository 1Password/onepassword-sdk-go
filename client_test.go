package onepassword

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoToken(t *testing.T) {
	clientFactory := NewTestClientFactory()

	// missing token
	_, err := clientFactory.NewClient(
		WithIntegrationInfo(DefaultIntegrationName, DefaultIntegrationVersion))
	assert.Equal(t, "cannot create a client without specifying a Service Account Token", err.Error())
}

func TestNoIntegrationName(t *testing.T) {
	token := "my_token"

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("", DefaultIntegrationVersion))
	assert.Equal(t, "cannot create a client without defining an app name and version. If you don't want to specify any, use the provided constants: 'DefaultIntegrationName', 'DefaultIntegrationVersion'", err.Error())
}

func TestInvalidIntegrationNameLength(t *testing.T) {
	token := "my_token"

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("12345678901234567890123456789012345678901234567890", DefaultIntegrationVersion))
	assert.Equal(t, "integration name can't be longer than 40 characters", err.Error())
}

func TestInvalidIntegrationNameCharacters(t *testing.T) {
	token := "my_token"

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("$", DefaultIntegrationVersion))
	assert.Equal(t, "integration name can only contain digits, letters and allowed symbols", err.Error())
}

func TestNoIntegrationVersion(t *testing.T) {
	token := "my_token"

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, ""))
	assert.Equal(t, "cannot create a client without defining an app name and version. If you don't want to specify any, use the provided constants: 'DefaultIntegrationName', 'DefaultIntegrationVersion'", err.Error())
}

func TestInvalidIntegrationVersionLength(t *testing.T) {
	token := "my_token"

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "12345678901234567890123456789012345678901234567890"))
	assert.Equal(t, "integration version can't be longer than 20 characters", err.Error())
}

func TestInvalidIntegrationVersionCharacters(t *testing.T) {
	token := "my_token"

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "$"))
	assert.Equal(t, "integration version can only contain digits, letters and allowed symbols", err.Error())
}

func NewTestClientFactory() *ClientFactory {
	return &ClientFactory{core: TestCore{}}
}

type TestCore struct{}

func (c TestCore) InitClient(config ClientConfig) (*uint64, error) {
	var id uint64
	return &id, nil
}

func (c TestCore) Invoke(invokeConfig Invocation) (*string, error) {
	response := "secret"
	return &response, nil
}

func (c TestCore) ReleaseClient(clientID uint64) {}
