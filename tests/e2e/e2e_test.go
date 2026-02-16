// Package e2e provides end-to-end tests for the truthsayer CLI.
//
// These tests build the truthsayer binary, create temporary directories
// with known anti-pattern files, run the CLI, and verify output, exit codes,
// and JSON structure.
package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// binaryPath holds the compiled truthsayer binary path, built once per test run.
var (
	binaryPath string
	buildOnce  sync.Once
	buildErr   error
)

// ensureBinary builds the truthsayer binary once and returns the path.
func ensureBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		// Use os.TempDir() directly so the binary outlives individual test cleanup.
		bin := filepath.Join(os.TempDir(), "truthsayer-e2e-test")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		// Find project root (two levels up from tests/e2e/)
		_, thisFile, _, _ := runtime.Caller(0)
		projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

		cmd := exec.Command("go", "build", "-o", bin, "./cmd/truthsayer/")
		cmd.Dir = projectRoot
		cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = err
			t.Logf("build output: %s", string(out))
		}
		binaryPath = bin
	})
	if buildErr != nil {
		t.Fatalf("failed to build truthsayer: %v", buildErr)
	}
	return binaryPath
}

// runTruthsayer executes the truthsayer binary with args, returning stdout, stderr, and exit code.
func runTruthsayer(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	bin := ensureBinary(t)
	cmd := exec.Command(bin, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("failed to run truthsayer: %v", err)
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// writeFile creates a file with the given content inside dir.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// jsonReport is a minimal struct matching the JSON report output.
type jsonReport struct {
	Version    string        `json:"version"`
	ScanTime   string        `json:"scan_time"`
	Path       string        `json:"path"`
	DurationMs int64         `json:"duration_ms"`
	Findings   []jsonFinding `json:"findings"`
	Summary    jsonSummary   `json:"summary"`
}

type jsonFinding struct {
	Rule       string `json:"rule"`
	Severity   string `json:"severity"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

type jsonSummary struct {
	Total        int   `json:"total"`
	Errors       int   `json:"errors"`
	Warnings     int   `json:"warnings"`
	Info         int   `json:"info"`
	FilesScanned int   `json:"files_scanned"`
	DurationMs   int64 `json:"duration_ms"`
}

// parseReport parses JSON output into a report struct.
func parseReport(t *testing.T, data string) jsonReport {
	t.Helper()
	var rpt jsonReport
	if err := json.Unmarshal([]byte(data), &rpt); err != nil {
		t.Fatalf("failed to parse JSON report: %v\nraw: %s", err, data)
	}
	return rpt
}

// findingsByRule filters findings to those matching the given rule ID.
func findingsByRule(findings []jsonFinding, ruleID string) []jsonFinding {
	var out []jsonFinding
	for _, f := range findings {
		if f.Rule == ruleID {
			out = append(out, f)
		}
	}
	return out
}

// hasRule returns true if any finding matches the given rule ID.
func hasRule(findings []jsonFinding, ruleID string) bool {
	return len(findingsByRule(findings, ruleID)) > 0
}

// ─── Test: Scan Go anti-patterns ─────────────────────────────────────────────

func TestScanGoAntiPatterns(t *testing.T) {
	dir := t.TempDir()

	// File with ignored error (blank identifier for error return)
	writeFile(t, dir, "bad.go", `package main

import "os"

func bad() {
	_, _ = os.Open("/tmp/nope")
}
`)

	stdout, _, exitCode := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	if !hasRule(rpt.Findings, "silent-fallback.ignored-error") {
		t.Error("expected finding for silent-fallback.ignored-error")
	}

	if rpt.Summary.FilesScanned < 1 {
		t.Error("expected at least 1 file scanned")
	}
}

// ─── Test: Scan Go — swallowed error ─────────────────────────────────────────

func TestScanGoSwallowedError(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "swallowed.go", `package main

import (
	"fmt"
	"log"
)

func process() {
	err := fmt.Errorf("connection lost")
	if err != nil {
		log.Println("error occurred:", err)
	}
	fmt.Println("continuing anyway")
}
`)

	stdout, _, exitCode := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if !hasRule(rpt.Findings, "error-context.swallowed-error") {
		t.Error("expected finding for error-context.swallowed-error")
	}

	// swallowed-error is warning, not error — so exit code should be 0 unless other rules fire
	_ = exitCode
}

// ─── Test: Scan bash anti-patterns ───────────────────────────────────────────

func TestScanBashAntiPatterns(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "bad.sh", `#!/bin/bash
some_command || true
echo "continuing"
`)

	stdout, _, exitCode := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 (error-severity bash findings), got %d", exitCode)
	}

	if !hasRule(rpt.Findings, "silent-fallback.hidden-failure-bash") {
		t.Error("expected finding for silent-fallback.hidden-failure-bash")
	}

	// Also test missing pipefail
	if !hasRule(rpt.Findings, "bad-defaults.missing-pipefail") {
		t.Error("expected finding for bad-defaults.missing-pipefail")
	}
}

// ─── Test: Scan JS/TS anti-patterns ──────────────────────────────────────────

func TestScanJSAntiPatterns(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "service.js", `
async function fetchData() {
  try {
    const resp = await fetch("/api/data");
    return resp.json();
  } catch (e) {
  }
}
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if !hasRule(rpt.Findings, "silent-fallback.js-empty-catch") {
		t.Error("expected finding for silent-fallback.js-empty-catch")
	}
}

// ─── Test: Scan Python anti-patterns ─────────────────────────────────────────

func TestScanPythonAntiPatterns(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "handler.py", `
import json

def process_data(raw):
    try:
        data = json.loads(raw)
    except:
        pass
    return data
`)

	stdout, _, exitCode := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 (bare except is error severity), got %d", exitCode)
	}

	if !hasRule(rpt.Findings, "silent-fallback.py-bare-except") {
		t.Error("expected finding for silent-fallback.py-bare-except")
	}

	if !hasRule(rpt.Findings, "silent-fallback.py-except-pass") {
		t.Error("expected finding for silent-fallback.py-except-pass")
	}
}

