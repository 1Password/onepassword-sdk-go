package onepassword

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// invalid name/version test cases may belong in integration_test.go
// currently these tests fail correctly, but they have incorrect tokens, so the accuracy of these fails is in question

func TestNoToken(t *testing.T) {
	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	// missing token
	_, err = clientFactory.NewClient(
		WithIntegrationInfo(DefaultIntegrationName, DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestNoIntegrationName(t *testing.T) {
	token := ""

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("", DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestInvalidIntegrationNameLength(t *testing.T) {
	token := ""

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("12345678901234567890123456789012345678901234567890", DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestInvalidIntegrationNameCharacters(t *testing.T) {
	token := ""

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo("$", DefaultIntegrationVersion))
	require.Error(t, err)
}

func TestNoIntegrationVersion(t *testing.T) {
	token := ""

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, ""))
	require.Error(t, err)
}

func TestInvalidIntegrationVersionLength(t *testing.T) {
	token := ""

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "12345678901234567890123456789012345678901234567890"))
	require.Error(t, err)
}

func TestInvalidIntegrationVersionCharacters(t *testing.T) {
	token := ""

	clientFactory, err := NewClientFactory(context.TODO())
	require.NoError(t, err)

	_, err = clientFactory.NewClient(
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "$"))
	require.Error(t, err)
}
