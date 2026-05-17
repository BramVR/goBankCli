package config

import "os"

const (
	EnvGoCardlessSecretID  = "GOBANKCLI_GOCARDLESS_SECRET_ID"
	EnvGoCardlessSecretKey = "GOBANKCLI_GOCARDLESS_SECRET_KEY"
)

type GoCardlessCredentials struct {
	SecretID  string
	SecretKey string
}

func GoCardlessCredentialsFromEnv() GoCardlessCredentials {
	return GoCardlessCredentials{
		SecretID:  os.Getenv(EnvGoCardlessSecretID),
		SecretKey: os.Getenv(EnvGoCardlessSecretKey),
	}
}

func (c GoCardlessCredentials) Configured() bool {
	return c.SecretID != "" && c.SecretKey != ""
}
