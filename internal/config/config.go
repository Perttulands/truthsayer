package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the .truthsayer.toml configuration.
type Config struct {
	Scan  ScanConfig  `toml:"scan"`
	Rules RulesConfig `toml:"rules"`
}

// ScanConfig holds file/directory exclusion settings.
type ScanConfig struct {
	ExcludeDirs     []string       `toml:"exclude_dirs"`
	ExcludePatterns []string       `toml:"exclude_patterns"`
	Languages       LanguageConfig `toml:"languages"`
}

// LanguageConfig controls which languages are scanned.
// nil (unset) means enabled (all languages on by default).
type LanguageConfig struct {
	Go         *bool `toml:"go"`
	JavaScript *bool `toml:"javascript"`
	TypeScript *bool `toml:"typescript"`
	Python     *bool `toml:"python"`
	Bash       *bool `toml:"bash"`
	Rust       *bool `toml:"rust"`
}

// IsEnabled returns whether a language is enabled. Unset (nil) defaults to true.
func (lc *LanguageConfig) IsEnabled(lang string) bool {
	var p *bool
	switch lang {
	case "go":
		p = lc.Go
	case "javascript":
		p = lc.JavaScript
	case "typescript":
		p = lc.TypeScript
	case "python":
		p = lc.Python
	case "bash":
		p = lc.Bash
	case "rust":
		p = lc.Rust
	default:
		return true
	}
	if p == nil {
		return true
	}
	return *p
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
