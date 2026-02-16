package engine

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
)

// testdataPyDir returns the absolute path to testdata/python/.
func testdataPyDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "python")
}

// pyRuleFixtures defines the expected rule-to-fixture mapping for all Python rules.
var pyRuleFixtures = []ruleFixture{
	// Silent-fallback (AST)
	{"silent-fallback.py-bare-except", "bare_except.py", "bare_except_negative.py"},
	{"silent-fallback.py-except-pass", "except_pass.py", "except_pass_negative.py"},
	{"silent-fallback.py-except-broad", "except_broad.py", "except_broad_negative.py"},
	{"silent-fallback.py-subprocess-no-check", "subprocess_no_check.py", "subprocess_no_check_negative.py"},
	{"silent-fallback.py-getattr-silent-default", "getattr_silent_default.py", "getattr_silent_default_negative.py"},
	{"silent-fallback.py-dict-get-none", "dict_get_none.py", "dict_get_none_negative.py"},

	// Error-context (AST)
	{"error-context.py-raise-from-none", "raise_from_none.py", "raise_from_none_negative.py"},
	{"error-context.py-bare-raise-different", "bare_raise_different.py", "bare_raise_different_negative.py"},
	{"error-context.py-generic-exception", "generic_exception.py", "generic_exception_negative.py"},
	{"error-context.py-string-exception", "string_exception.py", "string_exception_negative.py"},
	{"error-context.py-log-and-raise", "log_and_raise.py", "log_and_raise_negative.py"},

	// Bad-defaults (AST)
	{"bad-defaults.py-mutable-default-arg", "mutable_default.py", "mutable_default_negative.py"},
	{"bad-defaults.py-no-timeout-requests", "no_timeout_requests.py", "no_timeout_requests_negative.py"},
	{"bad-defaults.py-star-import", "star_import.py", "star_import_negative.py"},
	{"bad-defaults.py-global-state", "global_state.py", "global_state_negative.py"},
	{"bad-defaults.py-no-encoding-open", "no_encoding_open.py", "no_encoding_open_negative.py"},

	// Mock-leakage (AST)
	{"mock-leakage.py-unittest-import", "unittest_import.py", "unittest_import_negative.py"},
	{"mock-leakage.py-debug-flag", "debug_flag.py", "debug_flag_negative.py"},

	// Trace-gaps (AST)
	{"trace-gaps.py-silent-request", "silent_request.py", "silent_request_negative.py"},

	// Regex rules
	{"trace-gaps.print-debug", "print_debug.py", "print_debug_negative.py"},
	{"trace-gaps.no-logging-config", "no_logging_config.py", "no_logging_config_negative.py"},
	{"mock-leakage.pytest-fixture-in-src", "pytest_fixture_in_src.py", "pytest_fixture_in_src_negative.py"},
	{"bad-defaults.type-ignore-bare", "type_ignore_bare.py", "type_ignore_bare_negative.py"},
	{"bad-defaults.noqa-bare", "noqa_bare.py", "noqa_bare_negative.py"},
	{"config-smells.hardcoded-credentials-py", "hardcoded_credentials.py", "hardcoded_credentials_negative.py"},
	// requirements-unpinned omitted — .txt not in walker's supportedExts, not found in directory scan
}

// requirementsFixture is tested separately since .txt files require ScanFile (not directory walk).
var requirementsFixture = ruleFixture{
	"config-smells.requirements-unpinned", "requirements_unpinned.txt", "requirements_pinned.txt",
}

func TestPyIntegration_PositiveFixtures(t *testing.T) {
	pyDir := testdataPyDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	for _, rf := range pyRuleFixtures {
		t.Run(rf.ruleID, func(t *testing.T) {
			path := filepath.Join(pyDir, rf.positiveFixture)
			result, err := eng.ScanFile(path)
			if err != nil {
				t.Fatalf("ScanFile(%s): %v", rf.positiveFixture, err)
			}

			found := false
			for _, f := range result.Findings {
				if f.Rule == rf.ruleID {
					found = true
					break
				}
			}
			if !found {
				var ruleIDs []string
				for _, f := range result.Findings {
					ruleIDs = append(ruleIDs, f.Rule)
				}
				t.Errorf("expected rule %q in findings for %s, got: %v", rf.ruleID, rf.positiveFixture, ruleIDs)
			}
		})
	}
}

func TestPyIntegration_NegativeFixtures(t *testing.T) {
	pyDir := testdataPyDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	for _, rf := range pyRuleFixtures {
		t.Run(rf.ruleID, func(t *testing.T) {
			path := filepath.Join(pyDir, rf.negativeFixture)
			result, err := eng.ScanFile(path)
			if err != nil {
				t.Fatalf("ScanFile(%s): %v", rf.negativeFixture, err)
			}

			for _, f := range result.Findings {
				if f.Rule == rf.ruleID {
					t.Errorf("unexpected finding %q at line %d in negative fixture %s: %s",
						rf.ruleID, f.Line, rf.negativeFixture, f.Message)
				}
			}
		})
	}
}

