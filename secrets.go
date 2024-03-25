package onepassword

import (
	"context"
	"github.com/1password/onepassword-sdk-go/internal"
)

// SecretsAPI represents all operations the SDK client can perform on 1Password secrets.
type SecretsAPI interface {
	// Resolve returns the secret the provided secret reference points to.
	// Secret references point to fields in 1Password. They have the following format: op://<vault-name>/<item-name>[/<section-name>]/<field-name>
	// Read more about secret references: https://developer.1password.com/docs/cli/secret-references
	Resolve(ctx context.Context, reference string) (string, error)
}

// SecretsSource implements SecretsAPI relying on an inner client for operations with secrets.
type SecretsSource struct {
	InnerClient
}

func NewSecretsSource(inner InnerClient) *SecretsSource {
	return &SecretsSource{inner}
}

// Resolve returns the secret the provided secret reference points to.
// Secret reference syntax: op://<vault-name>/<item-name>[/<section-name>]/<field-name>
// Read more about secret references: https://developer.1password.com/docs/cli/secret-references
func (s SecretsSource) Resolve(ctx context.Context, reference string) (string, error) {
	res, err := s.core.Invoke(ctx, internal.InvokeConfig{
		ClientID: s.id,
		Invocation: internal.Invocation{
			MethodName:       "Resolve",
			SerializedParams: reference,
		},
	})
	if err != nil {
		return "", err
	}
	return *res, nil
}
