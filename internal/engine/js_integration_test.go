package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
)

// testdataJSDir returns the absolute path to testdata/js/.
func testdataJSDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "js")
}

// ruleFixture maps a rule ID to its positive fixture file (relative to testdata/js/).
type ruleFixture struct {
	ruleID          string
	positiveFixture string
	negativeFixture string
}

// jsRuleFixtures defines the expected rule-to-fixture mapping for all JS/TS rules.
var jsRuleFixtures = []ruleFixture{
	// Silent-fallback (AST)
	{"silent-fallback.js-empty-catch", "empty_catch.js", "empty_catch_negative.js"},
	{"silent-fallback.js-catch-return-null", "catch_return_null.js", "catch_return_null_negative.js"},
	{"silent-fallback.js-floating-promise", "floating_promise.ts", "floating_promise_negative.ts"},
	{"silent-fallback.js-callback-err-ignored", "callback_err_ignored.js", "callback_err_ignored_negative.js"},
	{"silent-fallback.js-optional-chain-silence", "optional_chain_silence.js", "optional_chain_silence_negative.js"},

	// Error-context (AST)
	{"error-context.js-rethrow-no-wrap", "rethrow_no_wrap.js", "rethrow_no_wrap_negative.js"},
	{"error-context.js-generic-error-message", "generic_error_message.js", "generic_error_message_negative.js"},
	{"error-context.js-promise-reject-non-error", "promise_reject.js", "promise_reject_negative.js"},
	{"error-context.js-console-error-no-throw", "console_error_no_throw.js", "console_error_no_throw_negative.js"},
	{"error-context.js-http-200-on-error", "http_200_on_error.js", "http_200_on_error_negative.js"},

	// Trace-gaps (AST)
	{"trace-gaps.js-no-error-handler-express", "no_error_handler_express.js", "no_error_handler_express_negative.js"},
	{"trace-gaps.js-missing-correlation-id", "missing_correlation_id.js", "missing_correlation_id_negative.js"},

	// Mock-leakage (AST)
	{"mock-leakage.js-test-import-in-src", "test_import_in_src.js", "test_import_in_src_negative.js"},
	{"mock-leakage.js-env-test-check", "env_test_check.js", "env_test_check_negative.js"},

	// Bad-defaults (AST)
	{"bad-defaults.no-timeout-fetch", "no_timeout_fetch.js", "no_timeout_fetch_negative.js"},
	{"bad-defaults.any-type-assertion", "any_assertion.ts", "any_assertion_negative.ts"},
	{"bad-defaults.non-null-assertion", "non_null_assertion.ts", "non_null_assertion_negative.ts"},
	{"bad-defaults.eval-usage", "eval_usage.js", "eval_usage_negative.js"},

	// Test-isolation (AST)
	{"test-isolation.no-afterall-cleanup", "no_afterall_cleanup.test.js", "no_afterall_cleanup_negative.test.js"},
	{"test-isolation.test-only-import", "test_only_import.js", "test_only_import_negative.js"},

	// Regex rules
	{"mock-leakage.jest-mock-in-src", "jest_mock_in_src.js", "jest_mock_in_src_negative.test.js"},
	{"mock-leakage.storybook-in-src", "storybook_in_src.tsx", "storybook_in_src_negative.stories.tsx"},
	{"bad-defaults.ts-ignore", "ts_ignore.ts", "ts_ignore_negative.ts"},
	{"bad-defaults.eslint-disable-no-reason", "eslint_disable.js", "eslint_disable_negative.js"},
	{"bad-defaults.no-strict-mode", "no_strict_mode.cjs", "no_strict_mode_negative.cjs"},
	{"trace-gaps.no-unhandled-rejection", "no_unhandled_rejection.js", "no_unhandled_rejection_negative.js"},
	{"trace-gaps.console-log-in-production", "console_log_production.ts", "console_log_production_negative.test.ts"},
	{"config-smells.hardcoded-api-url", "hardcoded_api_url.ts", "hardcoded_api_url_negative.ts"},
	// dotenv-no-example is omitted — requires filesystem context (.env without .env.example), not a single-file scan
}

