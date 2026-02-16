package rules

import "testing"

// --- PyTypeIgnoreBare ---

func TestPyTypeIgnoreBare_Positive(t *testing.T) {
	checker := &PyTypeIgnoreBare{}
	lines := []string{
		`x: int = "hello"  # type: ignore`,
		`y = some_func()  # type: ignore`,
		`z = other()  # type: ignore`,
	}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
	if findings[0].Rule != "bad-defaults.type-ignore-bare" {
		t.Fatalf("expected rule bad-defaults.type-ignore-bare, got %s", findings[0].Rule)
	}
}

func TestPyTypeIgnoreBare_SpecificCodeClean(t *testing.T) {
	checker := &PyTypeIgnoreBare{}
	lines := []string{
		`x: int = "hello"  # type: ignore[assignment]`,
		`y = some_func()  # type: ignore[no-any-return]`,
		`z = other()  # type: ignore[misc]`,
	}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for specific codes, got %d", len(findings))
	}
}

func TestPyTypeIgnoreBare_WithSpaces(t *testing.T) {
	checker := &PyTypeIgnoreBare{}
	lines := []string{`x = y  #  type:  ignore`}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestPyTypeIgnoreBare_InlineCommentClean(t *testing.T) {
	checker := &PyTypeIgnoreBare{}
	// This line has text after "type: ignore" so it's not bare
	lines := []string{`x = y  # type: ignore[override]  some reason`}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

// --- PyNoqaBare ---

func TestPyNoqaBare_Positive(t *testing.T) {
	checker := &PyNoqaBare{}
	lines := []string{
		`import os  # noqa`,
		`from sys import *  # noqa`,
		`x = 1; y = 2  # noqa`,
	}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
	if findings[0].Rule != "bad-defaults.noqa-bare" {
		t.Fatalf("expected rule bad-defaults.noqa-bare, got %s", findings[0].Rule)
	}
}

func TestPyNoqaBare_SpecificCodeClean(t *testing.T) {
	checker := &PyNoqaBare{}
	lines := []string{
		`import os  # noqa: F401`,
		`from sys import *  # noqa: F403`,
		`x = 1; y = 2  # noqa: E702`,
	}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for specific codes, got %d", len(findings))
	}
}

func TestPyNoqaBare_WithSpaces(t *testing.T) {
	checker := &PyNoqaBare{}
	lines := []string{`x = y  #  noqa`}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestPyNoqaBare_NoMatchClean(t *testing.T) {
	checker := &PyNoqaBare{}
	lines := []string{`x = 42`, `def func():`, `    return True`}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