// ─── Test: Clean code — exit code 0 ──────────────────────────────────────────

func TestCleanCodeExitZero(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "clean.go", `package main

import "fmt"

func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
`)

	_, _, exitCode := runTruthsayer(t, "scan", dir, "--format", "json")

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for clean code, got %d", exitCode)
	}
}

// ─── Test: JSON output structure ─────────────────────────────────────────────

func TestJSONOutputStructure(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "bad.go", `package main

import "os"

func leak() {
	_, _ = os.Open("/tmp/test")
}
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	// Verify top-level fields
	if rpt.Version == "" {
		t.Error("expected non-empty version")
	}
	if rpt.ScanTime == "" {
		t.Error("expected non-empty scan_time")
	}
	if rpt.Path == "" {
		t.Error("expected non-empty path")
	}

	// Verify findings have all required fields
	for i, f := range rpt.Findings {
		if f.Rule == "" {
			t.Errorf("finding[%d]: empty rule", i)
		}
		if f.Severity == "" {
			t.Errorf("finding[%d]: empty severity", i)
		}
		if f.Severity != "error" && f.Severity != "warning" && f.Severity != "info" {
			t.Errorf("finding[%d]: invalid severity %q", i, f.Severity)
		}
		if f.File == "" {
			t.Errorf("finding[%d]: empty file", i)
		}
		if f.Line <= 0 {
			t.Errorf("finding[%d]: invalid line %d", i, f.Line)
		}
		if f.Message == "" {
			t.Errorf("finding[%d]: empty message", i)
		}
	}

	// Verify summary
	if rpt.Summary.FilesScanned < 1 {
		t.Error("expected files_scanned >= 1")
	}
	if rpt.Summary.Total != len(rpt.Findings) {
		t.Errorf("summary.total (%d) != len(findings) (%d)", rpt.Summary.Total, len(rpt.Findings))
	}
	if rpt.Summary.Errors+rpt.Summary.Warnings+rpt.Summary.Info != rpt.Summary.Total {
		t.Error("summary counts don't add up")
	}
}

// ─── Test: Check single file ─────────────────────────────────────────────────

func TestCheckSingleFile(t *testing.T) {
	dir := t.TempDir()

	path := writeFile(t, dir, "single.sh", `#!/bin/bash
rm -rf /tmp/test || true
`)

	_, _, exitCode := runTruthsayer(t, "check", path)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 for file with error-severity finding, got %d", exitCode)
	}
}

// ─── Test: Check single file — clean ─────────────────────────────────────────

func TestCheckSingleFileClean(t *testing.T) {
	dir := t.TempDir()

	path := writeFile(t, dir, "clean.go", `package main

func add(a, b int) int {
	return a + b
}
`)

	_, _, exitCode := runTruthsayer(t, "check", path)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for clean file, got %d", exitCode)
	}
}

// ─── Test: Config — disable rule ─────────────────────────────────────────────

func TestConfigDisableRule(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "bad.sh", `#!/bin/bash
cmd || true
`)

	// Without config, the hidden-failure rule should fire
	stdout, _, exitCode := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)
	if !hasRule(rpt.Findings, "silent-fallback.hidden-failure-bash") {
		t.Fatal("precondition: hidden-failure-bash should fire without config")
	}
	if exitCode != 1 {
		t.Fatal("precondition: exit code should be 1")
	}

	// Now write config that disables both the hidden-failure and missing-pipefail rules
	writeFile(t, dir, ".truthsayer.toml", `[rules]
disable = ["silent-fallback.hidden-failure-bash", "bad-defaults.missing-pipefail", "silent-fallback.no-err-trap"]
`)

	stdout2, _, exitCode2 := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt2 := parseReport(t, stdout2)

	if hasRule(rpt2.Findings, "silent-fallback.hidden-failure-bash") {
		t.Error("disabled rule silent-fallback.hidden-failure-bash should not appear")
	}
	if hasRule(rpt2.Findings, "bad-defaults.missing-pipefail") {
		t.Error("disabled rule bad-defaults.missing-pipefail should not appear")
	}

	if exitCode2 != 0 {
		t.Errorf("expected exit code 0 after disabling error rules, got %d", exitCode2)
	}
}

// ─── Test: Config — severity override ────────────────────────────────────────

func TestConfigSeverityOverride(t *testing.T) {
	dir := t.TempDir()

	// swallowed-error is normally "warning"
	writeFile(t, dir, "swallowed.go", `package main

import (
	"fmt"
	"log"
)

func run() {
	err := fmt.Errorf("disk full")
	if err != nil {
		log.Println("oh no:", err)
	}
	fmt.Println("still running")
}
`)

	// First verify it's a warning by default
	stdout1, _, exitCode1 := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt1 := parseReport(t, stdout1)

	swallowed := findingsByRule(rpt1.Findings, "error-context.swallowed-error")
	if len(swallowed) == 0 {
		t.Fatal("precondition: should have swallowed-error finding")
	}
	if swallowed[0].Severity != "warning" {
		t.Fatalf("precondition: swallowed-error should be warning, got %s", swallowed[0].Severity)
	}
	// Warning-only should not trigger exit 1 (unless other error-level rules fire)
	_ = exitCode1

	// Now promote to error
	writeFile(t, dir, ".truthsayer.toml", `[rules.severity]
"error-context.swallowed-error" = "error"
`)

	stdout2, _, exitCode2 := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt2 := parseReport(t, stdout2)

	swallowed2 := findingsByRule(rpt2.Findings, "error-context.swallowed-error")
	if len(swallowed2) == 0 {
		t.Fatal("should still find swallowed-error after severity override")
	}
	if swallowed2[0].Severity != "error" {
		t.Errorf("expected severity 'error' after override, got %s", swallowed2[0].Severity)
	}
	if exitCode2 != 1 {
		t.Errorf("expected exit code 1 after promoting to error, got %d", exitCode2)
	}
}

// ─── Test: Config — exclude patterns ─────────────────────────────────────────

func TestConfigExcludePatterns(t *testing.T) {
	dir := t.TempDir()

	// Two files: one matches exclude pattern, one doesn't
	writeFile(t, dir, "bad.go", `package main

import "os"

func a() {
	_, _ = os.Open("/tmp/x")
}
`)
	writeFile(t, dir, "bad_generated.go", `package main

import "os"

func b() {
	_, _ = os.Open("/tmp/y")
}
`)

	// Exclude *_generated.go
	writeFile(t, dir, ".truthsayer.toml", `[scan]
exclude_patterns = ["*_generated.go"]
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	for _, f := range rpt.Findings {
		if strings.Contains(f.File, "generated") {
			t.Error("finding from excluded file *_generated.go should not appear")
		}
	}

	// bad.go should still produce findings
	if !hasRule(rpt.Findings, "silent-fallback.ignored-error") {
		t.Error("non-excluded file should still produce findings")
	}
}

