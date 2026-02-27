package rules

import "testing"

func TestGenericError_ErrorsNew(t *testing.T) {
	checker := &GenericError{}
	src := `package p

import "errors"

func foo() error {
	return errors.New("failed")
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

func TestGenericError_FmtErrorf(t *testing.T) {
	checker := &GenericError{}
	src := `package p

import "fmt"

func foo() error {
	return fmt.Errorf("unknown error")
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for fmt.Errorf generic message, got %d", len(findings))
	}
}

func TestGenericError_SpecificMessage(t *testing.T) {
	checker := &GenericError{}
	src := `package p

import "errors"

func foo() error {
	return errors.New("failed to connect to database at 10.0.0.1:5432")
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for specific message, got %d", len(findings))
	}
}

func TestGenericError_NonStringArg(t *testing.T) {
	checker := &GenericError{}
	src := `package p

import "fmt"

func foo(msg string) error {
	return fmt.Errorf(msg)
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for variable arg, got %d", len(findings))
	}
}
