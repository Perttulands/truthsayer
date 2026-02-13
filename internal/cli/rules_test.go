package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestRules_ListsAllRules(t *testing.T) {
	out := captureStdout(t, func() {
		code := runRules(nil)
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	// Must contain both built-in rules
	if !strings.Contains(out, "silent-fallback.empty-error-check") {
		t.Error("output missing silent-fallback.empty-error-check rule")
	}
	if !strings.Contains(out, "bad-defaults.missing-pipefail") {
		t.Error("output missing bad-defaults.missing-pipefail rule")
	}
}

func TestRules_ShowsSeverity(t *testing.T) {
	out := captureStdout(t, func() {
		runRules(nil)
	})

	if !strings.Contains(out, "ERROR") {
		t.Error("output missing severity ERROR")
	}
}

func TestRules_ShowsDescription(t *testing.T) {
	out := captureStdout(t, func() {
		runRules(nil)
	})

	if !strings.Contains(out, "Error checked but returned as nil") {
		t.Error("output missing description for empty-error-check")
	}
	if !strings.Contains(out, "Bash script without set -euo pipefail") {
		t.Error("output missing description for missing-pipefail")
	}
}

func TestRules_ShowsFileTypes(t *testing.T) {
	out := captureStdout(t, func() {
		runRules(nil)
	})

	if !strings.Contains(out, ".go") {
		t.Error("output missing .go file type")
	}
	if !strings.Contains(out, ".sh") {
		t.Error("output missing .sh file type")
	}
}

func TestRules_ShowsCount(t *testing.T) {
	out := captureStdout(t, func() {
		runRules(nil)
	})

	if !strings.Contains(out, "2 rules available") {
		t.Errorf("output missing rule count, got: %s", out)
	}
}

func TestRules_HasTableHeader(t *testing.T) {
	out := captureStdout(t, func() {
		runRules(nil)
	})

	if !strings.Contains(out, "ID") || !strings.Contains(out, "SEVERITY") || !strings.Contains(out, "DESCRIPTION") {
		t.Error("output missing table header columns")
	}
}

func TestRules_EnabledNoConfig(t *testing.T) {
	// No config file → all rules enabled
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	out := captureStdout(t, func() {
		code := runRules([]string{"--enabled"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "silent-fallback.empty-error-check") {
		t.Error("output missing silent-fallback.empty-error-check")
	}
	if !strings.Contains(out, "bad-defaults.missing-pipefail") {
		t.Error("output missing bad-defaults.missing-pipefail")
	}
	if !strings.Contains(out, "2 rules enabled") {
		t.Errorf("expected '2 rules enabled', got: %s", out)
	}
}

func TestRules_EnabledWithDisabledRule(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile(filepath.Join(dir, ".truthsayer.toml"), []byte(`
[rules]
disable = ["silent-fallback.empty-error-check"]
`), 0644)

	out := captureStdout(t, func() {
		code := runRules([]string{"--enabled"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if strings.Contains(out, "silent-fallback.empty-error-check") {
		t.Error("disabled rule should not appear in --enabled output")
	}
	if !strings.Contains(out, "bad-defaults.missing-pipefail") {
		t.Error("enabled rule should appear in --enabled output")
	}
	if !strings.Contains(out, "1 rules enabled") {
		t.Errorf("expected '1 rules enabled', got: %s", out)
	}
}

func TestRules_EnabledWithConfigFlag(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "custom.toml")
	os.WriteFile(cfgPath, []byte(`
[rules]
disable = ["bad-defaults.missing-pipefail"]
`), 0644)

	out := captureStdout(t, func() {
		code := runRules([]string{"--config", cfgPath, "--enabled"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if strings.Contains(out, "bad-defaults.missing-pipefail") {
		t.Error("disabled rule should not appear in --enabled output")
	}
	if !strings.Contains(out, "silent-fallback.empty-error-check") {
		t.Error("enabled rule should appear in --enabled output")
	}
}

func TestRules_EnabledShowsOverriddenSeverity(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile(filepath.Join(dir, ".truthsayer.toml"), []byte(`
[rules.severity]
"silent-fallback.empty-error-check" = "warning"
`), 0644)

	out := captureStdout(t, func() {
		code := runRules([]string{"--enabled"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	// The overridden rule should show WARNING, not ERROR
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "silent-fallback.empty-error-check") {
			if !strings.Contains(line, "WARNING") {
				t.Errorf("expected WARNING for overridden rule, got line: %s", line)
			}
			if strings.Contains(line, "ERROR") {
				t.Errorf("should not show ERROR for overridden rule, got line: %s", line)
			}
		}
	}
}
