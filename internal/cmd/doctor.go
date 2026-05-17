package cmd

import (
	"context"
	"os"

	"gobankcli/internal/config"
)

type DoctorCmd struct{}

type doctorReport struct {
	ConfigPath           string `json:"config_path"`
	ConfigExists         bool   `json:"config_exists"`
	DBPath               string `json:"db_path"`
	DBExists             bool   `json:"db_exists"`
	DefaultProvider      string `json:"default_provider"`
	GoCardlessSecretID   string `json:"gocardless_secret_id"`
	GoCardlessSecretKey  string `json:"gocardless_secret_key"`
	GoCardlessConfigured bool   `json:"gocardless_configured"`
}

func (c DoctorCmd) Run(_ context.Context, app *App) error {
	creds := config.GoCardlessCredentialsFromEnv()
	report := doctorReport{
		ConfigPath:           app.Config.SourcePath,
		ConfigExists:         fileExists(app.Config.SourcePath),
		DBPath:               app.Config.Paths.DB,
		DBExists:             fileExists(app.Config.Paths.DB),
		DefaultProvider:      app.Config.DefaultProvider,
		GoCardlessSecretID:   presence(creds.SecretID),
		GoCardlessSecretKey:  presence(creds.SecretKey),
		GoCardlessConfigured: creds.Configured(),
	}
	return app.Out.Write(report)
}

func presence(value string) string {
	if value == "" {
		return "missing"
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
