package rules

import "testing"

func TestNilOnError_NilNilReturn(t *testing.T) {
	checker := &NilOnError{}
	src := `package p
type User struct{}

func load() (*User, error) {
	_, err := fetch()
	if err != nil {
		return nil, nil
	}
	return &User{}, nil
}`

	findings := runASTCheckerOnSource(t, checker, "service.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Fatalf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestNilOnError_EmptyStringNilReturn(t *testing.T) {
	checker := &NilOnError{}
	src := `package p

func load() (string, error) {
	_, err := fetch()
	if err != nil {
		return "", nil
	}
	return "ok", nil
}`

	findings := runASTCheckerOnSource(t, checker, "service.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestNilOnError_ReturningErrIsAllowed(t *testing.T) {
	checker := &NilOnError{}
	src := `package p

type User struct{}

func load() (*User, error) {
	_, err := fetch()
	if err != nil {
		return nil, err
	}
	return &User{}, nil
}`

	findings := runASTCheckerOnSource(t, checker, "service.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
