package rules

import "testing"

func TestBareReturnOnError_BareReturn(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func load() (data string, err error) {
	if err != nil {
		return
	}
	return "ok", nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Fatalf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestBareReturnOnError_ExplicitZeroValues(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func load() (data string, err error) {
	if err != nil {
		return "", nil
	}
	return "ok", nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestBareReturnOnError_NamedReturnRequired(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func load() (string, error) {
	if err != nil {
		return "", nil
	}
	return "ok", nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