// ─── Test: Config — exclude dirs ─────────────────────────────────────────────

func TestConfigExcludeDirs(t *testing.T) {
	dir := t.TempDir()

	// File in excluded subdir
	writeFile(t, dir, "generated/bad.go", `package main

import "os"

func x() {
	_, _ = os.Open("/tmp/z")
}
`)
	// File in non-excluded location
	writeFile(t, dir, "src/ok.go", `package main

func y() int { return 1 }
`)

	writeFile(t, dir, ".truthsayer.toml", `[scan]
exclude_dirs = ["generated"]
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	for _, f := range rpt.Findings {
		if strings.Contains(f.File, "generated") {
			t.Error("finding from excluded dir 'generated' should not appear")
		}
	}
}

// ─── Test: Config via --config flag ──────────────────────────────────────────

func TestConfigExplicitPath(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "bad.sh", `#!/bin/bash
cmd || true
`)

	// Write config in a different location
	configPath := writeFile(t, dir, "custom/my-config.toml", `[rules]
disable = ["silent-fallback.hidden-failure-bash", "bad-defaults.missing-pipefail", "silent-fallback.no-err-trap"]
`)

	stdout, _, exitCode := runTruthsayer(t, "scan", dir, "--config", configPath, "--format", "json")
	rpt := parseReport(t, stdout)

	if hasRule(rpt.Findings, "silent-fallback.hidden-failure-bash") {
		t.Error("disabled rule should not appear when using --config")
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0 after disabling error rules via --config, got %d", exitCode)
	}
	_ = rpt
}

// ─── Test: Doctor command ────────────────────────────────────────────────────

func TestDoctorCommand(t *testing.T) {
	stdout, _, exitCode := runTruthsayer(t, "doctor")

	if exitCode != 0 {
		t.Errorf("expected doctor exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "truthsayer doctor") {
		t.Error("expected 'truthsayer doctor' header in output")
	}
	if !strings.Contains(stdout, "Version") {
		t.Error("expected Version check in doctor output")
	}
	if !strings.Contains(stdout, "Rules") {
		t.Error("expected Rules check in doctor output")
	}
}

// ─── Test: Rules command ─────────────────────────────────────────────────────

func TestRulesCommand(t *testing.T) {
	stdout, _, exitCode := runTruthsayer(t, "rules")

	if exitCode != 0 {
		t.Errorf("expected rules exit code 0, got %d", exitCode)
	}

	// Should list known rule IDs
	if !strings.Contains(stdout, "silent-fallback.ignored-error") {
		t.Error("expected silent-fallback.ignored-error in rules output")
	}
	if !strings.Contains(stdout, "silent-fallback.js-empty-catch") {
		t.Error("expected JS rules in rules output")
	}
	if !strings.Contains(stdout, "silent-fallback.py-bare-except") {
		t.Error("expected Python rules in rules output")
	}

	// Should show column headers
	if !strings.Contains(stdout, "ID") || !strings.Contains(stdout, "SEVERITY") {
		t.Error("expected column headers in rules output")
	}
}

// ─── Test: Rules --enabled with config ───────────────────────────────────────

func TestRulesEnabledWithConfig(t *testing.T) {
	dir := t.TempDir()

	configPath := writeFile(t, dir, "cfg.toml", `[rules]
disable = ["silent-fallback.ignored-error"]
`)

	// Run rules --enabled from the temp dir, so config loads from current dir won't interfere
	bin := ensureBinary(t)
	cmd := exec.Command(bin, "rules", "--enabled", "--config", configPath)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("rules --enabled failed with exit code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		t.Fatal(err)
	}

	stdout := string(out)
	if strings.Contains(stdout, "silent-fallback.ignored-error") {
		t.Error("disabled rule should not appear in --enabled output")
	}
}

// ─── Test: Version command ───────────────────────────────────────────────────

func TestVersionCommand(t *testing.T) {
	stdout, _, exitCode := runTruthsayer(t, "--version")

	if exitCode != 0 {
		t.Errorf("expected version exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "truthsayer version") {
		t.Error("expected version output")
	}
}

// ─── Test: Unknown command ───────────────────────────────────────────────────

func TestUnknownCommand(t *testing.T) {
	_, _, exitCode := runTruthsayer(t, "bogus")

	if exitCode != 2 {
		t.Errorf("expected exit code 2 for unknown command, got %d", exitCode)
	}
}

// ─── Test: Scan nonexistent path ─────────────────────────────────────────────

func TestScanNonexistentPath(t *testing.T) {
	_, _, exitCode := runTruthsayer(t, "scan", "/nonexistent/path")

	if exitCode != 2 {
		t.Errorf("expected exit code 2 for nonexistent path, got %d", exitCode)
	}
}

// ─── Test: Scan a file (should suggest check) ────────────────────────────────

func TestScanFileInsteadOfDir(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "file.go", `package main`)

	_, stderr, exitCode := runTruthsayer(t, "scan", path)

	if exitCode != 2 {
		t.Errorf("expected exit code 2 when scanning a file with 'scan', got %d", exitCode)
	}
	if !strings.Contains(stderr, "not a directory") {
		t.Error("expected 'not a directory' error message")
	}
}

// ─── Test: Multi-language scan ───────────────────────────────────────────────

func TestMultiLanguageScan(t *testing.T) {
	dir := t.TempDir()

	// Go anti-pattern
	writeFile(t, dir, "app.go", `package main

import "os"

func openFile() {
	_, _ = os.Open("/tmp/data")
}
`)

	// Bash anti-pattern
	writeFile(t, dir, "deploy.sh", `#!/bin/bash
rm -rf /tmp/deploy || true
`)

	// JS anti-pattern
	writeFile(t, dir, "api.js", `
async function load() {
  try {
    const r = await fetch("/api");
    return r.json();
  } catch (err) {
  }
}
`)

	// Python anti-pattern
	writeFile(t, dir, "worker.py", `
def process(data):
    try:
        result = int(data)
    except:
        pass
    return result
`)

	stdout, _, exitCode := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 with multi-language errors, got %d", exitCode)
	}

	// Go
	if !hasRule(rpt.Findings, "silent-fallback.ignored-error") {
		t.Error("expected Go finding: ignored-error")
	}

	// Bash
	if !hasRule(rpt.Findings, "silent-fallback.hidden-failure-bash") {
		t.Error("expected Bash finding: hidden-failure-bash")
	}

	// JS
	if !hasRule(rpt.Findings, "silent-fallback.js-empty-catch") {
		t.Error("expected JS finding: js-empty-catch")
	}

	// Python
	if !hasRule(rpt.Findings, "silent-fallback.py-bare-except") {
		t.Error("expected Python finding: py-bare-except")
	}

	// Verify at least 4 files were scanned
	if rpt.Summary.FilesScanned < 4 {
		t.Errorf("expected >= 4 files scanned, got %d", rpt.Summary.FilesScanned)
	}
}

// ─── Test: Terminal output format (default) ──────────────────────────────────

func TestTerminalOutputFormat(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "bad.go", `package main

import "os"

func f() {
	_, _ = os.Open("/tmp/x")
}
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir)

	// Terminal output should have severity labels and summary
	if !strings.Contains(stdout, "ERROR") && !strings.Contains(stdout, "WARN") {
		t.Error("expected severity labels in terminal output")
	}
	if !strings.Contains(stdout, "Summary:") {
		t.Error("expected Summary line in terminal output")
	}
}

