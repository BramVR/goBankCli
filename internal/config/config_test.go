package config

import (
	"path/filepath"
	"testing"
)

func TestExpandPathHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	got := ExpandPath("~/Finance/gobankcli")
	want := filepath.Join(home, "Finance", "gobankcli")
	if got != want {
		t.Fatalf("ExpandPath() = %q, want %q", got, want)
	}
}

func TestGoCardlessCredentialsConfigured(t *testing.T) {
	t.Setenv(EnvGoCardlessSecretID, "id")
	t.Setenv(EnvGoCardlessSecretKey, "key")
	creds := GoCardlessCredentialsFromEnv()
	if !creds.Configured() {
		t.Fatal("credentials should be configured")
	}
}
