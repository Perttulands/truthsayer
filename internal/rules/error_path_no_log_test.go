package rules

import "testing"

func TestErrorPathNoLog_AllowsStderrOutput(t *testing.T) {
	checker := &ErrorPathNoLog{}
	src := `package main
import (
	"fmt"
	"os"
)
func f(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	return 0
}`

	findings := runASTCheckerOnSource(t, checker, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestErrorPathNoLog_FindsUnhandledErrorPath(t *testing.T) {
	checker := &ErrorPathNoLog{}
	src := `package main
func f(err error) error {
	if err != nil {
		_ = "cleanup"
		return err
	}
	return nil
}`

	findings := runASTCheckerOnSource(t, checker, "main.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
