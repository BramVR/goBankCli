package config

import (
	"os"
	"path/filepath"
	"strings"
)

func DefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "gobankcli", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "gobankcli", "config.toml")
}

func DefaultDBPath() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "gobankcli", "gobankcli.db")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "gobankcli", "gobankcli.db")
}

func DefaultExportsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Finance", "gobankcli", "exports")
}

func ExpandPath(path string) string {
	if path == "" {
		return ""
	}
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	return os.ExpandEnv(path)
}
