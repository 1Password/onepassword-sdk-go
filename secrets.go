package onepassword

type SecretsAPI interface {
	Resolve(reference string) (*string, error)
}

type SecretsSource struct {
	clientID uint64
}

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
