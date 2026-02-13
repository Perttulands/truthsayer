package cli

import (
	"bytes"
	"os"
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
