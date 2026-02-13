package rules

import (
	"testing"

	"github.com/perttulands/truthsayer/internal/finding"
)

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
