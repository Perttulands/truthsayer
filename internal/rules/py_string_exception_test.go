package rules

import (
	"testing"
)

func TestPyStringException(t *testing.T) {
	checker := &PyStringException{}

	t.Run("detects raise string", func(t *testing.T) {
		src := `raise "something went wrong"`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "error-context.py-string-exception" {
			t.Errorf("unexpected rule: %s", findings[0].Rule)
		}
	})

	t.Run("detects multiple string exceptions", func(t *testing.T) {
		src := `
raise "error one"
raise "error two"
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("clean on proper exception", func(t *testing.T) {
		src := `
raise Exception("something went wrong")
raise ValueError("error occurred")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on bare raise", func(t *testing.T) {
		src := `
try:
    x()
except Exception:
    raise
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})
}
