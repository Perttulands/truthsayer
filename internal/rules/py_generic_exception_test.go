package rules

import (
	"testing"
)

func TestPyGenericException(t *testing.T) {
	checker := &PyGenericException{}

	t.Run("detects raise Exception", func(t *testing.T) {
		src := `raise Exception("something failed")`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "error-context.py-generic-exception" {
			t.Errorf("unexpected rule: %s", findings[0].Rule)
		}
	})

	t.Run("detects raise BaseException", func(t *testing.T) {
		src := `raise BaseException("critical")`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("detects multiple generic exceptions", func(t *testing.T) {
		src := `
raise Exception("a")
raise Exception("b")
raise BaseException("c")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 3 {
			t.Fatalf("expected 3 findings, got %d", len(findings))
		}
	})

	t.Run("clean on specific exceptions", func(t *testing.T) {
		src := `
raise ValueError("invalid")
raise TypeError("wrong type")
raise RuntimeError("runtime error")
raise KeyError("missing key")
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

	t.Run("clean on custom exception", func(t *testing.T) {
		src := `
class CustomError(Exception):
    pass

raise CustomError("custom")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("detects Exception in function", func(t *testing.T) {
		src := `
def process():
    if not valid:
        raise Exception("invalid data")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})
}
