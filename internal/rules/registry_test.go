package rules

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// mockJSChecker is a test double for JSASTChecker.
type mockJSChecker struct {
	id string
}

func (m *mockJSChecker) Meta() Rule {
	return Rule{
		ID:       m.id,
		Category: "test",
		Name:     "mock JS rule",
		Severity: finding.SeverityWarning,
		FileTypes: []string{".js", ".ts"},
		ScanType: ScanTypeAST,
	}
}

func (m *mockJSChecker) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	return nil
}

// mockPyChecker is a test double for PyASTChecker.
type mockPyChecker struct {
	id string
}

func (m *mockPyChecker) Meta() Rule {
	return Rule{
		ID:       m.id,
		Category: "test",
		Name:     "mock Python rule",
		Severity: finding.SeverityError,
		FileTypes: []string{".py"},
		ScanType: ScanTypeAST,
	}
}

func (m *mockPyChecker) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	return nil
}

func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry()
	all := reg.AllRules()
	if len(all) < 2 {
		t.Fatalf("expected at least 2 default rules, got %d", len(all))
	}

	ids := make(map[string]bool)
	for _, r := range all {
		ids[r.ID] = true
	}
	if !ids["silent-fallback.empty-error-check"] {
		t.Error("missing silent-fallback.empty-error-check rule")
	}
	if !ids["silent-fallback.bare-return-on-error"] {
		t.Error("missing silent-fallback.bare-return-on-error rule")
	}
	if !ids["error-context.http-200-on-error"] {
		t.Error("missing error-context.http-200-on-error rule")
	}
	if !ids["error-context.nil-on-error"] {
		t.Error("missing error-context.nil-on-error rule")
	}
	if !ids["trace-gaps.no-request-id"] {
		t.Error("missing trace-gaps.no-request-id rule")
	}
	if !ids["bad-defaults.missing-pipefail"] {
		t.Error("missing bad-defaults.missing-pipefail rule")
	}
	if !ids["trace-gaps.no-stderr-capture"] {
		t.Error("missing trace-gaps.no-stderr-capture rule")
	}
	if !ids["bad-defaults.magic-number"] {
		t.Error("missing bad-defaults.magic-number rule")
	}
	if !ids["config-smells.missing-gitignore"] {
		t.Error("missing config-smells.missing-gitignore rule")
	}
	if !ids["mock-leakage.debug-guard"] {
		t.Error("missing mock-leakage.debug-guard rule")
	}
	if !ids["test-isolation.test-leaked-server"] {
		t.Error("missing test-isolation.test-leaked-server rule")
	}
	if !ids["test-isolation.test-leaked-sse"] {
		t.Error("missing test-isolation.test-leaked-sse rule")
	}
	if !ids["test-isolation.test-missing-cleanup"] {
		t.Error("missing test-isolation.test-missing-cleanup rule")
	}
}

func TestDisable(t *testing.T) {
	reg := DefaultRegistry()
	reg.Disable("silent-fallback.empty-error-check")
	checkers := reg.ASTCheckers()
	for _, c := range checkers {
		if c.Meta().ID == "silent-fallback.empty-error-check" {
			t.Error("disabled rule still returned by ASTCheckers")
		}
	}
}

func TestSetSeverity_Known(t *testing.T) {
	reg := DefaultRegistry()
	ok := reg.SetSeverity("silent-fallback.empty-error-check", "warning")
	if !ok {
		t.Fatal("SetSeverity returned false for known rule")
	}
}

func TestSetSeverity_Unknown(t *testing.T) {
	reg := DefaultRegistry()
	ok := reg.SetSeverity("nonexistent.rule", "warning")
	if ok {
		t.Fatal("SetSeverity returned true for unknown rule")
	}
}

func TestApplyOverrides(t *testing.T) {
	reg := DefaultRegistry()
	reg.SetSeverity("silent-fallback.empty-error-check", "info")

	findings := []finding.Finding{
		{Rule: "silent-fallback.empty-error-check", Severity: finding.SeverityError},
		{Rule: "bad-defaults.missing-pipefail", Severity: finding.SeverityError},
	}
	reg.ApplyOverrides(findings)

	if findings[0].Severity != finding.SeverityInfo {
		t.Errorf("expected info severity, got %s", findings[0].Severity)
	}
	if findings[1].Severity != finding.SeverityError {
		t.Errorf("expected error severity unchanged, got %s", findings[1].Severity)
	}
}

func TestApplyOverrides_NoOverrides(t *testing.T) {
	reg := DefaultRegistry()
	findings := []finding.Finding{
		{Rule: "silent-fallback.empty-error-check", Severity: finding.SeverityError},
	}
	reg.ApplyOverrides(findings)
	if findings[0].Severity != finding.SeverityError {
		t.Errorf("severity should be unchanged, got %s", findings[0].Severity)
	}
}

func TestEnabledRules_All(t *testing.T) {
	reg := DefaultRegistry()
	enabled := reg.EnabledRules()
	all := reg.AllRules()
	if len(enabled) != len(all) {
		t.Errorf("expected %d enabled rules (all), got %d", len(all), len(enabled))
	}
}

func TestEnabledRules_WithDisabled(t *testing.T) {
	reg := DefaultRegistry()
	all := reg.AllRules()
	reg.Disable("silent-fallback.empty-error-check")
	enabled := reg.EnabledRules()
	if len(enabled) != len(all)-1 {
		t.Fatalf("expected %d enabled rules, got %d", len(all)-1, len(enabled))
	}
	for _, r := range enabled {
		if r.ID == "silent-fallback.empty-error-check" {
			t.Error("disabled rule still in enabled list")
		}
	}
}

