package onepassword

import (
	"context"
	"github.com/1password/onepassword-sdk-go/internal"
)

// SecretsAPI represents all operations the SDK client can perform on 1Password secrets.
type SecretsAPI interface {
	Resolve(ctx context.Context, reference string) (string, error)
}

// SecretsSource implements SecretsAPI relying on an inner client for operations with secrets.
type SecretsSource struct {
	InnerClient
}

func NewSecretsSource(inner InnerClient) *SecretsSource {
	return &SecretsSource{inner}
}

// Resolve returns the secret the provided reference points to.
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
