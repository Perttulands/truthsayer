package rules

import "testing"

func TestEmptyErrorCheck_ReturnsNil(t *testing.T) {
	checker := &EmptyErrorCheck{}
	src := `package p

func foo() error {
	err := doSomething()
	if err != nil {
		return nil
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

func TestEmptyErrorCheck_ReturnsError(t *testing.T) {
	checker := &EmptyErrorCheck{}
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
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when error is returned, got %d", len(findings))
	}
}

func TestEmptyErrorCheck_MultiReturnNil(t *testing.T) {
	checker := &EmptyErrorCheck{}
	src := `package p

func foo() (string, error) {
	result, err := doSomething()
	if err != nil {
		return nil, nil
	}
	return result, nil
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for multi-return nil, got %d", len(findings))
	}
}

func TestEmptyErrorCheck_WalkFuncExempt(t *testing.T) {
	checker := &EmptyErrorCheck{}
	src := `package p

import (
	"os"
	"path/filepath"
)

func foo() {
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		return nil
	})
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for error-only return func literal, got %d", len(findings))
	}
}

func TestEmptyErrorCheck_BodyNotJustReturn(t *testing.T) {
	checker := &EmptyErrorCheck{}
	src := `package p

import "log"

func foo() error {
	err := doSomething()
	if err != nil {
		log.Println(err)
		return nil
	}
	return nil
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when body has logging, got %d", len(findings))
	}
}
