package beads

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateProblemBead_UsesExpectedTitleAndPriority(t *testing.T) {
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args.txt")
	mockBR := writeMockBR(t, dir, argsFile, `echo "br-123"`)

	creator := NewBeadCreatorWithCommand(mockBR)

	id, err := creator.CreateProblemBead("silent-fallback.empty-error-check", "pkg/handler.go", 3)
	if err != nil {
		t.Fatalf("CreateProblemBead returned error: %v", err)
	}
	if id != "br-123" {
		t.Fatalf("expected bead ID br-123, got %q", id)
	}

	args := readArgs(t, argsFile)
	expected := []string{
		"create",
		"--title",
		"[truthsayer] silent-fallback.empty-error-check: 3 errors in pkg/handler.go",
		"--priority",
		"1",
	}
	if strings.Join(args, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("unexpected args:\n%s", strings.Join(args, "\n"))
	}
}

func TestCreateWarningBead_UsesWarningPriority(t *testing.T) {
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args.txt")
	mockBR := writeMockBR(t, dir, argsFile, `echo "br-456"`)

	creator := NewBeadCreatorWithCommand(mockBR)

	id, err := creator.CreateWarningBead("trace-gaps.no-request-id", "api/server.go", 2)
	if err != nil {
		t.Fatalf("CreateWarningBead returned error: %v", err)
	}
	if id != "br-456" {
		t.Fatalf("expected bead ID br-456, got %q", id)
	}

	args := readArgs(t, argsFile)
	if !containsPair(args, "--priority", "2") {
		t.Fatalf("expected warning priority 2, args: %v", args)
	}
}

func TestCreateProblemBead_Timeout(t *testing.T) {
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args.txt")
	mockBR := writeMockBR(t, dir, argsFile, `sleep 2; echo "br-late"`)

	creator := NewBeadCreatorWithCommand(mockBR)
	creator.SetTimeout(100 * time.Millisecond)

	_, err := creator.CreateProblemBead("rule.id", "file.go", 1)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func writeMockBR(t *testing.T, dir string, argsFile string, body string) string {
	t.Helper()
	path := filepath.Join(dir, "br")
	script := fmt.Sprintf(`#!/bin/sh
printf "%%s\n" "$@" > %q
%s
`, argsFile, body)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write mock br: %v", err)
	}
	return path
}

func readArgs(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read args: %v", err)
	}
	raw := strings.TrimSpace(string(data))
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}

func containsPair(args []string, key string, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
}
