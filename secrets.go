package onepassword

// SecretsAPI represents all operations the SDK client can perform on 1Password secrets.
type SecretsAPI interface {
	Resolve(reference string) (string, error)
}

// SecretsSource implements SecretsAPI relying on an inner client for operations with secrets.
type SecretsSource struct {
	InnerClient
}

func NewSecretsSource(inner InnerClient) *SecretsSource {
	return &SecretsSource{inner}
}

// Resolve returns the secret the provided reference points to.
func (s SecretsSource) Resolve(secretReference string) (string, error) {
	res, err := clientInvoke(s.InnerClient, "Resolve", map[string]interface{}{
		"secret_reference": secretReference,
	})
	if err != nil {
		return "", err
	}
	return *res, nil
}
