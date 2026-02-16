package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NoConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Rules.Disable) != 0 {
		t.Errorf("expected no disabled rules, got %v", cfg.Rules.Disable)
	}
}

func TestLoad_ExplicitConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "custom.toml")
	os.WriteFile(cfgPath, []byte(`
[rules]
disable = ["bad-defaults.missing-pipefail"]
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Rules.Disable) != 1 || cfg.Rules.Disable[0] != "bad-defaults.missing-pipefail" {
		t.Errorf("expected disabled rule, got %v", cfg.Rules.Disable)
	}
}

func TestLoad_DotTruthsayerToml(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".truthsayer.toml"), []byte(`
[rules]
disable = ["silent-fallback.empty-error-check"]
`), 0644)

	cfg, err := Load(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Rules.Disable) != 1 || cfg.Rules.Disable[0] != "silent-fallback.empty-error-check" {
		t.Errorf("expected disabled rule, got %v", cfg.Rules.Disable)
	}
}

func TestLoad_ExplicitOverridesDefault(t *testing.T) {
	dir := t.TempDir()
	// Create both default and explicit config
	os.WriteFile(filepath.Join(dir, ".truthsayer.toml"), []byte(`
[rules]
disable = ["rule-a"]
`), 0644)

	explicit := filepath.Join(dir, "other.toml")
	os.WriteFile(explicit, []byte(`
[rules]
disable = ["rule-b"]
`), 0644)

	cfg, err := Load(dir, explicit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Explicit config should win
	if len(cfg.Rules.Disable) != 1 || cfg.Rules.Disable[0] != "rule-b" {
		t.Errorf("expected rule-b disabled, got %v", cfg.Rules.Disable)
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.toml")
	os.WriteFile(cfgPath, []byte(`[rules
not valid toml
`), 0644)

	_, err := Load(".", cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestLoad_ExplicitNotFound(t *testing.T) {
	_, err := Load(".", "/nonexistent/config.toml")
	if err == nil {
		t.Fatal("expected error for missing explicit config, got nil")
	}
}

func TestLoad_SeverityOverride(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "sev.toml")
	os.WriteFile(cfgPath, []byte(`
[rules.severity]
"silent-fallback.empty-error-check" = "warning"
"bad-defaults.missing-pipefail" = "info"
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Rules.Severity["silent-fallback.empty-error-check"] != "warning" {
		t.Errorf("expected warning severity, got %v", cfg.Rules.Severity["silent-fallback.empty-error-check"])
	}
	if cfg.Rules.Severity["bad-defaults.missing-pipefail"] != "info" {
		t.Errorf("expected info severity, got %v", cfg.Rules.Severity["bad-defaults.missing-pipefail"])
	}
}

func TestLoad_MultipleDisabledRules(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "multi.toml")
	os.WriteFile(cfgPath, []byte(`
[rules]
disable = ["rule-a", "rule-b", "rule-c"]
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Rules.Disable) != 3 {
		t.Errorf("expected 3 disabled rules, got %d", len(cfg.Rules.Disable))
	}
}

func TestLoad_ScanExcludeDirs(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "excl.toml")
	os.WriteFile(cfgPath, []byte(`
[scan]
exclude_dirs = ["generated", "legacy"]
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Scan.ExcludeDirs) != 2 {
		t.Fatalf("expected 2 exclude_dirs, got %d", len(cfg.Scan.ExcludeDirs))
	}
	if cfg.Scan.ExcludeDirs[0] != "generated" || cfg.Scan.ExcludeDirs[1] != "legacy" {
		t.Errorf("unexpected exclude_dirs: %v", cfg.Scan.ExcludeDirs)
	}
}

func TestLoad_ScanExcludePatterns(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "pat.toml")
	os.WriteFile(cfgPath, []byte(`
[scan]
exclude_patterns = ["*_generated.go", "*.pb.go", "*.min.js"]
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Scan.ExcludePatterns) != 3 {
		t.Fatalf("expected 3 exclude_patterns, got %d", len(cfg.Scan.ExcludePatterns))
	}
	if cfg.Scan.ExcludePatterns[1] != "*.pb.go" {
		t.Errorf("unexpected pattern: %v", cfg.Scan.ExcludePatterns)
	}
}

func TestLoad_NoScanSection_EmptyDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "rules-only.toml")
	os.WriteFile(cfgPath, []byte(`
[rules]
disable = ["rule-a"]
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Scan.ExcludeDirs) != 0 {
		t.Errorf("expected no exclude_dirs, got %v", cfg.Scan.ExcludeDirs)
	}
	if len(cfg.Scan.ExcludePatterns) != 0 {
		t.Errorf("expected no exclude_patterns, got %v", cfg.Scan.ExcludePatterns)
	}
}

func TestLoad_EmptyConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "empty.toml")
	os.WriteFile(cfgPath, []byte(""), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Rules.Disable) != 0 {
		t.Errorf("expected no disabled rules, got %v", cfg.Rules.Disable)
	}
	if len(cfg.Rules.Severity) != 0 {
		t.Errorf("expected no severity overrides, got %v", cfg.Rules.Severity)
	}
}

func TestLoad_LanguageConfig_DisablePython(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lang.toml")
	os.WriteFile(cfgPath, []byte(`
[scan.languages]
python = false
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Scan.Languages.Python == nil || *cfg.Scan.Languages.Python {
		t.Error("expected python disabled")
	}
	if !cfg.Scan.Languages.IsEnabled("go") {
		t.Error("go should be enabled by default")
	}
	if !cfg.Scan.Languages.IsEnabled("javascript") {
		t.Error("javascript should be enabled by default")
	}
	if cfg.Scan.Languages.IsEnabled("python") {
		t.Error("python should be disabled")
	}
}

func TestLoad_LanguageConfig_DisableMultiple(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lang.toml")
	os.WriteFile(cfgPath, []byte(`
[scan.languages]
javascript = false
typescript = false
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Scan.Languages.IsEnabled("javascript") {
		t.Error("javascript should be disabled")
	}
	if cfg.Scan.Languages.IsEnabled("typescript") {
		t.Error("typescript should be disabled")
	}
	if !cfg.Scan.Languages.IsEnabled("go") {
		t.Error("go should still be enabled")
	}
	if !cfg.Scan.Languages.IsEnabled("python") {
		t.Error("python should still be enabled")
	}
}

func TestLoad_LanguageConfig_NoSection_AllEnabled(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nolang.toml")
	os.WriteFile(cfgPath, []byte(`
[rules]
disable = []
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, lang := range []string{"go", "javascript", "typescript", "python", "bash"} {
		if !cfg.Scan.Languages.IsEnabled(lang) {
			t.Errorf("%s should be enabled by default", lang)
		}
	}
}

func TestLanguageConfig_IsEnabled_UnknownLang(t *testing.T) {
	lc := &LanguageConfig{}
	if !lc.IsEnabled("rust") {
		t.Error("unknown languages should default to enabled")
	}
}

func TestLoad_LanguageConfig_ExplicitEnable(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lang.toml")
	os.WriteFile(cfgPath, []byte(`
[scan.languages]
go = true
python = false
`), 0644)

	cfg, err := Load(".", cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Scan.Languages.IsEnabled("go") {
		t.Error("go should be explicitly enabled")
	}
	if cfg.Scan.Languages.IsEnabled("python") {
		t.Error("python should be disabled")
	}
}
