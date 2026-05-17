package config

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	SourcePath      string       `toml:"-"`
	DefaultProvider string       `toml:"default_provider"`
	DefaultCountry  string       `toml:"default_country"`
	Paths           Paths        `toml:"paths"`
	Connections     []Connection `toml:"connections,omitempty"`
}

type Paths struct {
	DB      string `toml:"db"`
	Exports string `toml:"exports"`
}

type Connection struct {
	Name          string `toml:"name"`
	Provider      string `toml:"provider"`
	InstitutionID string `toml:"institution_id"`
	Country       string `toml:"country"`
}

func Default() Config {
	return Config{
		SourcePath:      DefaultConfigPath(),
		DefaultProvider: "gocardless",
		DefaultCountry:  "BE",
		Paths: Paths{
			DB:      DefaultDBPath(),
			Exports: DefaultExportsPath(),
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path != "" {
		cfg.SourcePath = ExpandPath(path)
	}
	sourcePath := cfg.SourcePath
	b, err := os.ReadFile(cfg.SourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg.Expand()
			return cfg, nil
		}
		return Config{}, err
	}
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	cfg.SourcePath = sourcePath
	if cfg.DefaultProvider == "" {
		cfg.DefaultProvider = "gocardless"
	}
	if cfg.DefaultCountry == "" {
		cfg.DefaultCountry = "BE"
	}
	if cfg.Paths.DB == "" {
		cfg.Paths.DB = DefaultDBPath()
	}
	if cfg.Paths.Exports == "" {
		cfg.Paths.Exports = DefaultExportsPath()
	}
	cfg.SourcePath = ExpandPath(cfg.SourcePath)
	cfg.Expand()
	return cfg, nil
}

func (c *Config) Expand() {
	c.SourcePath = ExpandPath(c.SourcePath)
	c.Paths.DB = ExpandPath(c.Paths.DB)
	c.Paths.Exports = ExpandPath(c.Paths.Exports)
}

func (c Config) Public() Config {
	c.SourcePath = ""
	return c
}
