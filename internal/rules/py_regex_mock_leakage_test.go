package rules

import "testing"

// --- PyPytestFixtureInSrc ---

func TestPyPytestFixtureInSrc_Positive(t *testing.T) {
	checker := &PyPytestFixtureInSrc{}
	lines := []string{
		`import pytest`,
		``,
		`@pytest.fixture`,
		`def database():`,
		`    return connect_db()`,
		``,
		`@pytest.fixture(scope="session")`,
		`def app():`,
		`    return create_app()`,
	}
	findings := checker.CheckLines("src/helpers.py", lines)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	if findings[0].Rule != "mock-leakage.pytest-fixture-in-src" {
		t.Fatalf("expected rule mock-leakage.pytest-fixture-in-src, got %s", findings[0].Rule)
	}
	if findings[0].Line != 3 {
		t.Fatalf("expected line 3, got %d", findings[0].Line)
	}
	if findings[1].Line != 7 {
		t.Fatalf("expected line 7, got %d", findings[1].Line)
	}
}

func TestPyPytestFixtureInSrc_TestFileSkipped(t *testing.T) {
	checker := &PyPytestFixtureInSrc{}
	lines := []string{
		`@pytest.fixture`,
		`def database():`,
		`    return connect_db()`,
	}
	findings := checker.CheckLines("tests/conftest.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings in conftest, got %d", len(findings))
	}
}

func TestPyPytestFixtureInSrc_TestDirSkipped(t *testing.T) {
	checker := &PyPytestFixtureInSrc{}
	lines := []string{`@pytest.fixture`, `def db():`}
	findings := checker.CheckLines("test/fixtures.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings in test dir, got %d", len(findings))
	}
}

func TestPyPytestFixtureInSrc_CommentsSkipped(t *testing.T) {
	checker := &PyPytestFixtureInSrc{}
	lines := []string{`# @pytest.fixture`, `def real_function():`}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for commented fixture, got %d", len(findings))
	}
}

func TestPyPytestFixtureInSrc_NoFixtureClean(t *testing.T) {
	checker := &PyPytestFixtureInSrc{}
	lines := []string{`import os`, ``, `def get_data():`, `    return 42`}
	findings := checker.CheckLines("src/data.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
