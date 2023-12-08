package main

type SecretsAPI interface {
	Resolve(reference string) (*string, error)
}
