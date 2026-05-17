package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type InitCmd struct {
	Force bool `help:"Overwrite an existing config file."`
}

type initReport struct {
	ConfigPath string `json:"config_path"`
	DBPath     string `json:"db_path"`
	ExportsDir string `json:"exports_dir"`
	Written    bool   `json:"written"`
}

func (c InitCmd) Run(_ context.Context, app *App) error {
	cfg := app.Config
	if err := os.MkdirAll(filepath.Dir(cfg.SourcePath), 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Paths.DB), 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.Paths.Exports, 0o700); err != nil {
		return err
	}

	written := false
	if _, err := os.Stat(cfg.SourcePath); err != nil || c.Force {
		b, err := toml.Marshal(cfg.Public())
		if err != nil {
			return err
		}
		if err := os.WriteFile(cfg.SourcePath, b, 0o600); err != nil {
			return err
		}
		written = true
	} else {
		fmt.Fprintln(app.Stderr, "config exists; use --force to overwrite")
	}

	return app.Out.Write(initReport{
		ConfigPath: cfg.SourcePath,
		DBPath:     cfg.Paths.DB,
		ExportsDir: cfg.Paths.Exports,
		Written:    written,
	})
}
