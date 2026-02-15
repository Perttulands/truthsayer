package rules

import "testing"

func TestErrorStringCompare_DetectsErrErrorComparison(t *testing.T) {
	src := `package main

func handle(err error) {
	if err.Error() == "not found" {
		return
	}
}`
	findings := runASTCheckerOnSource(t, &ErrorStringCompare{}, "main.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "error-context.error-string-compare" {
		t.Fatalf("expected rule error-context.error-string-compare, got %s", findings[0].Rule)
	}
}

func TestErrorStringCompare_DetectsNotEqualComparison(t *testing.T) {
	src := `package main

func handle(err error) {
	if err.Error() != "ok" {
		return
	}
}`
	findings := runASTCheckerOnSource(t, &ErrorStringCompare{}, "main.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestErrorStringCompare_SkipsNonErrorMethods(t *testing.T) {
	src := `package main

func handle(s string) {
	if s == "not found" {
		return
	}
}`
	findings := runASTCheckerOnSource(t, &ErrorStringCompare{}, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestErrorStringCompare_SkipsTestFiles(t *testing.T) {
	src := `package main

func TestHandle(err error) {
	if err.Error() == "expected error" {
		return
	}
}`
	findings := runASTCheckerOnSource(t, &ErrorStringCompare{}, "main_test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
