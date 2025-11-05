package user

import "slices"

type Provider string

const (
	ProviderAnonymous Provider = "anonymous"
	ProviderGoogle    Provider = "google"
	ProviderFacebook  Provider = "facebook"
)

var validProviders = []Provider{
	ProviderAnonymous,
	ProviderGoogle,
	ProviderFacebook,
}

// isValid checks if the given provider is valid.
func (p Provider) isValid() error {
	if !slices.Contains(validProviders, p) {
		return ErrInvalidProvider(p)
	}
	return nil
}
