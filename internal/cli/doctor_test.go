package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_ShowsVersion(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	out := captureStdout(t, func() {
		code := runDoctor(nil)
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "truthsayer") {
		t.Error("output missing 'truthsayer' name")
	}
}

func TestDoctor_ShowsRuleCount(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	out := captureStdout(t, func() {
		runDoctor(nil)
	})

	if !strings.Contains(out, "2 rules enabled") {
		t.Errorf("expected '2 rules enabled', got: %s", out)
	}
}

func TestDoctor_NoConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	out := captureStdout(t, func() {
		code := runDoctor(nil)
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "no config file") {
		t.Errorf("expected 'no config file' message, got: %s", out)
	}
}

func TestDoctor_ValidConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile(filepath.Join(dir, ".truthsayer.toml"), []byte(`
[rules]
disable = ["silent-fallback.empty-error-check"]
`), 0644)

	out := captureStdout(t, func() {
		code := runDoctor(nil)
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "config valid") {
		t.Errorf("expected 'config valid' message, got: %s", out)
	}
	if !strings.Contains(out, "1 rules enabled") {
		t.Errorf("expected '1 rules enabled' with one disabled, got: %s", out)
	}
}

func TestDoctor_InvalidConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile(filepath.Join(dir, ".truthsayer.toml"), []byte(`not valid toml {{{{`), 0644)

	out := captureStdout(t, func() {
		code := runDoctor(nil)
		if code != 1 {
			t.Errorf("expected exit code 1 for invalid config, got %d", code)
		}
	})

	if !strings.Contains(out, "config invalid") {
		t.Errorf("expected 'config invalid' message, got: %s", out)
	}
}

func TestDoctor_ExplicitConfigFlag(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	cfgPath := filepath.Join(t.TempDir(), "custom.toml")
	os.WriteFile(cfgPath, []byte(`
[rules]
disable = ["bad-defaults.missing-pipefail"]
`), 0644)

	out := captureStdout(t, func() {
		code := runDoctor([]string{"--config", cfgPath})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "config valid") {
		t.Errorf("expected 'config valid', got: %s", out)
	}
	if !strings.Contains(out, "1 rules enabled") {
		t.Errorf("expected '1 rules enabled', got: %s", out)
	}
}

func TestDoctor_ShowsChecks(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	out := captureStdout(t, func() {
		runDoctor(nil)
	})

	// Should have check labels
	if !strings.Contains(out, "Installation") {
		t.Error("output missing Installation check")
	}
	if !strings.Contains(out, "Config") {
		t.Error("output missing Config check")
	}
	if !strings.Contains(out, "Rules") {
		t.Error("output missing Rules check")
	}
}
