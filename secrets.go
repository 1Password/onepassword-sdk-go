package onepassword

import "github.com/1password/1password-go-sdk/ffi"

// SecretsAPI represents all operations the SDK client can perform on 1Password secrets.
type SecretsAPI interface {
	Resolve(reference string) (*string, error)
}

// SecretsSource implements SecretsAPI relying on an inner client for operations with secrets.
type SecretsSource struct {
	InnerClient
}

func NewSecretsSource(inner InnerClient) *SecretsSource {
	return &SecretsSource{inner}
}

// Resolve returns the secret the provided reference points to.
func (s SecretsSource) Resolve(reference string) (*string, error) {
	res, err := s.core.Invoke(ffi.InvokeConfig{
		ClientID: s.id,
		Invocation: ffi.Invocation{
			MethodName:       "Resolve",
			SerializedParams: reference,
		},
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
