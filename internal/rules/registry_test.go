package rules

import "testing"

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
	if !ids["bad-defaults.missing-pipefail"] {
		t.Error("missing bad-defaults.missing-pipefail rule")
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
