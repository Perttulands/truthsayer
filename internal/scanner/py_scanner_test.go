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

// noopPyChecker is a PyASTChecker that records calls for testing.
type noopPyChecker struct {
	mu    sync.Mutex
	calls []string
}

func (c *noopPyChecker) Meta() rules.Rule {
	return rules.Rule{
		ID:        "test.noop-py",
		Category:  "test",
		Name:      "Noop Python checker",
		FileTypes: []string{".py"},
		ScanType:  rules.ScanTypeAST,
	}
}

func (c *noopPyChecker) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, path)
	return nil
}

// findingPyChecker returns a finding for every file it scans.
type findingPyChecker struct{}

func (c *findingPyChecker) Meta() rules.Rule {
	return rules.Rule{
		ID:        "test.finding-py",
		Category:  "test",
		Name:      "Finding Python checker",
		FileTypes: []string{".py"},
		ScanType:  rules.ScanTypeAST,
	}
}

func (c *findingPyChecker) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	root := tree.RootNode()
	if root == nil {
		return nil
	}
	return []finding.Finding{{
		Rule:    "test.finding-py",
		File:    path,
		Line:    1,
		Message: "test finding",
	}}
}

func TestPyScanner_ScanPyFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "app.py", `def hello():\n    return "world"`)

	checker := &noopPyChecker{}
	sc := NewPyScanner([]rules.PyASTChecker{checker})

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

func TestPyScanner_ReturnsFindings(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "app.py", `def hello(): pass`)

	checker := &findingPyChecker{}
	sc := NewPyScanner([]rules.PyASTChecker{checker})

	findings, _, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "test.finding-py" {
		t.Errorf("expected rule test.finding-py, got %s", findings[0].Rule)
	}
}

func TestPyScanner_MultipleCheckers(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "app.py", `x = 1`)

	noop := &noopPyChecker{}
	fnd := &findingPyChecker{}
	sc := NewPyScanner([]rules.PyASTChecker{noop, fnd})

	findings, _, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(findings) != 1 {
		t.Errorf("expected 1 finding (from findingPyChecker only), got %d", len(findings))
	}
	if len(noop.calls) != 1 {
		t.Errorf("expected noop checker to be called once, got %d", len(noop.calls))
	}
}

func TestPyScanner_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	var paths []string
	for i := 0; i < 10; i++ {
		name := filepath.Join(dir, "file_"+string(rune('a'+i))+".py")
		if err := os.WriteFile(name, []byte(`def f(): return 1`), 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, name)
	}

	checker := &noopPyChecker{}
	sc := NewPyScanner([]rules.PyASTChecker{checker})

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

func TestPyScanner_NonexistentFile(t *testing.T) {
	sc := NewPyScanner(nil)
	_, _, err := sc.Scan("/nonexistent/file.py")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestPyScanner_ParsesModernPython(t *testing.T) {
	dir := t.TempDir()
	// Test with modern Python syntax: type hints, f-strings, walrus operator
	source := `
from typing import Optional

def process(items: list[str], limit: int = 10) -> Optional[str]:
    if (n := len(items)) > limit:
        return f"too many: {n}"
    return None

class Handler:
    async def handle(self, request: dict) -> dict:
        return {"status": "ok"}
`
	path := writeTestFile(t, dir, "modern.py", source)

	checker := &noopPyChecker{}
	sc := NewPyScanner([]rules.PyASTChecker{checker})

	_, lines, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan failed on modern Python: %v", err)
	}
	if len(lines) < 10 {
		t.Errorf("expected at least 10 lines, got %d", len(lines))
	}
	if len(checker.calls) != 1 {
		t.Errorf("expected checker called once, got %d", len(checker.calls))
	}
}