func TestJSIntegration_PositiveFixtures(t *testing.T) {
	jsDir := testdataJSDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	for _, rf := range jsRuleFixtures {
		t.Run(rf.ruleID, func(t *testing.T) {
			path := filepath.Join(jsDir, rf.positiveFixture)
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

func TestJSIntegration_NegativeFixtures(t *testing.T) {
	jsDir := testdataJSDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	for _, rf := range jsRuleFixtures {
		t.Run(rf.ruleID, func(t *testing.T) {
			path := filepath.Join(jsDir, rf.negativeFixture)
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

func TestJSIntegration_DirectoryScan(t *testing.T) {
	jsDir := testdataJSDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(jsDir)
	if err != nil {
		t.Fatalf("Scan(%s): %v", jsDir, err)
	}

	if result.FilesScanned == 0 {
		t.Fatal("expected files to be scanned, got 0")
	}

	// Collect all rule IDs found in the scan.
	foundRules := make(map[string]bool)
	for _, f := range result.Findings {
		foundRules[f.Rule] = true
	}

	// Every JS/TS rule should produce at least one finding across all positive fixtures.
	for _, rf := range jsRuleFixtures {
		if !foundRules[rf.ruleID] {
			t.Errorf("rule %q produced no findings across testdata/js/ directory scan", rf.ruleID)
		}
	}
}

func TestJSIntegration_NoGoInterference(t *testing.T) {
	// Scan a Go file — verify no JS-specific rules fire on it.
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

	// Also put a JS file with anti-patterns alongside it.
	jsCode := `try { x() } catch (e) {}`
	jsPath := filepath.Join(tmp, "app.js")
	if err := writeFile(jsPath, jsCode); err != nil {
		t.Fatal(err)
	}

	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if result.FilesScanned != 2 {
		t.Errorf("expected 2 files scanned, got %d", result.FilesScanned)
	}

	// Verify findings are correctly attributed to their files.
	for _, f := range result.Findings {
		if strings.HasSuffix(f.File, ".go") && isJSRuleID(f.Rule) {
			t.Errorf("JS rule %q fired on Go file %s", f.Rule, f.File)
		}
		if strings.HasSuffix(f.File, ".js") && isGoRuleID(f.Rule) {
			t.Errorf("Go rule %q fired on JS file %s", f.Rule, f.File)
		}
	}

	// JS empty-catch should fire on the JS file.
	hasJSFinding := false
	for _, f := range result.Findings {
		if f.Rule == "silent-fallback.js-empty-catch" && strings.HasSuffix(f.File, "app.js") {
			hasJSFinding = true
			break
		}
	}
	if !hasJSFinding {
		t.Error("expected silent-fallback.js-empty-catch finding on app.js")
	}
}

func TestJSIntegration_SeverityLevels(t *testing.T) {
	jsDir := testdataJSDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	result, err := eng.Scan(jsDir)
	if err != nil {
		t.Fatal(err)
	}

	// Verify we get findings at all three severity levels.
	severities := make(map[finding.Severity]bool)
	for _, f := range result.Findings {
		severities[f.Severity] = true
	}
	for _, sev := range []finding.Severity{finding.SeverityError, finding.SeverityWarning, finding.SeverityInfo} {
		if !severities[sev] {
			t.Errorf("expected findings with severity %q, found none", sev)
		}
	}
}

// isJSRuleID returns true if the rule ID belongs to a JS/TS-specific rule.
func isJSRuleID(id string) bool {
	return strings.Contains(id, ".js-") ||
		id == "bad-defaults.no-timeout-fetch" ||
		id == "bad-defaults.any-type-assertion" ||
		id == "bad-defaults.non-null-assertion" ||
		id == "bad-defaults.eval-usage" ||
		id == "test-isolation.no-afterall-cleanup" ||
		id == "test-isolation.test-only-import" ||
		id == "mock-leakage.jest-mock-in-src" ||
		id == "mock-leakage.storybook-in-src" ||
		id == "bad-defaults.ts-ignore" ||
		id == "bad-defaults.eslint-disable-no-reason" ||
		id == "bad-defaults.no-strict-mode" ||
		id == "trace-gaps.no-unhandled-rejection" ||
		id == "trace-gaps.console-log-in-production" ||
		id == "config-smells.hardcoded-api-url" ||
		id == "config-smells.dotenv-no-example"
}

// isGoRuleID returns true if the rule ID belongs to a Go-only rule.
func isGoRuleID(id string) bool {
	goRules := []string{
		"silent-fallback.", "error-context.", "trace-gaps.", "bad-defaults.",
	}
	for _, prefix := range goRules {
		if strings.HasPrefix(id, prefix) && !isJSRuleID(id) && !isPyRuleID(id) {
			return true
		}
	}
	return false
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
