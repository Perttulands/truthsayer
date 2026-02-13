package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile creates a file with given content in a temp dir.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// Go file that triggers error-severity finding (empty error check).
const goWithError = `package foo

import "fmt"

func bad() error {
	err := fmt.Errorf("fail")
	if err != nil {
		return nil
	}
	return nil
}
`

// Go file with no anti-patterns.
const goClean = `package foo

func good() int {
	return 42
}
`

func TestScan_ExitCode1_WhenErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.go", goWithError)

	code := runScan([]string{dir})
	if code != 1 {
		t.Errorf("expected exit code 1 (errors found), got %d", code)
	}
}

func TestScan_ExitCode0_WhenClean(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "good.go", goClean)

	code := runScan([]string{dir})
	if code != 0 {
		t.Errorf("expected exit code 0 (no errors), got %d", code)
	}
}

func TestScan_ExitCode2_BadPath(t *testing.T) {
	code := runScan([]string{"/nonexistent/path"})
	if code != 2 {
		t.Errorf("expected exit code 2 (tool error), got %d", code)
	}
}

func TestScan_ExitCode2_NoArgs(t *testing.T) {
	code := runScan(nil)
	if code != 2 {
		t.Errorf("expected exit code 2 (missing path), got %d", code)
	}
}

func TestCheck_ExitCode1_WhenErrors(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.go", goWithError)

	code := runCheck([]string{path})
	if code != 1 {
		t.Errorf("expected exit code 1 (errors found), got %d", code)
	}
}

func TestCheck_ExitCode0_WhenClean(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "good.go", goClean)

	code := runCheck([]string{path})
	if code != 0 {
		t.Errorf("expected exit code 0 (no errors), got %d", code)
	}
}

func TestCheck_ExitCode2_BadPath(t *testing.T) {
	code := runCheck([]string{"/nonexistent/file.go"})
	if code != 2 {
		t.Errorf("expected exit code 2 (tool error), got %d", code)
	}
}

func TestReport_ExitCode1_WhenErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.go", goWithError)
	output := filepath.Join(t.TempDir(), "report.json")

	code := runReport([]string{"--output", output, dir})
	if code != 1 {
		t.Errorf("expected exit code 1 (errors found), got %d", code)
	}

	// Verify report file was still written
	info, err := os.Stat(output)
	if err != nil {
		t.Fatalf("report file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("report file is empty")
	}
}

func TestReport_ExitCode0_WhenClean(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "good.go", goClean)
	output := filepath.Join(t.TempDir(), "report.json")

	code := runReport([]string{"--output", output, dir})
	if code != 0 {
		t.Errorf("expected exit code 0 (no errors), got %d", code)
	}
}

func TestReport_ExitCode2_BadPath(t *testing.T) {
	code := runReport([]string{"/nonexistent/path"})
	if code != 2 {
		t.Errorf("expected exit code 2 (tool error), got %d", code)
	}
}
