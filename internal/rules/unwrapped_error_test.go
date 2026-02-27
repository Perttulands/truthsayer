package rules

import "testing"

func TestUnwrappedError_BareReturn(t *testing.T) {
	checker := &UnwrappedError{}
	src := `package p

func foo() error {
	err := doSomething()
	if err != nil {
		return err
	}
	return nil
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

func TestUnwrappedError_WrappedError(t *testing.T) {
	checker := &UnwrappedError{}
	src := `package p

import "fmt"

func foo() error {
	err := doSomething()
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}
	return nil
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for wrapped error, got %d", len(findings))
	}
}

func TestUnwrappedError_MultiReturn(t *testing.T) {
	checker := &UnwrappedError{}
	src := `package p

func foo() (string, error) {
	result, err := doSomething()
	if err != nil {
		return "", err
	}
	return result, nil
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for bare err in multi-return, got %d", len(findings))
	}
}

func TestUnwrappedError_NoErrCheck(t *testing.T) {
	checker := &UnwrappedError{}
	src := `package p

func foo() error {
	return nil
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when no err check, got %d", len(findings))
	}
}
