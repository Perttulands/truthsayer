package scanner

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/perttulands/truthsayer/internal/rules"
)

func testdataPath(rel string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", rel)
}

func TestGoScanner_EmptyErrorCheck(t *testing.T) {
	s := NewGoScanner([]rules.ASTChecker{&rules.EmptyErrorCheck{}})

	findings, err := s.Scan(testdataPath("go/empty_error_check.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("expected findings for empty error check, got none")
	}
	f := findings[0]
	if f.Rule != "silent-fallback.empty-error-check" {
		t.Errorf("expected rule silent-fallback.empty-error-check, got %s", f.Rule)
	}
	if f.Message == "" {
		t.Error("expected non-empty message")
	}
	if f.Suggestion == "" {
		t.Error("expected non-empty suggestion")
	}
}

func TestGoScanner_ProperHandling(t *testing.T) {
	s := NewGoScanner([]rules.ASTChecker{&rules.EmptyErrorCheck{}})

	findings, err := s.Scan(testdataPath("go/proper_error_handling.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings for proper error handling, got %d", len(findings))
	}
}