func TestPyIntegration_RequirementsUnpinned(t *testing.T) {
	pyDir := testdataPyDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	rf := requirementsFixture

	// Positive: requirements_unpinned.txt should trigger.
	t.Run("positive", func(t *testing.T) {
		path := filepath.Join(pyDir, rf.positiveFixture)
		result, err := eng.ScanFile(path)
		if err != nil {
			t.Fatalf("ScanFile(%s): %v", rf.positiveFixture, err)
		}
		found := false
		for _, f := range result.Findings {
			if f.Rule == rf.ruleID {
				found = true
				break
			}
		}
		if !found {
			var ruleIDs []string
			for _, f := range result.Findings {
				ruleIDs = append(ruleIDs, f.Rule)
			}
			t.Errorf("expected rule %q in findings for %s, got: %v", rf.ruleID, rf.positiveFixture, ruleIDs)
		}
	})

	// Negative: requirements_pinned.txt should be clean.
	t.Run("negative", func(t *testing.T) {
		path := filepath.Join(pyDir, rf.negativeFixture)
		result, err := eng.ScanFile(path)
		if err != nil {
			t.Fatalf("ScanFile(%s): %v", rf.negativeFixture, err)
		}
		for _, f := range result.Findings {
			if f.Rule == rf.ruleID {
				t.Errorf("unexpected finding %q at line %d in negative fixture %s: %s",
					rf.ruleID, f.Line, rf.negativeFixture, f.Message)
			}
		}
	})
}

func TestPyIntegration_DirectoryScan(t *testing.T) {
	pyDir := testdataPyDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(pyDir)
	if err != nil {
		t.Fatalf("Scan(%s): %v", pyDir, err)
	}

	if result.FilesScanned == 0 {
		t.Fatal("expected files to be scanned, got 0")
	}

	// Collect all rule IDs found in the scan.
	foundRules := make(map[string]bool)
	for _, f := range result.Findings {
		foundRules[f.Rule] = true
	}

	// Every Python rule should produce at least one finding across all positive fixtures.
	for _, rf := range pyRuleFixtures {
		if !foundRules[rf.ruleID] {
			t.Errorf("rule %q produced no findings across testdata/python/ directory scan", rf.ruleID)
		}
	}
}

func TestPyIntegration_NoGoJSInterference(t *testing.T) {
	// Scan Go and JS files — verify no Python-specific rules fire on them.
	tmp := t.TempDir()
	goCode := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	goPath := filepath.Join(tmp, "main.go")
	if err := writeFile(goPath, goCode); err != nil {
		t.Fatal(err)
	}

	jsCode := `try { x() } catch (e) {}`
	jsPath := filepath.Join(tmp, "app.js")
	if err := writeFile(jsPath, jsCode); err != nil {
		t.Fatal(err)
	}

	// Python file with anti-patterns.
	pyCode := `
try:
    risky()
except:
    pass
`
	pyPath := filepath.Join(tmp, "app.py")
	if err := writeFile(pyPath, pyCode); err != nil {
		t.Fatal(err)
	}

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if result.FilesScanned != 3 {
		t.Errorf("expected 3 files scanned, got %d", result.FilesScanned)
	}

	// Verify findings are correctly attributed to their files.
	for _, f := range result.Findings {
		if strings.HasSuffix(f.File, ".go") && isPyRuleID(f.Rule) {
			t.Errorf("Python rule %q fired on Go file %s", f.Rule, f.File)
		}
		if strings.HasSuffix(f.File, ".js") && isPyRuleID(f.Rule) {
			t.Errorf("Python rule %q fired on JS file %s", f.Rule, f.File)
		}
		if strings.HasSuffix(f.File, ".py") && isGoRuleID(f.Rule) {
			t.Errorf("Go rule %q fired on Python file %s", f.Rule, f.File)
		}
		if strings.HasSuffix(f.File, ".py") && isJSRuleID(f.Rule) {
			t.Errorf("JS rule %q fired on Python file %s", f.Rule, f.File)
		}
	}

	// Python bare-except should fire on the Python file.
	hasPyFinding := false
	for _, f := range result.Findings {
		if f.Rule == "silent-fallback.py-bare-except" && strings.HasSuffix(f.File, "app.py") {
			hasPyFinding = true
			break
		}
	}
	if !hasPyFinding {
		t.Error("expected silent-fallback.py-bare-except finding on app.py")
	}
}

func TestPyIntegration_SeverityLevels(t *testing.T) {
	pyDir := testdataPyDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(pyDir)
	if err != nil {
		t.Fatal(err)
	}

	// Verify we get findings at all three severity levels from Python rules.
	severities := make(map[finding.Severity]bool)
	for _, f := range result.Findings {
		if isPyRuleID(f.Rule) {
			severities[f.Severity] = true
		}
	}
	for _, sev := range []finding.Severity{finding.SeverityError, finding.SeverityWarning, finding.SeverityInfo} {
		if !severities[sev] {
			t.Errorf("expected Python findings with severity %q, found none", sev)
		}
	}
}

// isPyRuleID returns true if the rule ID belongs to a Python-specific rule.
func isPyRuleID(id string) bool {
	return strings.Contains(id, ".py-") ||
		id == "trace-gaps.print-debug" ||
		id == "trace-gaps.no-logging-config" ||
		id == "mock-leakage.pytest-fixture-in-src" ||
		id == "bad-defaults.type-ignore-bare" ||
		id == "bad-defaults.noqa-bare" ||
		id == "config-smells.hardcoded-credentials-py" ||
		id == "config-smells.requirements-unpinned"
}
