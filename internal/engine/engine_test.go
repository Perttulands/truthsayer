package engine

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/config"
	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
)

type countingRegexChecker struct {
	mu    sync.Mutex
	calls int
}

func (c *countingRegexChecker) Meta() rules.Rule {
	return rules.Rule{
		ID:        "test.counter",
		Category:  "test",
		Name:      "counter",
		Severity:  finding.SeverityInfo,
		FileTypes: []string{".sh"},
		ScanType:  rules.ScanTypeRegex,
	}
}

func (c *countingRegexChecker) CheckLines(path string, lines []string) []finding.Finding {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
	return nil
}

func (c *countingRegexChecker) CallCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

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

func TestEngine_ScanFile_Go(t *testing.T) {
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
	if result.FilesScanned != 1 {
		t.Errorf("expected 1 file scanned, got %d", result.FilesScanned)
	}
	if len(result.Findings) == 0 {
		t.Error("expected findings for bad.go, got 0")
	}
	// All findings should reference this file
	for _, f := range result.Findings {
		if f.File != path {
			t.Errorf("finding file = %q, want %q", f.File, path)
		}
	}
}

func TestEngine_ScanFile_Bash(t *testing.T) {
	tmp := t.TempDir()
	bashCode := "#!/bin/bash\necho hello\n"
	path := filepath.Join(tmp, "no_pipefail.sh")
	os.WriteFile(path, []byte(bashCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.FilesScanned != 1 {
		t.Errorf("expected 1 file scanned, got %d", result.FilesScanned)
	}
	if len(result.Findings) == 0 {
		t.Error("expected findings for no_pipefail.sh, got 0")
	}
}

func TestEngine_ScanFile_NotExist(t *testing.T) {
	reg := rules.DefaultRegistry()
	eng := New(reg)

	_, err := eng.ScanFile("/nonexistent/file.go")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestEngine_ScanSkipsExcludedDirs(t *testing.T) {
	tmp := t.TempDir()

	// Create a file with anti-patterns in a normal directory
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
	os.MkdirAll(filepath.Join(tmp, "src"), 0o755)
	os.WriteFile(filepath.Join(tmp, "src", "main.go"), []byte(goCode), 0o644)

	// Put identical anti-pattern files inside excluded directories
	for _, dir := range []string{"vendor", "node_modules", ".git"} {
		os.MkdirAll(filepath.Join(tmp, dir), 0o755)
		os.WriteFile(filepath.Join(tmp, dir, "bad.go"), []byte(goCode), 0o644)
	}

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}

	// Only src/main.go should be scanned
	if result.FilesScanned != 1 {
		t.Errorf("expected 1 file scanned (src/main.go only), got %d", result.FilesScanned)
	}

	// All findings should reference src/main.go, not excluded dirs
	for _, f := range result.Findings {
		for _, dir := range []string{"vendor", "node_modules", ".git"} {
			if filepath.Base(filepath.Dir(f.File)) == dir {
				t.Errorf("finding from excluded dir %q: %s", dir, f.File)
			}
		}
	}
}

func TestEngine_ScanFile_JS(t *testing.T) {
	tmp := t.TempDir()
	jsCode := `
function doWork() {
    try {
        riskyOp();
    } catch (e) {
        console.error(e);
        throw e;
    }
}
`
	path := filepath.Join(tmp, "app.js")
	os.WriteFile(path, []byte(jsCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.FilesScanned != 1 {
		t.Errorf("expected 1 file scanned, got %d", result.FilesScanned)
	}
	// Engine should not error — it routes to JSScanner
}

func TestEngine_ScanFile_TS(t *testing.T) {
	tmp := t.TempDir()
	tsCode := `
interface User {
    name: string;
    age: number;
}

function greet(user: User): string {
    return "Hello, " + user.name;
}
`
	path := filepath.Join(tmp, "app.ts")
	os.WriteFile(path, []byte(tsCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.FilesScanned != 1 {
		t.Errorf("expected 1 file scanned, got %d", result.FilesScanned)
	}
}

func TestEngine_ScanFile_Python(t *testing.T) {
	tmp := t.TempDir()
	pyCode := `
def greet(name):
    return f"Hello, {name}"

if __name__ == "__main__":
    print(greet("world"))
`
	path := filepath.Join(tmp, "app.py")
	os.WriteFile(path, []byte(pyCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.FilesScanned != 1 {
		t.Errorf("expected 1 file scanned, got %d", result.FilesScanned)
	}
}

func TestEngine_RoutesFilesByExtension(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main\n"), 0o644)
	os.WriteFile(filepath.Join(tmp, "app.js"), []byte("console.log('hi');\n"), 0o644)
	os.WriteFile(filepath.Join(tmp, "types.ts"), []byte("const x: number = 1;\n"), 0o644)
	os.WriteFile(filepath.Join(tmp, "script.py"), []byte("print('hi')\n"), 0o644)
	os.WriteFile(filepath.Join(tmp, "deploy.sh"), []byte("#!/bin/bash\necho hi\n"), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if result.FilesScanned != 5 {
		t.Errorf("expected 5 files scanned, got %d", result.FilesScanned)
	}
}

func TestEngine_ScanSkipsNewExcludedDirs(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "src"), 0o755)
	os.WriteFile(filepath.Join(tmp, "src", "main.py"), []byte("print('hi')\n"), 0o644)

	for _, dir := range []string{"__pycache__", ".venv", "dist", "build"} {
		os.MkdirAll(filepath.Join(tmp, dir), 0o755)
		os.WriteFile(filepath.Join(tmp, dir, "cached.py"), []byte("print('cached')\n"), 0o644)
	}

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if result.FilesScanned != 1 {
		t.Errorf("expected 1 file scanned (src/main.py only), got %d", result.FilesScanned)
	}
}

func TestEngine_LazyInit_NoJSFiles(t *testing.T) {
	// When scanning only Go files, JS/Python scanners should not be initialized
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main\n"), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)

	_, err := eng.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}

	// js and py fields should still be nil (lazy init not triggered)
	if eng.js != nil {
		t.Error("JSScanner should not be initialized when no JS files are scanned")
	}
	if eng.py != nil {
		t.Error("PyScanner should not be initialized when no Python files are scanned")
	}
}

func TestIsJSExt(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".js", true},
		{".jsx", true},
		{".ts", true},
		{".tsx", true},
		{".mjs", true},
		{".cjs", true},
		{".go", false},
		{".py", false},
		{".sh", false},
	}
	for _, tt := range tests {
		if got := isJSExt(tt.ext); got != tt.want {
			t.Errorf("isJSExt(%q) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}

func TestIsPyExt(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".py", true},
		{".pyi", true},
		{".go", false},
		{".js", false},
	}
	for _, tt := range tests {
		if got := isPyExt(tt.ext); got != tt.want {
			t.Errorf("isPyExt(%q) = %v, want %v", tt.ext, got, tt.want)
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

func boolPtr(b bool) *bool { return &b }

func TestEngine_DisablePython_NoFindings(t *testing.T) {
	tmp := t.TempDir()
	// Python file with anti-pattern (bare except)
	pyCode := "try:\n    x = 1\nexcept:\n    pass\n"
	path := filepath.Join(tmp, "bad.py")
	os.WriteFile(path, []byte(pyCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)
	eng.SetLanguages(&config.LanguageConfig{Python: boolPtr(false)})

	result, err := eng.ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings with python disabled, got %d", len(result.Findings))
	}
}

func TestEngine_DisableJS_NoFindings(t *testing.T) {
	tmp := t.TempDir()
	// JS file with anti-pattern (empty catch)
	jsCode := "try { risky(); } catch (e) {}\n"
	path := filepath.Join(tmp, "bad.js")
	os.WriteFile(path, []byte(jsCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)
	eng.SetLanguages(&config.LanguageConfig{JavaScript: boolPtr(false)})

	result, err := eng.ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings with javascript disabled, got %d", len(result.Findings))
	}
}

func TestEngine_DisableGo_NoFindings(t *testing.T) {
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
	eng.SetLanguages(&config.LanguageConfig{Go: boolPtr(false)})

	result, err := eng.ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings with go disabled, got %d", len(result.Findings))
	}
}

func TestEngine_DisablePython_ScanDir_NoPyFindings(t *testing.T) {
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
	// Python file with anti-pattern
	pyCode := "try:\n    x = 1\nexcept:\n    pass\n"
	os.WriteFile(filepath.Join(tmp, "bad.py"), []byte(pyCode), 0o644)

	reg := rules.DefaultRegistry()
	eng := New(reg)
	eng.SetLanguages(&config.LanguageConfig{Python: boolPtr(false)})

	result, err := eng.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}

	// Should have Go findings but no Python findings
	for _, f := range result.Findings {
		if filepath.Ext(f.File) == ".py" {
			t.Errorf("found Python finding with python disabled: %s", f.Rule)
		}
	}
	if len(result.Findings) == 0 {
		t.Error("expected Go findings, got 0")
	}
}

func TestEngine_NilLangs_AllEnabled(t *testing.T) {
	reg := rules.DefaultRegistry()
	eng := New(reg)
	// No SetLanguages call — all should be enabled
	if !eng.langEnabled("go") {
		t.Error("go should be enabled by default")
	}
	if !eng.langEnabled("python") {
		t.Error("python should be enabled by default")
	}
	if !eng.langEnabled("javascript") {
		t.Error("javascript should be enabled by default")
	}
}

func TestExtLang(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".go", "go"},
		{".js", "javascript"},
		{".jsx", "javascript"},
		{".mjs", "javascript"},
		{".cjs", "javascript"},
		{".ts", "typescript"},
		{".tsx", "typescript"},
		{".py", "python"},
		{".pyi", "python"},
		{".sh", "bash"},
		{".bash", "bash"},
		{".toml", ""},
		{".json", ""},
	}
	for _, tt := range tests {
		if got := extLang(tt.ext); got != tt.want {
			t.Errorf("extLang(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}
}

func TestEngine_FileCacheByMtime(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "script.sh")
	if err := os.WriteFile(path, []byte("#!/bin/bash\necho hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	reg := rules.NewRegistry()
	counter := &countingRegexChecker{}
	reg.RegisterRegex(counter)
	eng := New(reg)

	if _, err := eng.ScanFile(path); err != nil {
		t.Fatal(err)
	}
	if counter.CallCount() != 1 {
		t.Fatalf("expected checker to run once, got %d", counter.CallCount())
	}

	if _, err := eng.ScanFile(path); err != nil {
		t.Fatal(err)
	}
	if counter.CallCount() != 1 {
		t.Fatalf("expected cached result on second scan, got %d calls", counter.CallCount())
	}

	if err := os.WriteFile(path, []byte("#!/bin/bash\necho changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(path, now, now); err != nil {
		t.Fatal(err)
	}

	if _, err := eng.ScanFile(path); err != nil {
		t.Fatal(err)
	}
	if counter.CallCount() != 2 {
		t.Fatalf("expected cache invalidation after mtime change, got %d calls", counter.CallCount())
	}
}

func TestEngine_SetParallelism(t *testing.T) {
	reg := rules.NewRegistry()
	eng := New(reg)
	eng.SetParallelism(7)
	if eng.parallelism != 7 {
		t.Fatalf("expected parallelism=7, got %d", eng.parallelism)
	}

	eng.SetParallelism(0)
	if eng.parallelism != 7 {
		t.Fatalf("parallelism should remain unchanged on invalid value, got %d", eng.parallelism)
	}
}
