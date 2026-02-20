package engine

import (
	"os"
	"path/filepath"
	"strings"
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
		if f.Context == "" {
			t.Error("finding missing Context")
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
		if f.Context == "" {
			t.Error("finding missing Context")
		}
	}
}

func TestFindingsHaveContext_AllLanguageBackends(t *testing.T) {
	type testCase struct {
		name     string
		fileName string
		source   string
	}

	cases := []testCase{
		{
			name:     "go-ast",
			fileName: "bad.go",
			source: `package main

import "fmt"

func bad() error {
	err := fmt.Errorf("x")
	if err != nil {
		return nil
	}
	return nil
}
`,
		},
		{
			name:     "js-ast",
			fileName: "bad.js",
			source: `function run() {
  try {
    work()
  } catch (err) {}
}
`,
		},
		{
			name:     "py-ast",
			fileName: "bad.py",
			source: `def run():
    try:
        work()
    except:
        pass
`,
		},
		{
			name:     "bash-regex",
			fileName: "bad.sh",
			source: `#!/bin/bash
cmd || true
`,
		},
		{
			name:     "config-regex",
			fileName: ".env",
			source:   "API_KEY=secret123456\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, tc.fileName)
			if err := os.WriteFile(path, []byte(tc.source), 0o644); err != nil {
				t.Fatalf("write source file: %v", err)
			}

			eng := New(rules.DefaultRegistry())
			result, err := eng.ScanFile(path)
			if err != nil {
				t.Fatalf("scan file: %v", err)
			}
			if len(result.Findings) == 0 {
				t.Fatalf("expected findings for %s", tc.fileName)
			}
			for _, f := range result.Findings {
				if strings.TrimSpace(f.Context) == "" {
					t.Fatalf("expected non-empty context for rule %s", f.Rule)
				}
				if !strings.Contains(f.Context, f.Code) {
					t.Fatalf("expected context to include code snippet %q, got:\n%s", f.Code, f.Context)
				}
			}
		})
	}
}
