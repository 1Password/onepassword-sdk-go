package onepassword

// SecretsAPI represent all operations that can be done on secrets.
type SecretsAPI interface {
	Resolve(reference string) (*string, error)
}

type SecretsSource struct {
	clientID uint64
}

// Resolve returns the secret the provided reference points to.
func (s SecretsSource) Resolve(reference string) (*string, error) {
	res, err := Invoke(Invocation{
		ClientID:         s.clientID,
		MethodName:       "Resolve",
		SerializedParams: reference,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