// ─── Test: Empty directory scan ──────────────────────────────────────────────

func TestEmptyDirectoryScan(t *testing.T) {
	dir := t.TempDir()

	stdout, _, exitCode := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for empty dir, got %d", exitCode)
	}
	if len(rpt.Findings) != 0 {
		t.Errorf("expected 0 findings for empty dir, got %d", len(rpt.Findings))
	}
	if rpt.Summary.FilesScanned != 0 {
		t.Errorf("expected 0 files scanned for empty dir, got %d", rpt.Summary.FilesScanned)
	}
}

// ─── Test: Finding fields populated ──────────────────────────────────────────

func TestFindingFieldsPopulated(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "sample.sh", `#!/bin/bash
dangerous_cmd 2>/dev/null
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if len(rpt.Findings) == 0 {
		t.Fatal("expected at least one finding")
	}

	for _, f := range rpt.Findings {
		if f.Code == "" {
			t.Errorf("finding %s: code should not be empty", f.Rule)
		}
		if f.Line <= 0 {
			t.Errorf("finding %s: line should be > 0", f.Rule)
		}
		if f.File == "" {
			t.Errorf("finding %s: file should not be empty", f.Rule)
		}
	}
}

// ─── Test: Python-specific rules ─────────────────────────────────────────────

func TestPythonExceptBroad(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "app.py", `
def load(path):
    try:
        with open(path) as f:
            return f.read()
    except Exception as e:
        return ""
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if !hasRule(rpt.Findings, "silent-fallback.py-except-broad") {
		t.Error("expected finding for py-except-broad")
	}
}

