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
