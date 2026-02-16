package rules

import (
	"testing"
)

func TestPyLogAndRaise(t *testing.T) {
	checker := &PyLogAndRaise{}

	t.Run("detects logging.error and raise", func(t *testing.T) {
		src := `
try:
    connect()
except ConnectionError as e:
    logging.error(e)
    raise
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "error-context.py-log-and-raise" {
			t.Errorf("unexpected rule: %s", findings[0].Rule)
		}
		if findings[0].Line != 5 {
			t.Errorf("expected line 5, got %d", findings[0].Line)
		}
	})

	t.Run("detects logger.exception and raise", func(t *testing.T) {
		src := `
try:
    parse()
except ValueError as e:
    logger.exception(e)
    raise
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("detects log.critical and raise", func(t *testing.T) {
		src := `
try:
    load()
except Exception as e:
    log.critical(e)
    raise RuntimeError("failed")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("detects logging.warning and raise", func(t *testing.T) {
		src := `
try:
    x()
except Exception as e:
    logging.warning(e)
    raise
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("clean on only logging no raise", func(t *testing.T) {
		src := `
try:
    connect()
except ConnectionError as e:
    logging.error(e)
    return None
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on only raise no logging", func(t *testing.T) {
		src := `
try:
    parse()
except ValueError as e:
    raise
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on logging.info and raise", func(t *testing.T) {
		src := `
try:
    load()
except Exception as e:
    logging.info("retrying")
    raise
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on print and raise", func(t *testing.T) {
		src := `
try:
    x()
except Exception as e:
    print(e)
    raise
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("detects multiple except blocks", func(t *testing.T) {
		src := `
try:
    a()
except ValueError as e:
    logging.error(e)
    raise
try:
    b()
except KeyError as e:
    logger.exception(e)
    raise
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})
}
