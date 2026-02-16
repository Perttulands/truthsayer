package engine

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/perttulands/truthsayer/internal/rules"
)

func testdataSecurityDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "security")
}

var securityRuleFixtures = []ruleFixture{
	{"code-quality.unused-variable", "unused_variable.js", "unused_variable_negative.js"},
	{"code-quality.unreachable-code", "unreachable_code.py", "unreachable_code_negative.py"},
	{"code-quality.error-swallowing", "error_swallowing.js", "error_swallowing_negative.js"},
	{"config-smells.hardcoded-credentials", "hardcoded_credentials.ts", "hardcoded_credentials_negative.ts"},
	{"security.sql-injection", "sql_injection.py", "sql_injection_negative.py"},
	{"security.command-injection", "command_injection.go", "command_injection_negative.go"},
}

func TestSecurityIntegration_PositiveFixtures(t *testing.T) {
	secDir := testdataSecurityDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	for _, rf := range securityRuleFixtures {
		t.Run(rf.ruleID, func(t *testing.T) {
			path := filepath.Join(secDir, rf.positiveFixture)
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
				t.Fatalf("expected finding %q in %s", rf.ruleID, rf.positiveFixture)
			}
		})
	}
}

func TestSecurityIntegration_NegativeFixtures(t *testing.T) {
	secDir := testdataSecurityDir(t)
	reg := rules.DefaultRegistry()
	eng := New(reg)

	for _, rf := range securityRuleFixtures {
		t.Run(rf.ruleID, func(t *testing.T) {
			path := filepath.Join(secDir, rf.negativeFixture)
			result, err := eng.ScanFile(path)
			if err != nil {
				t.Fatalf("ScanFile(%s): %v", rf.negativeFixture, err)
			}

			for _, f := range result.Findings {
				if f.Rule == rf.ruleID {
					t.Fatalf("unexpected %q finding in negative fixture %s at line %d", rf.ruleID, rf.negativeFixture, f.Line)
				}
			}
		})
	}
}
