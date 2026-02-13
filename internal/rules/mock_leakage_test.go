package rules

import "testing"

func TestTestFixtureRef_SkipsOwnRuleSource(t *testing.T) {
	checker := &TestFixtureRef{}
	lines := []string{
		`Description: "Reference to testdata/ or fixture path in production code",`,
	}

	findings := checker.CheckLines("internal/rules/mock_leakage.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestTestFixtureRef_SkipsRegexpCompileLine(t *testing.T) {
	checker := &TestFixtureRef{}
	lines := []string{
		`fixtureRefPattern := regexp.MustCompile("(?:testdata/|fixture[s]?/|test_data/)")`,
	}

	findings := checker.CheckLines("internal/app/main.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestTestFixtureRef_FindsFixturePath(t *testing.T) {
	checker := &TestFixtureRef{}
	lines := []string{
		`path := "testdata/input.json"`,
	}

	findings := checker.CheckLines("internal/app/main.go", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
