package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/debt"
)

func TestRunDebt_TextOutput(t *testing.T) {
	dir := t.TempDir()
	store := debt.NewStore(filepath.Join(dir, debt.DefaultPath))
	if err := store.Add(debt.Entry{
		RuleID:    "trace-gaps.long-function-no-log",
		File:      "svc.go",
		Line:      12,
		Code:      "func x() {}",
		Message:   "missing logs",
		Reasoning: "needs follow-up",
		CreatedAt: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed debt: %v", err)
	}

	out := captureStdout(t, func() {
		code := runDebt([]string{dir})
		if code != 0 {
			t.Fatalf("expected exit code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "trace-gaps.long-function-no-log") {
		t.Fatalf("expected rule id in output, got:\n%s", out)
	}
}

func TestRunDebt_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, debt.DefaultPath)
	if err := os.WriteFile(path, []byte("[]\n"), 0o644); err != nil {
		t.Fatalf("write debt file: %v", err)
	}

	out := captureStdout(t, func() {
		code := runDebt([]string{"--format", "json", path})
		if code != 0 {
			t.Fatalf("expected exit code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "[") {
		t.Fatalf("expected json output, got:\n%s", out)
	}
}

