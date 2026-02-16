package engine

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/perttulands/truthsayer/internal/config"
	"github.com/perttulands/truthsayer/internal/rules"
)

// testdataMixedDir returns the absolute path to testdata/mixed/.
func testdataMixedDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "mixed")
}

func TestCrossLanguage_MixedDirectoryScan(t *testing.T) {
	mixedDir := testdataMixedDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(mixedDir)
	if err != nil {
		t.Fatalf("Scan(%s): %v", mixedDir, err)
	}

	if result.FilesScanned != 4 {
		t.Errorf("expected 4 files scanned (.go, .js, .ts, .py), got %d", result.FilesScanned)
	}

	// Collect findings by file extension.
	goFindings := 0
	jsFindings := 0
	tsFindings := 0
	pyFindings := 0
	for _, f := range result.Findings {
		ext := filepath.Ext(f.File)
		switch ext {
		case ".go":
			goFindings++
		case ".js":
			jsFindings++
		case ".ts":
			tsFindings++
		case ".py":
			pyFindings++
		}
	}

	// Each language should produce at least one finding from the mixed fixtures.
	if jsFindings == 0 {
		t.Error("expected JS findings from app.js (empty catch, eval), got 0")
	}
	if tsFindings == 0 {
		t.Error("expected TS findings from handler.ts (as any), got 0")
	}
	if pyFindings == 0 {
		t.Error("expected Python findings from process.py (bare except, mutable default), got 0")
	}

	// Verify specific expected rules fire.
	expectedRules := map[string]bool{
		"silent-fallback.js-empty-catch":   false,
		"bad-defaults.eval-usage":          false,
		"bad-defaults.any-type-assertion":  false,
		"silent-fallback.py-bare-except":   false,
		"silent-fallback.py-except-pass":   false,
		"bad-defaults.py-mutable-default-arg": false,
	}
	for _, f := range result.Findings {
		if _, ok := expectedRules[f.Rule]; ok {
			expectedRules[f.Rule] = true
		}
	}
	for rule, found := range expectedRules {
		if !found {
			t.Errorf("expected rule %q in mixed directory scan findings", rule)
		}
	}
}

func TestCrossLanguage_NoCrossContamination(t *testing.T) {
	mixedDir := testdataMixedDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(mixedDir)
	if err != nil {
		t.Fatalf("Scan(%s): %v", mixedDir, err)
	}

	for _, f := range result.Findings {
		ext := filepath.Ext(f.File)
		switch {
		case ext == ".go":
			if isJSRuleID(f.Rule) {
				t.Errorf("JS rule %q fired on Go file %s", f.Rule, f.File)
			}
			if isPyRuleID(f.Rule) {
				t.Errorf("Python rule %q fired on Go file %s", f.Rule, f.File)
			}
		case ext == ".js" || ext == ".ts":
			if isGoRuleID(f.Rule) {
				t.Errorf("Go rule %q fired on %s file %s", f.Rule, ext, f.File)
			}
			if isPyRuleID(f.Rule) {
				t.Errorf("Python rule %q fired on %s file %s", f.Rule, ext, f.File)
			}
		case ext == ".py":
			if isGoRuleID(f.Rule) {
				t.Errorf("Go rule %q fired on Python file %s", f.Rule, f.File)
			}
			if isJSRuleID(f.Rule) {
				t.Errorf("JS rule %q fired on Python file %s", f.Rule, f.File)
			}
		}
	}
}

func TestCrossLanguage_DisablePython(t *testing.T) {
	mixedDir := testdataMixedDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)
	eng.SetLanguages(&config.LanguageConfig{Python: boolPtr(false)})

	result, err := eng.Scan(mixedDir)
	if err != nil {
		t.Fatalf("Scan(%s): %v", mixedDir, err)
	}

	// No Python findings should appear.
	for _, f := range result.Findings {
		if filepath.Ext(f.File) == ".py" {
			t.Errorf("Python finding %q on %s with python disabled", f.Rule, f.File)
		}
	}

	// JS and Go findings should still appear.
	hasJS := false
	for _, f := range result.Findings {
		ext := filepath.Ext(f.File)
		if ext == ".js" || ext == ".ts" {
			hasJS = true
			break
		}
	}
	if !hasJS {
		t.Error("expected JS/TS findings even with python disabled")
	}
}

