package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// Defaults represents repository-wide defaults stored in .mm/defaults.toml.
type Defaults struct {
	Backend  BackendDefaults  `toml:"backend"`
	Frontend FrontendDefaults `toml:"frontend"`
	Database DatabaseDefaults `toml:"database"`
}

// BackendDefaults configures backend language defaults.
type BackendDefaults struct {
	Lang string `toml:"lang"`
}

// FrontendDefaults configures frontend defaults.
type FrontendDefaults struct {
	Lang           string `toml:"lang"`
	Server         string `toml:"server"`
	Client         string `toml:"client"`
	PackageManager string `toml:"packageManager"`
}

// DatabaseDefaults configures database defaults.
type DatabaseDefaults struct {
	Engine string `toml:"engine"`
}

// ServiceConfig represents per-service configuration stored in service.toml.
type ServiceConfig struct {
	General      GeneralConfig                `toml:"general"`
	Dependencies DependenciesConfig           `toml:"dependencies"`
	Environment  map[string]map[string]string `toml:"environment"`
}

// GeneralConfig holds basic service metadata.
type GeneralConfig struct {
	Lang     string `toml:"lang"`
	Database string `toml:"database,omitempty"`
	External bool   `toml:"external,omitempty"`
}

// DependenciesConfig lists service dependencies by name.
type DependenciesConfig struct {
	Services []string `toml:"services"`
}

// HasDatabase reports whether the service declares a database.
func (s ServiceConfig) HasDatabase() bool {
	return s.General.Database != ""
}

// HasDependencies reports whether the service declares dependent services.
func (s ServiceConfig) HasDependencies() bool {
	return len(s.Dependencies.Services) > 0
}

// DefaultDefaults returns opinionated defaults for new repositories.
func DefaultDefaults() Defaults {
	return Defaults{
		Backend: BackendDefaults{Lang: "go"},
		Frontend: FrontendDefaults{
			Lang:           "ts",
			Server:         "next.js",
			Client:         "react",
			PackageManager: "pnpm",
		},
		Database: DatabaseDefaults{Engine: "postgres"},
	}
}

// LoadDefaults reads .mm/defaults.toml.
func LoadDefaults(root string) (Defaults, error) {
	path := filepath.Join(root, ".mm", "defaults.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Defaults{}, fmt.Errorf("defaults not found, run mm init: %w", err)
		}
		return Defaults{}, err
	}
	var cfg Defaults
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Defaults{}, err
	}
	return cfg, nil
}

// SaveDefaults writes .mm/defaults.toml.
func SaveDefaults(root string, cfg Defaults) error {
	path := filepath.Join(root, ".mm", "defaults.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadServiceConfig reads services/<name>/service.toml.
func LoadServiceConfig(root, serviceName string) (ServiceConfig, error) {
	path := filepath.Join(root, "services", serviceName, "service.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return ServiceConfig{}, err
	}
	var cfg ServiceConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return ServiceConfig{}, err
	}
	return cfg, nil
}

// SaveServiceConfig writes services/<name>/service.toml.
func SaveServiceConfig(root, serviceName string, cfg ServiceConfig) error {
	path := filepath.Join(root, "services", serviceName, "service.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