func TestEnabledRules_AllDisabled(t *testing.T) {
	reg := DefaultRegistry()
	all := reg.AllRules()
	for _, r := range all {
		reg.Disable(r.ID)
	}
	enabled := reg.EnabledRules()
	if len(enabled) != 0 {
		t.Errorf("expected 0 enabled rules, got %d", len(enabled))
	}
}

func TestRegisterJSAST(t *testing.T) {
	reg := NewRegistry()
	js1 := &mockJSChecker{id: "test.js-rule-1"}
	js2 := &mockJSChecker{id: "test.js-rule-2"}
	reg.RegisterJSAST(js1)
	reg.RegisterJSAST(js2)

	checkers := reg.JSASTCheckers()
	if len(checkers) != 2 {
		t.Fatalf("expected 2 JS AST checkers, got %d", len(checkers))
	}
	if checkers[0].Meta().ID != "test.js-rule-1" {
		t.Errorf("expected test.js-rule-1, got %s", checkers[0].Meta().ID)
	}
	if checkers[1].Meta().ID != "test.js-rule-2" {
		t.Errorf("expected test.js-rule-2, got %s", checkers[1].Meta().ID)
	}
}

func TestRegisterPyAST(t *testing.T) {
	reg := NewRegistry()
	py1 := &mockPyChecker{id: "test.py-rule-1"}
	py2 := &mockPyChecker{id: "test.py-rule-2"}
	reg.RegisterPyAST(py1)
	reg.RegisterPyAST(py2)

	checkers := reg.PyASTCheckers()
	if len(checkers) != 2 {
		t.Fatalf("expected 2 Python AST checkers, got %d", len(checkers))
	}
	if checkers[0].Meta().ID != "test.py-rule-1" {
		t.Errorf("expected test.py-rule-1, got %s", checkers[0].Meta().ID)
	}
	if checkers[1].Meta().ID != "test.py-rule-2" {
		t.Errorf("expected test.py-rule-2, got %s", checkers[1].Meta().ID)
	}
}

func TestJSASTCheckers_Disabled(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterJSAST(&mockJSChecker{id: "test.js-enabled"})
	reg.RegisterJSAST(&mockJSChecker{id: "test.js-disabled"})
	reg.Disable("test.js-disabled")

	checkers := reg.JSASTCheckers()
	if len(checkers) != 1 {
		t.Fatalf("expected 1 enabled JS checker, got %d", len(checkers))
	}
	if checkers[0].Meta().ID != "test.js-enabled" {
		t.Errorf("expected test.js-enabled, got %s", checkers[0].Meta().ID)
	}
}

func TestPyASTCheckers_Disabled(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterPyAST(&mockPyChecker{id: "test.py-enabled"})
	reg.RegisterPyAST(&mockPyChecker{id: "test.py-disabled"})
	reg.Disable("test.py-disabled")

	checkers := reg.PyASTCheckers()
	if len(checkers) != 1 {
		t.Fatalf("expected 1 enabled Python checker, got %d", len(checkers))
	}
	if checkers[0].Meta().ID != "test.py-enabled" {
		t.Errorf("expected test.py-enabled, got %s", checkers[0].Meta().ID)
	}
}

func TestAllRules_IncludesJSAndPy(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterJSAST(&mockJSChecker{id: "test.js-rule"})
	reg.RegisterPyAST(&mockPyChecker{id: "test.py-rule"})

	all := reg.AllRules()
	if len(all) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(all))
	}
	ids := map[string]bool{}
	for _, r := range all {
		ids[r.ID] = true
	}
	if !ids["test.js-rule"] {
		t.Error("missing JS rule in AllRules")
	}
	if !ids["test.py-rule"] {
		t.Error("missing Python rule in AllRules")
	}
}

func TestEnabledRules_IncludesJSAndPy(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterJSAST(&mockJSChecker{id: "test.js-rule"})
	reg.RegisterPyAST(&mockPyChecker{id: "test.py-rule"})

	enabled := reg.EnabledRules()
	if len(enabled) != 2 {
		t.Fatalf("expected 2 enabled rules, got %d", len(enabled))
	}

	// Disable one, check it's excluded
	reg.Disable("test.py-rule")
	enabled = reg.EnabledRules()
	if len(enabled) != 1 {
		t.Fatalf("expected 1 enabled rule after disable, got %d", len(enabled))
	}
	if enabled[0].ID != "test.js-rule" {
		t.Errorf("expected test.js-rule, got %s", enabled[0].ID)
	}
}

func TestSetSeverity_JSAndPyRules(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterJSAST(&mockJSChecker{id: "test.js-rule"})
	reg.RegisterPyAST(&mockPyChecker{id: "test.py-rule"})

	if !reg.SetSeverity("test.js-rule", "info") {
		t.Error("SetSeverity returned false for JS rule")
	}
	if !reg.SetSeverity("test.py-rule", "warning") {
		t.Error("SetSeverity returned false for Python rule")
	}

	enabled := reg.EnabledRules()
	for _, r := range enabled {
		switch r.ID {
		case "test.js-rule":
			if r.Severity != finding.SeverityInfo {
				t.Errorf("JS rule severity: expected info, got %s", r.Severity)
			}
		case "test.py-rule":
			if r.Severity != finding.SeverityWarning {
				t.Errorf("Python rule severity: expected warning, got %s", r.Severity)
			}
		}
	}
}
