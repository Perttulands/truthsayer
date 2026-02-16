package rules

import (
	"testing"
)

func TestPyBareRaiseDifferent(t *testing.T) {
	checker := &PyBareRaiseDifferent{}

	t.Run("detects raise different without from", func(t *testing.T) {
		src := `
try:
    connect()
except ConnectionError:
    raise RuntimeError("failed")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "error-context.py-bare-raise-different" {
			t.Errorf("unexpected rule: %s", findings[0].Rule)
		}
		if findings[0].Line != 5 {
			t.Errorf("expected line 5, got %d", findings[0].Line)
		}
	})

	t.Run("detects multiple bare raise different", func(t *testing.T) {
		src := `
try:
    x()
except ValueError:
    raise TypeError("a")
try:
    y()
except KeyError:
    raise RuntimeError("b")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("clean on raise with from", func(t *testing.T) {
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
except ValueError:
    raise
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on raise identifier (re-raise variable)", func(t *testing.T) {
		src := `
try:
    x()
except ValueError as e:
    raise e
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("detects raise in nested except", func(t *testing.T) {
		src := `
try:
    x()
except ValueError:
    try:
        y()
    except KeyError:
        raise TypeError("nested")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})
}
