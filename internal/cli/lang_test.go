package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseLangFlag_SingleLanguage(t *testing.T) {
	tests := []struct {
		input    string
		wantLang string // canonical name to check is enabled
	}{
		{"go", "go"},
		{"python", "python"},
		{"py", "python"},
		{"js", "javascript"},
		{"javascript", "javascript"},
		{"ts", "typescript"},
		{"typescript", "typescript"},
		{"bash", "bash"},
		{"shell", "bash"},
		{"sh", "bash"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lc, err := parseLangFlag(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !lc.IsEnabled(tt.wantLang) {
				t.Errorf("expected %s to be enabled", tt.wantLang)
			}
		})
	}
}

func TestParseLangFlag_MultipleLanguages(t *testing.T) {
	lc, err := parseLangFlag("go,python")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !lc.IsEnabled("go") {
		t.Error("expected go to be enabled")
	}
	if !lc.IsEnabled("python") {
		t.Error("expected python to be enabled")
	}
	if lc.IsEnabled("javascript") {
		t.Error("expected javascript to be disabled")
	}
	if lc.IsEnabled("typescript") {
		t.Error("expected typescript to be disabled")
	}
	if lc.IsEnabled("bash") {
		t.Error("expected bash to be disabled")
	}
}

func TestParseLangFlag_JSAndTS(t *testing.T) {
	lc, err := parseLangFlag("js,ts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !lc.IsEnabled("javascript") {
		t.Error("expected javascript to be enabled")
	}
	if !lc.IsEnabled("typescript") {
		t.Error("expected typescript to be enabled")
	}
	if lc.IsEnabled("go") {
		t.Error("expected go to be disabled")
	}
}

func TestParseLangFlag_CaseInsensitive(t *testing.T) {
	lc, err := parseLangFlag("Go,PYTHON")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !lc.IsEnabled("go") {
		t.Error("expected go to be enabled")
	}
	if !lc.IsEnabled("python") {
		t.Error("expected python to be enabled")
	}
}

func TestParseLangFlag_WithSpaces(t *testing.T) {
	lc, err := parseLangFlag("go, python")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !lc.IsEnabled("go") {
		t.Error("expected go to be enabled")
	}
	if !lc.IsEnabled("python") {
		t.Error("expected python to be enabled")
	}
}

func TestParseLangFlag_UnknownLanguage(t *testing.T) {
	_, err := parseLangFlag("rust")
	if err == nil {
		t.Fatal("expected error for unknown language")
	}
	if !strings.Contains(err.Error(), "unknown language") {
		t.Errorf("expected 'unknown language' in error, got: %v", err)
	}
}

func TestParseLangFlag_Empty(t *testing.T) {
	_, err := parseLangFlag("")
	if err == nil {
		t.Fatal("expected error for empty --lang")
	}
}

func TestLangFilterExts_Go(t *testing.T) {
	exts, err := langFilterExts("go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exts[".go"] {
		t.Error("expected .go extension")
	}
	if exts[".py"] {
		t.Error("unexpected .py extension")
	}
}

func TestLangFilterExts_JS(t *testing.T) {
	exts, err := langFilterExts("js")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, ext := range []string{".js", ".jsx", ".mjs", ".cjs"} {
		if !exts[ext] {
			t.Errorf("expected %s extension", ext)
		}
	}
}

func TestLangFilterExts_Multiple(t *testing.T) {
	exts, err := langFilterExts("go,python")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exts[".go"] {
		t.Error("expected .go extension")
	}
	if !exts[".py"] {
		t.Error("expected .py extension")
	}
	if !exts[".pyi"] {
		t.Error("expected .pyi extension")
	}
}

