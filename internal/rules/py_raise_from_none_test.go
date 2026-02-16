package rules

import (
	"testing"
)

func TestPyRaiseFromNone(t *testing.T) {
	checker := &PyRaiseFromNone{}

	t.Run("detects raise from None", func(t *testing.T) {
		src := `
try:
    connect()
except ConnectionError as e:
    raise RuntimeError("failed") from None
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "error-context.py-raise-from-none" {
			t.Errorf("unexpected rule: %s", findings[0].Rule)
		}
		if findings[0].Line != 5 {
			t.Errorf("expected line 5, got %d", findings[0].Line)
		}
	})

	t.Run("detects multiple raise from None", func(t *testing.T) {
		src := `
try:
    x()
except ValueError as e:
    raise TypeError("a") from None
try:
    y()
except KeyError as e:
    raise RuntimeError("b") from None
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("clean on raise from e", func(t *testing.T) {
		src := `
try:
    connect()
except ConnectionError as e:
    raise RuntimeError("failed") from e
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

	t.Run("clean on raise without from", func(t *testing.T) {
		src := `
try:
    x()
except ValueError:
    raise TypeError("oops")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})
}
