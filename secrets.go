package onepassword

import "strconv"

type SecretsAPI interface {
	Resolve(reference string) (*string, error)
}

type SecretsSource struct {
	clientID uint64
}

func (s SecretsSource) Resolve(reference string) (*string, error) {
	println("Invoking with client id: " + strconv.FormatUint(s.clientID, 10))
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
