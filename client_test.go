package onepassword

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func preTest() {
	sharedCore = TestCore{}
}

func TestNoToken(t *testing.T) {
	preTest()
	// missing token
	_, err := NewClient(context.TODO(),
		WithIntegrationInfo(DefaultIntegrationName, DefaultIntegrationVersion))
	assert.Equal(t, "cannot create a client without specifying a Service Account Token", err.Error())
}

func TestNoIntegrationName(t *testing.T) {
	preTest()
	token := "my_token"

	_, err := NewClient(context.TODO(),
		WithServiceAccountToken(token),
		WithIntegrationInfo("", DefaultIntegrationVersion))
	assert.Equal(t, "cannot create a client without defining an app name and version. If you don't want to specify any, use the provided constants: 'DefaultIntegrationName', 'DefaultIntegrationVersion'", err.Error())
}

func TestInvalidIntegrationNameLength(t *testing.T) {
	preTest()
	token := "my_token"

	_, err := NewClient(context.TODO(),
		WithServiceAccountToken(token),
		WithIntegrationInfo("12345678901234567890123456789012345678901234567890", DefaultIntegrationVersion))
	assert.Equal(t, "integration name can't be longer than 40 characters", err.Error())
}

func TestInvalidIntegrationNameCharacters(t *testing.T) {
	preTest()
	token := "my_token"

	_, err := NewClient(context.TODO(),
		WithServiceAccountToken(token),
		WithIntegrationInfo("$", DefaultIntegrationVersion))
	assert.Equal(t, "integration name can only contain digits, letters and allowed symbols", err.Error())
}

func TestNoIntegrationVersion(t *testing.T) {
	preTest()
	token := "my_token"

	_, err := NewClient(context.TODO(),
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, ""))
	assert.Equal(t, "cannot create a client without defining an app name and version. If you don't want to specify any, use the provided constants: 'DefaultIntegrationName', 'DefaultIntegrationVersion'", err.Error())
}

func TestInvalidIntegrationVersionLength(t *testing.T) {
	preTest()
	token := "my_token"

	_, err := NewClient(context.TODO(),
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "12345678901234567890123456789012345678901234567890"))
	assert.Equal(t, "integration version can't be longer than 20 characters", err.Error())
}

func TestInvalidIntegrationVersionCharacters(t *testing.T) {
	preTest()
	token := "my_token"

	_, err := NewClient(context.TODO(),
		WithServiceAccountToken(token),
		WithIntegrationInfo(DefaultIntegrationName, "$"))
	assert.Equal(t, "integration version can only contain digits, letters and allowed symbols", err.Error())
}

type TestCore struct {
}

func (c TestCore) InitClient(config ClientConfig) (*uint64, error) {
	res := uint64(0)
	return &res, nil
}

func (c TestCore) Invoke(invokeConfig Invocation) (*string, error) {
	response := "secret"
	return &response, nil
}

func (c TestCore) ReleaseClient(clientID uint64) {}
