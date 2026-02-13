package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
)

func TestEngine_ScanDirectory(t *testing.T) {
	// Create a temp directory with test files
	tmp := t.TempDir()

	// Go file with anti-pattern
	goCode := `package main

import "fmt"

func handler() error {
	err := fmt.Errorf("fail")
	if err != nil {
		return nil
	}
	return nil
}
`
	os.WriteFile(filepath.Join(tmp, "bad.go"), []byte(goCode), 0o644)

	// Bash file with anti-pattern
	bashCode := "#!/bin/bash\necho hello\n"
	os.WriteFile(filepath.Join(tmp, "bad.sh"), []byte(bashCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if result.FilesScanned != 2 {
		t.Errorf("expected 2 files scanned, got %d", result.FilesScanned)
	}

	if len(result.Findings) < 2 {
		t.Fatalf("expected at least 2 findings (go + bash), got %d", len(result.Findings))
	}

	// Verify sorted: errors first
	for i := 1; i < len(result.Findings); i++ {
		if result.Findings[i-1].Severity == finding.SeverityInfo &&
			result.Findings[i].Severity == finding.SeverityError {
			t.Error("findings not sorted by severity")
		}
	}
}

func TestEngine_ScanEmptyDirectory(t *testing.T) {
	tmp := t.TempDir()
	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if result.FilesScanned != 0 {
		t.Errorf("expected 0 files scanned, got %d", result.FilesScanned)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(result.Findings))
	}
}