func TestCrossLanguage_DisableJS(t *testing.T) {
	mixedDir := testdataMixedDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)
	eng.SetLanguages(&config.LanguageConfig{
		JavaScript: boolPtr(false),
		TypeScript: boolPtr(false),
	})

	result, err := eng.Scan(mixedDir)
	if err != nil {
		t.Fatalf("Scan(%s): %v", mixedDir, err)
	}

	// No JS/TS findings should appear.
	for _, f := range result.Findings {
		ext := filepath.Ext(f.File)
		if ext == ".js" || ext == ".ts" {
			t.Errorf("JS/TS finding %q on %s with javascript/typescript disabled", f.Rule, f.File)
		}
	}

	// Python findings should still appear.
	hasPy := false
	for _, f := range result.Findings {
		if filepath.Ext(f.File) == ".py" {
			hasPy = true
			break
		}
	}
	if !hasPy {
		t.Error("expected Python findings even with JS/TS disabled")
	}
}

func TestCrossLanguage_LangFilterGoOnly(t *testing.T) {
	// Simulates --lang go: only Go is enabled, everything else disabled.
	mixedDir := testdataMixedDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)
	eng.SetLanguages(&config.LanguageConfig{
		Go:         boolPtr(true),
		JavaScript: boolPtr(false),
		TypeScript: boolPtr(false),
		Python:     boolPtr(false),
		Bash:       boolPtr(false),
	})

	result, err := eng.Scan(mixedDir)
	if err != nil {
		t.Fatalf("Scan(%s): %v", mixedDir, err)
	}

	// Only Go findings should appear.
	for _, f := range result.Findings {
		ext := filepath.Ext(f.File)
		if ext != ".go" {
			t.Errorf("non-Go finding %q on %s with --lang go", f.Rule, f.File)
		}
	}
}

func TestCrossLanguage_LangFilterPythonOnly(t *testing.T) {
	// Simulates --lang python: only Python is enabled.
	mixedDir := testdataMixedDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)
	eng.SetLanguages(&config.LanguageConfig{
		Go:         boolPtr(false),
		JavaScript: boolPtr(false),
		TypeScript: boolPtr(false),
		Python:     boolPtr(true),
		Bash:       boolPtr(false),
	})

	result, err := eng.Scan(mixedDir)
	if err != nil {
		t.Fatalf("Scan(%s): %v", mixedDir, err)
	}

	// Only Python findings should appear.
	for _, f := range result.Findings {
		ext := filepath.Ext(f.File)
		if ext != ".py" {
			t.Errorf("non-Python finding %q on %s with --lang python", f.Rule, f.File)
		}
	}

	// Verify we do get Python findings.
	hasPy := false
	for _, f := range result.Findings {
		if strings.Contains(f.Rule, ".py-") {
			hasPy = true
			break
		}
	}
	if !hasPy {
		t.Error("expected Python-specific findings with --lang python")
	}
}

func TestCrossLanguage_AllLanguagesEnabled(t *testing.T) {
	// Explicit enable of all languages produces same result as default (nil langs).
	mixedDir := testdataMixedDir(t)
	reg := rules.DefaultRegistry()

	// Default (nil langs).
	eng1 := New(reg)
	result1, err := eng1.Scan(mixedDir)
	if err != nil {
		t.Fatal(err)
	}

	// Explicit all-enabled.
	eng2 := New(reg)
	eng2.SetLanguages(&config.LanguageConfig{
		Go:         boolPtr(true),
		JavaScript: boolPtr(true),
		TypeScript: boolPtr(true),
		Python:     boolPtr(true),
		Bash:       boolPtr(true),
	})
	result2, err := eng2.Scan(mixedDir)
	if err != nil {
		t.Fatal(err)
	}

	if result1.FilesScanned != result2.FilesScanned {
		t.Errorf("files scanned differ: default=%d, explicit=%d", result1.FilesScanned, result2.FilesScanned)
	}
	if len(result1.Findings) != len(result2.Findings) {
		t.Errorf("finding count differs: default=%d, explicit=%d", len(result1.Findings), len(result2.Findings))
	}
}
