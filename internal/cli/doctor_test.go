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

func TestDoctor_ShowsPerLanguageRuleCounts(t *testing.T) {
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

	// Should show per-language breakdown
	if !strings.Contains(out, "Go") {
		t.Errorf("expected per-language rule count for Go, got: %s", out)
	}
	if !strings.Contains(out, "JS/TS") {
		t.Errorf("expected per-language rule count for JS/TS, got: %s", out)
	}
	if !strings.Contains(out, "Python") {
		t.Errorf("expected per-language rule count for Python, got: %s", out)
	}
}

func TestDoctor_ShowsParserStatus(t *testing.T) {
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

	if !strings.Contains(out, "JS/TS AST parser") {
		t.Errorf("expected JS/TS parser status, got: %s", out)
	}
	if !strings.Contains(out, "Python AST parser") {
		t.Errorf("expected Python parser status, got: %s", out)
	}
	if !strings.Contains(out, "tree-sitter") {
		t.Errorf("expected tree-sitter mentioned in parser status, got: %s", out)
	}
}

func TestDoctor_CountsMultiLanguageFiles(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create files for each language
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log()"), 0644)
	os.WriteFile(filepath.Join(dir, "app.ts"), []byte("const x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "script.py"), []byte("print()"), 0644)
	os.WriteFile(filepath.Join(dir, "run.sh"), []byte("#!/bin/bash"), 0644)

	out := captureStdout(t, func() {
		code := runDoctor(nil)
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	// Should count JS/TS and Python files in addition to Go and bash
	if !strings.Contains(out, "1 Go") {
		t.Errorf("expected '1 Go' in file count, got: %s", out)
	}
	if !strings.Contains(out, "2 JS/TS") {
		t.Errorf("expected '2 JS/TS' in file count, got: %s", out)
	}
	if !strings.Contains(out, "1 Python") {
		t.Errorf("expected '1 Python' in file count, got: %s", out)
	}
	if !strings.Contains(out, "1 bash") {
		t.Errorf("expected '1 bash' in file count, got: %s", out)
	}
}

func TestDoctor_ParserStatusAvailable(t *testing.T) {
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

	// Both parsers should report "available" since tree-sitter is compiled in
	if !strings.Contains(out, "available (tree-sitter)") {
		t.Errorf("expected 'available (tree-sitter)' in parser status, got: %s", out)
	}
}
