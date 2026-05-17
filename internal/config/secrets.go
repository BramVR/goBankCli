package config

import "os"

const (
	EnvGoCardlessSecretID  = "GOBANKCLI_GOCARDLESS_SECRET_ID"
	EnvGoCardlessSecretKey = "GOBANKCLI_GOCARDLESS_SECRET_KEY"

	EnvEnableBankingApplicationID = "GOBANKCLI_ENABLEBANKING_APP_ID"
	EnvEnableBankingPrivateKey    = "GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH"
	EnvEnableBankingAPI           = "GOBANKCLI_ENABLEBANKING_API"
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

type EnableBankingCredentials struct {
	ApplicationID  string
	PrivateKeyPath string
	API            string
}

func EnableBankingCredentialsFromEnv() EnableBankingCredentials {
	return EnableBankingCredentials{
		ApplicationID:  os.Getenv(EnvEnableBankingApplicationID),
		PrivateKeyPath: os.Getenv(EnvEnableBankingPrivateKey),
		API:            os.Getenv(EnvEnableBankingAPI),
	}
}

func (c EnableBankingCredentials) Configured() bool {
	return c.ApplicationID != "" && c.PrivateKeyPath != ""
}
