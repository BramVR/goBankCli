package cmd

import (
	"context"
	"os"

	"gobankcli/internal/config"
)

type DoctorCmd struct{}

type doctorReport struct {
	ConfigPath                  string `json:"config_path"`
	ConfigExists                bool   `json:"config_exists"`
	DBPath                      string `json:"db_path"`
	DBExists                    bool   `json:"db_exists"`
	DefaultProvider             string `json:"default_provider"`
	GoCardlessSecretID          string `json:"gocardless_secret_id"`
	GoCardlessSecretKey         string `json:"gocardless_secret_key"`
	GoCardlessConfigured        bool   `json:"gocardless_configured"`
	EnableBankingApplicationID  string `json:"enablebanking_application_id"`
	EnableBankingPrivateKeyPath string `json:"enablebanking_private_key_path"`
	EnableBankingAPI            string `json:"enablebanking_api"`
	EnableBankingConfigured     bool   `json:"enablebanking_configured"`
}

func (c DoctorCmd) Run(_ context.Context, app *App) error {
	gocardlessCreds := config.GoCardlessCredentialsFromEnv()
	enableBankingCreds := config.EnableBankingCredentialsFromEnv()
	report := doctorReport{
		ConfigPath:                  app.Config.SourcePath,
		ConfigExists:                fileExists(app.Config.SourcePath),
		DBPath:                      app.Config.Paths.DB,
		DBExists:                    fileExists(app.Config.Paths.DB),
		DefaultProvider:             app.Config.DefaultProvider,
		GoCardlessSecretID:          presence(gocardlessCreds.SecretID),
		GoCardlessSecretKey:         presence(gocardlessCreds.SecretKey),
		GoCardlessConfigured:        gocardlessCreds.Configured(),
		EnableBankingApplicationID:  presence(enableBankingCreds.ApplicationID),
		EnableBankingPrivateKeyPath: presence(enableBankingCreds.PrivateKeyPath),
		EnableBankingAPI:            optionalPresence(enableBankingCreds.API),
		EnableBankingConfigured:     enableBankingCreds.Configured(),
	}
	return app.Out.Write(report)
}

func presence(value string) string {
	if value == "" {
		return "missing"
	}
	return "set"
}

func optionalPresence(value string) string {
	if value == "" {
		return "default"
	}
	return "set"
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
