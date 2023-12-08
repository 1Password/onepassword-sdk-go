package onepassword

import (
	"context"
)

type SecretsAPI interface {
	Resolve(reference string) (*string, error)
}

type SecretsSource struct {
	Context context.Context
}

func (s SecretsSource) Resolve(reference string) (*string, error) {
	res, err := Invoke(s.Context, "Resolve", reference)
	if err != nil {
		return nil, err
	}
	return res, nil
}