// ─── Test: Python mutable default ────────────────────────────────────────────

func TestPythonMutableDefault(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "utils.py", `
def append_item(item, items=[]):
    items.append(item)
    return items
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if !hasRule(rpt.Findings, "bad-defaults.py-mutable-default-arg") {
		t.Error("expected finding for py-mutable-default-arg")
	}
}

// ─── Test: JS non-null assertion ─────────────────────────────────────────────

func TestJSNonNullAssertion(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "comp.ts", `
function getLength(s: string | null): number {
  return s!.length;
}
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if !hasRule(rpt.Findings, "bad-defaults.non-null-assertion") {
		t.Error("expected finding for non-null-assertion")
	}
}

// ─── Test: Scan with --config pointing to bad TOML ───────────────────────────

func TestBadConfigFile(t *testing.T) {
	dir := t.TempDir()

	badConfig := writeFile(t, dir, "bad.toml", `this is not valid toml [[[[`)

	_, _, exitCode := runTruthsayer(t, "scan", dir, "--config", badConfig)

	if exitCode != 2 {
		t.Errorf("expected exit code 2 for invalid config, got %d", exitCode)
	}
}

// ─── Test: Findings sorted by severity ───────────────────────────────────────

func TestFindingsSortedBySeverity(t *testing.T) {
	dir := t.TempDir()

	// Create file that should trigger both error and warning level findings
	writeFile(t, dir, "mixed.go", `package main

import (
	"fmt"
	"log"
	"os"
)

func mixed() {
	_, _ = os.Open("/tmp/a")
	err := fmt.Errorf("problem")
	if err != nil {
		log.Println("error:", err)
	}
}
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	if len(rpt.Findings) < 2 {
		t.Skip("need at least 2 findings for sort test")
	}

	// Verify errors come before warnings/info
	sevRank := map[string]int{"error": 0, "warning": 1, "info": 2}
	for i := 1; i < len(rpt.Findings); i++ {
		prevRank := sevRank[rpt.Findings[i-1].Severity]
		currRank := sevRank[rpt.Findings[i].Severity]
		if prevRank > currRank {
			t.Errorf("findings not sorted: %s (%s) before %s (%s)",
				rpt.Findings[i-1].Rule, rpt.Findings[i-1].Severity,
				rpt.Findings[i].Rule, rpt.Findings[i].Severity)
			break
		}
	}
}

// ─── Test: No args prints usage ──────────────────────────────────────────────

func TestNoArgsPrintsUsage(t *testing.T) {
	_, stderr, exitCode := runTruthsayer(t)

	if exitCode != 2 {
		t.Errorf("expected exit code 2 for no args, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Error("expected usage text on stderr")
	}
}

// ─── Test: Config invalid severity value ─────────────────────────────────────

func TestConfigInvalidSeverity(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "ok.go", `package main

func f() {}
`)
	writeFile(t, dir, ".truthsayer.toml", `[rules.severity]
"silent-fallback.ignored-error" = "critical"
`)

	_, _, exitCode := runTruthsayer(t, "scan", dir)

	if exitCode != 2 {
		t.Errorf("expected exit code 2 for invalid severity, got %d", exitCode)
	}
}

// ─── Test: Multiple bash patterns ────────────────────────────────────────────

func TestBashMultiplePatterns(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "script.sh", `#!/bin/bash
set -e
cmd1 || true
cmd2 2>/dev/null
cmd3 || :
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	bashFindings := findingsByRule(rpt.Findings, "silent-fallback.hidden-failure-bash")
	if len(bashFindings) < 3 {
		t.Errorf("expected at least 3 hidden-failure-bash findings (|| true, 2>/dev/null, || :), got %d", len(bashFindings))
	}
}

// ─── Test: REASON comment suppresses bash severity ───────────────────────────

func TestBashReasonComment(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "justified.sh", `#!/bin/bash
set -euo pipefail
trap 'echo err' ERR
# REASON: optional cleanup, failure is acceptable
rm -rf /tmp/cache || true
`)

	stdout, _, _ := runTruthsayer(t, "scan", dir, "--format", "json")
	rpt := parseReport(t, stdout)

	bashFindings := findingsByRule(rpt.Findings, "silent-fallback.hidden-failure-bash")
	for _, f := range bashFindings {
		if f.Severity == "error" {
			t.Error("hidden-failure with REASON comment should not be error severity")
		}
	}
}
