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
	if !strings.Contains(out, "Version") {
		t.Error("output missing Version check")
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

	if !strings.Contains(out, "enabled") {
		t.Errorf("expected 'enabled', got: %s", out)
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

	if !strings.Contains(out, "Config ... ok") {
		t.Errorf("expected 'Config ... ok' message, got: %s", out)
	}
	if !strings.Contains(out, "enabled") {
		t.Errorf("expected 'enabled' in rule count, got: %s", out)
	}
	if !strings.Contains(out, "disabled") {
		t.Errorf("expected 'disabled' count shown, got: %s", out)
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

	if !strings.Contains(out, "FAIL") {
		t.Errorf("expected 'FAIL' message, got: %s", out)
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

	if !strings.Contains(out, "Config ... ok") {
		t.Errorf("expected 'Config ... ok', got: %s", out)
	}
	if !strings.Contains(out, "enabled") {
		t.Errorf("expected 'enabled', got: %s", out)
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

	if !strings.Contains(out, "Version") {
		t.Error("output missing Version check")
	}
	if !strings.Contains(out, "Config") {
		t.Error("output missing Config check")
	}
	if !strings.Contains(out, "Rules") {
		t.Error("output missing Rules check")
	}
	if !strings.Contains(out, "Files") {
		t.Error("output missing Files check")
	}
}
