package rules

import "testing"

func TestIgnoredError_BlankIdentifier(t *testing.T) {
	checker := &IgnoredError{}
	src := `package p

func foo() {
	result, _ := doSomething()
	_ = result
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestIgnoredError_ErrorHandled(t *testing.T) {
	checker := &IgnoredError{}
	src := `package p

func foo() error {
	result, err := doSomething()
	if err != nil {
		return err
	}
	_ = result
	return nil
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when error is handled, got %d", len(findings))
	}
}

func TestIgnoredError_SkipsTestFiles(t *testing.T) {
	checker := &IgnoredError{}
	src := `package p

func TestFoo() {
	result, _ := doSomething()
	_ = result
}
`
	findings := runASTCheckerOnSource(t, checker, "handler_test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestIgnoredError_SingleAssign(t *testing.T) {
	checker := &IgnoredError{}
	src := `package p

func foo() {
	_ = doSomething()
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for single-value assignment, got %d", len(findings))
	}
}
