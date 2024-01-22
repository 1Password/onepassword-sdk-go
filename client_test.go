package onepassword

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoToken(t *testing.T) {
	clientFactory := NewTestClientFactory()

	// missing token
	_, err := clientFactory.NewClient(
		WithIntegrationInfo(DefaultIntegrationName, DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestNoIntegrationNameOrVersion(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("", ""))
	require.Error(t, err)
}

func TestNoIntegrationName(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("", DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestInvalidIntegrationNameLength(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("12345678901234567890123456789012345678901234567890", DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestInvalidIntegrationNameCharacters(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("$", DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestNoIntegrationVersion(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, ""))
	require.Error(t, err)
}

func TestInvalidIntegrationVersionLength(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "12345678901234567890123456789012345678901234567890"))
	require.Error(t, err)
}

func TestInvalidIntegrationVersionCharacters(t *testing.T) {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")

	clientFactory := NewTestClientFactory()

	_, err := clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "$"))
	require.Error(t, err)
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
