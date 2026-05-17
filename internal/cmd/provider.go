package cmd

import (
	"fmt"
	"strings"

	"gobankcli/internal/config"
	"gobankcli/internal/provider"
	"gobankcli/internal/provider/enablebanking"
	"gobankcli/internal/provider/gocardless"
)

func newProvider(name string) (provider.Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", gocardless.Name:
		creds := config.GoCardlessCredentialsFromEnv()
		return gocardless.New(provider.Config{
			Credentials: map[string]string{
				gocardless.CredentialSecretID:  creds.SecretID,
				gocardless.CredentialSecretKey: creds.SecretKey,
			},
		})
	case enablebanking.Name:
		creds := config.EnableBankingCredentialsFromEnv()
		return enablebanking.New(provider.Config{
			BaseURL: creds.API,
			Credentials: map[string]string{
				enablebanking.CredentialApplicationID: creds.ApplicationID,
				enablebanking.CredentialPrivateKey:    creds.PrivateKeyPath,
			},
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}
