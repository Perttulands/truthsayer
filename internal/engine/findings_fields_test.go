package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/perttulands/truthsayer/internal/rules"
)

// TestFindingsHaveAllFields verifies US-003: every finding includes
// file path, line number, code snippet, and suggestion.
func TestFindingsHaveAllFields_Go(t *testing.T) {
	tmp := t.TempDir()
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
	path := filepath.Join(tmp, "bad.go")
	os.WriteFile(path, []byte(goCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected findings, got none")
	}
	for _, f := range result.Findings {
		if f.File == "" {
			t.Error("finding missing File")
		}
		if f.Line == 0 {
			t.Error("finding missing Line (is 0)")
		}
		if f.Code == "" {
			t.Error("finding missing Code snippet")
		}
		if f.Message == "" {
			t.Error("finding missing Message")
		}
		if f.Suggestion == "" {
			t.Error("finding missing Suggestion")
		}
		// Code should be the actual source line, not a generic placeholder
		if f.Code == "if err != nil { return nil }" {
			// This is the hardcoded placeholder — it should now be the actual source
			// The actual source has "if err != nil {" on one line
			t.Error("Code should be the actual source line, not a hardcoded placeholder")
		}
	}
}

func TestFindingsHaveAllFields_Bash(t *testing.T) {
	tmp := t.TempDir()
	bashCode := "#!/bin/bash\necho hello\n"
	path := filepath.Join(tmp, "bad.sh")
	os.WriteFile(path, []byte(bashCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected findings, got none")
	}
	for _, f := range result.Findings {
		if f.File == "" {
			t.Error("finding missing File")
		}
		if f.Line == 0 {
			t.Error("finding missing Line (is 0)")
		}
		if f.Code == "" {
			t.Error("finding missing Code snippet")
		}
		if f.Message == "" {
			t.Error("finding missing Message")
		}
		if f.Suggestion == "" {
			t.Error("finding missing Suggestion")
		}
	}
}
