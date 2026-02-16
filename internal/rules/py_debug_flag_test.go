package rules

import (
	"testing"
)

func TestPyDebugFlag(t *testing.T) {
	checker := &PyDebugFlag{}

	t.Run("detects if __debug__ with side effects", func(t *testing.T) {
		src := `if __debug__:
    print("debug mode")
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "mock-leakage.py-debug-flag" {
			t.Errorf("expected rule mock-leakage.py-debug-flag, got %s", findings[0].Rule)
		}
	})

	t.Run("detects if DEBUG with side effects", func(t *testing.T) {
		src := `if DEBUG:
    logging.info("debugging enabled")
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("detects if __debug__ with return", func(t *testing.T) {
		src := `def f():
    if __debug__:
        return "debug"
    return "release"
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("skips if __debug__ with pass only", func(t *testing.T) {
		src := `if __debug__:
    pass
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		src := `if __debug__:
    print("debug in test")
`
		findings := runPyCheckerOnSource(t, checker, "tests/test_app.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings in test file, got %d", len(findings))
		}
	})

	t.Run("no false positive on DEBUG_MODE", func(t *testing.T) {
		src := `if DEBUG_MODE:
    print("debug mode")
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings for DEBUG_MODE, got %d", len(findings))
		}
	})

	t.Run("no false positive on regular if", func(t *testing.T) {
		src := `if x > 0:
    print("positive")
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("multiple debug guards", func(t *testing.T) {
		src := `if __debug__:
    print("one")

if DEBUG:
    logging.error("two")
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})
}