func TestRules_LangFilterPython(t *testing.T) {
	out := captureStdout(t, func() {
		code := runRules([]string{"--lang", "python"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	// Python rules should appear
	if !strings.Contains(out, "silent-fallback.py-bare-except") {
		t.Error("output missing Python rule silent-fallback.py-bare-except")
	}
	if !strings.Contains(out, ".py") {
		t.Error("output missing .py file type")
	}

	// Go rules should NOT appear
	if strings.Contains(out, "silent-fallback.empty-error-check") {
		t.Error("Go rule should not appear in --lang python output")
	}

	// JS rules should NOT appear
	if strings.Contains(out, "silent-fallback.js-empty-catch") {
		t.Error("JS rule should not appear in --lang python output")
	}
}

func TestRules_LangFilterGo(t *testing.T) {
	out := captureStdout(t, func() {
		code := runRules([]string{"--lang", "go"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "silent-fallback.empty-error-check") {
		t.Error("output missing Go rule")
	}
	// Python rules should NOT appear
	if strings.Contains(out, "silent-fallback.py-bare-except") {
		t.Error("Python rule should not appear in --lang go output")
	}
}

func TestRules_LangFilterJS(t *testing.T) {
	out := captureStdout(t, func() {
		code := runRules([]string{"--lang", "js"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "silent-fallback.js-empty-catch") {
		t.Error("output missing JS rule")
	}
	if strings.Contains(out, "silent-fallback.empty-error-check") {
		t.Error("Go rule should not appear in --lang js output")
	}
}

func TestRules_LangFilterMultiple(t *testing.T) {
	out := captureStdout(t, func() {
		code := runRules([]string{"--lang", "go,python"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "silent-fallback.empty-error-check") {
		t.Error("output missing Go rule")
	}
	if !strings.Contains(out, "silent-fallback.py-bare-except") {
		t.Error("output missing Python rule")
	}
	if strings.Contains(out, "silent-fallback.js-empty-catch") {
		t.Error("JS rule should not appear in --lang go,python output")
	}
}

func TestRules_LangFilterWithEnabled(t *testing.T) {
	origDir, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	out := captureStdout(t, func() {
		code := runRules([]string{"--lang", "python", "--enabled"})
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "rules enabled") {
		t.Errorf("expected 'rules enabled' label, got: %s", out)
	}
	if !strings.Contains(out, "silent-fallback.py-bare-except") {
		t.Error("output missing Python rule")
	}
	if strings.Contains(out, "silent-fallback.empty-error-check") {
		t.Error("Go rule should not appear in --lang python --enabled output")
	}
}

func TestRules_LangFilterUnknown(t *testing.T) {
	// Capture stderr for error output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	code := runRules([]string{"--lang", "rust"})

	w.Close()
	os.Stderr = oldStderr

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	errOut := string(buf[:n])

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(errOut, "unknown language") {
		t.Errorf("expected unknown language error, got: %s", errOut)
	}
}

func TestParseScanOptions_LangFlag(t *testing.T) {
	opts, err := parseScanOptions([]string{"--lang", "go,python", "/tmp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.lang != "go,python" {
		t.Errorf("expected lang='go,python', got %q", opts.lang)
	}
	if opts.path != "/tmp" {
		t.Errorf("expected path='/tmp', got %q", opts.path)
	}
}

func TestParseScanOptions_LangFlagMissing(t *testing.T) {
	_, err := parseScanOptions([]string{"--lang"})
	if err == nil {
		t.Fatal("expected error for --lang without value")
	}
	if !strings.Contains(err.Error(), "--lang requires a value") {
		t.Errorf("expected '--lang requires a value' error, got: %v", err)
	}
}

func TestParseScanOptions_ParallelFlag_DefaultWorkers(t *testing.T) {
	opts, err := parseScanOptions([]string{"--parallel", "/tmp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.parallel != runtime.NumCPU() {
		t.Fatalf("expected parallel workers=%d, got %d", runtime.NumCPU(), opts.parallel)
	}
	if opts.path != "/tmp" {
		t.Fatalf("expected path=/tmp, got %q", opts.path)
	}
}

func TestParseScanOptions_ParallelFlag_WithValue(t *testing.T) {
	opts, err := parseScanOptions([]string{"--parallel", "3", "/tmp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.parallel != 3 {
		t.Fatalf("expected parallel workers=3, got %d", opts.parallel)
	}
}

func TestParseScanOptions_ParallelFlag_Invalid(t *testing.T) {
	_, err := parseScanOptions([]string{"--parallel", "0", "/tmp"})
	if err == nil {
		t.Fatal("expected error for invalid parallel value")
	}
	if !strings.Contains(err.Error(), "--parallel must be >= 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScan_LangFilterPythonOnly(t *testing.T) {
	// Create a temporary directory with Go and Python files
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.go"), []byte(`package main
func main() {
	val, _ := foo()
	_ = val
}
func foo() (int, error) { return 0, nil }
`), 0644)
	os.WriteFile(filepath.Join(dir, "bad.py"), []byte(`
try:
    do_something()
except:
    pass
`), 0644)

	out := captureStdout(t, func() {
		code := runScan([]string{"--lang", "python", dir})
		// Exit code 1 (Python errors found) or 0 (no errors)
		if code == 2 {
			t.Errorf("unexpected tool error, exit code 2")
		}
	})

	// Python findings should appear
	if !strings.Contains(out, "py-bare-except") && !strings.Contains(out, "py-except-pass") {
		t.Error("expected Python findings in output")
	}
	// Go findings should NOT appear
	if strings.Contains(out, "ignored-error") || strings.Contains(out, "empty-error-check") {
		t.Error("Go findings should not appear with --lang python")
	}
}
