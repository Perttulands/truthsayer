package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the .truthsayer.toml configuration.
type Config struct {
	Rules RulesConfig `toml:"rules"`
}

// RulesConfig holds rule enable/disable and severity settings.
type RulesConfig struct {
	Disable  []string          `toml:"disable"`
	Severity map[string]string `toml:"severity"`
}

// Load reads configuration from a TOML file.
// If explicit is non-empty, it is used directly (error if not found).
// Otherwise, .truthsayer.toml in scanDir is tried.
// If no config file exists, returns default (all rules enabled).
func Load(scanDir, explicit string) (*Config, error) {
	cfg := &Config{
		Rules: RulesConfig{
			Severity: make(map[string]string),
		},
	}

	path := explicit
	if path == "" {
		candidate := filepath.Join(scanDir, ".truthsayer.toml")
		if _, err := os.Stat(candidate); err == nil {
			path = candidate
		}
	}

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	if cfg.Rules.Severity == nil {
		cfg.Rules.Severity = make(map[string]string)
	}

	return cfg, nil
}
