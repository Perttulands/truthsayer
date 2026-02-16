package scanner

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
)

// noopJSChecker is a JSASTChecker that records calls for testing.
type noopJSChecker struct {
	mu    sync.Mutex
	calls []string
}

func (c *noopJSChecker) Meta() rules.Rule {
	return rules.Rule{
		ID:       "test.noop-js",
		Category: "test",
		Name:     "Noop JS checker",
		FileTypes: []string{".js", ".ts", ".tsx", ".jsx", ".mjs", ".cjs"},
		ScanType: rules.ScanTypeAST,
	}
}

func (c *noopJSChecker) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, path)
	return nil
}

// findingJSChecker returns a finding for every file it scans.
type findingJSChecker struct{}

func (c *findingJSChecker) Meta() rules.Rule {
	return rules.Rule{
		ID:       "test.finding-js",
		Category: "test",
		Name:     "Finding JS checker",
		FileTypes: []string{".js", ".ts", ".tsx", ".jsx", ".mjs", ".cjs"},
		ScanType: rules.ScanTypeAST,
	}
}

func (c *findingJSChecker) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	root := tree.RootNode()
	if root == nil {
		return nil
	}
	return []finding.Finding{{
		Rule:    "test.finding-js",
		File:    path,
		Line:    1,
		Message: "test finding",
	}}
}

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestJSScanner_ScanJSFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "app.js", `function hello() { return "world"; }`)

	checker := &noopJSChecker{}
	sc := NewJSScanner([]rules.JSASTChecker{checker})

	findings, lines, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings from noop, got %d", len(findings))
	}
	if len(lines) == 0 {
		t.Error("expected source lines to be returned")
	}
	if len(checker.calls) != 1 || checker.calls[0] != path {
		t.Errorf("expected checker to be called once with %s, got %v", path, checker.calls)
	}
}

func TestJSScanner_ScanTSFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "app.ts", `const greet = (name: string): string => name;`)

	checker := &noopJSChecker{}
	sc := NewJSScanner([]rules.JSASTChecker{checker})

	_, _, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(checker.calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(checker.calls))
	}
}

func TestJSScanner_ScanTSXFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "App.tsx", `const App = () => <div>Hello</div>;`)

	checker := &noopJSChecker{}
	sc := NewJSScanner([]rules.JSASTChecker{checker})

	_, _, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(checker.calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(checker.calls))
	}
}

func TestJSScanner_ReturnsFindings(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "app.js", `function hello() {}`)

	checker := &findingJSChecker{}
	sc := NewJSScanner([]rules.JSASTChecker{checker})

	findings, _, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "test.finding-js" {
		t.Errorf("expected rule test.finding-js, got %s", findings[0].Rule)
	}
}

func TestJSScanner_LanguageDetection(t *testing.T) {
	tests := []struct {
		filename string
		content  string
	}{
		{"app.js", `const x = 1;`},
		{"app.mjs", `export const x = 1;`},
		{"app.cjs", `const x = require('fs');`},
		{"app.jsx", `const App = () => <div/>;`},
		{"app.ts", `const x: number = 1;`},
		{"app.tsx", `const App = (): JSX.Element => <div/>;`},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			dir := t.TempDir()
			path := writeTestFile(t, dir, tt.filename, tt.content)

			checker := &noopJSChecker{}
			sc := NewJSScanner([]rules.JSASTChecker{checker})

			_, _, err := sc.Scan(path)
			if err != nil {
				t.Fatalf("Scan(%s) failed: %v", tt.filename, err)
			}
			if len(checker.calls) != 1 {
				t.Errorf("expected checker called for %s", tt.filename)
			}
		})
	}
}

func TestJSScanner_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	var paths []string
	for i := 0; i < 10; i++ {
		name := filepath.Join(dir, "file_"+string(rune('a'+i))+".js")
		if err := os.WriteFile(name, []byte(`function f() { return 1; }`), 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, name)
	}

	checker := &noopJSChecker{}
	sc := NewJSScanner([]rules.JSASTChecker{checker})

	var wg sync.WaitGroup
	for _, p := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			_, _, err := sc.Scan(path)
			if err != nil {
				t.Errorf("concurrent Scan failed: %v", err)
			}
		}(p)
	}
	wg.Wait()

	checker.mu.Lock()
	defer checker.mu.Unlock()
	if len(checker.calls) != 10 {
		t.Errorf("expected 10 calls, got %d", len(checker.calls))
	}
}

func TestJSScanner_NonexistentFile(t *testing.T) {
	sc := NewJSScanner(nil)
	_, _, err := sc.Scan("/nonexistent/file.js")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestJSScanner_UnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "app.go", `package main`)

	sc := NewJSScanner(nil)
	_, _, err := sc.Scan(path)
	if err == nil {
		t.Error("expected error for unsupported extension")
	}
}
