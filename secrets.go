package onepassword

// `SecretsAPI` represents all operations the SDK client can perform on 1Password secrets.
type SecretsAPI interface {
	Resolve(reference string) (*string, error)
}

type SecretsSource struct {
	clientID uint64
	core     Core
}

func NewSecretsSource(id uint64, core Core) *SecretsSource {
	return &SecretsSource{clientID: id, core: core}
}

// Resolve returns the secret the provided reference points to.
func (s SecretsSource) Resolve(reference string) (*string, error) {
	res, err := s.core.Invoke(Invocation{
		ClientID:         s.clientID,
		MethodName:       "Resolve",
		SerializedParams: reference,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
