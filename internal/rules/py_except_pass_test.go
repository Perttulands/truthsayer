package rules

import "testing"

func TestPyExceptPass_Positive(t *testing.T) {
	src := `
try:
    something()
except ValueError:
    pass
`
	findings := runPyCheckerOnSource(t, &PyExceptPass{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.py-except-pass" {
		t.Errorf("wrong rule: %s", findings[0].Rule)
	}
}

func TestPyExceptPass_BareExceptPass(t *testing.T) {
	src := `
try:
    something()
except:
    pass
`
	findings := runPyCheckerOnSource(t, &PyExceptPass{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestPyExceptPass_NegativeWithHandling(t *testing.T) {
	src := `
try:
    something()
except ValueError:
    log.error("failed")
`
	findings := runPyCheckerOnSource(t, &PyExceptPass{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPyExceptPass_NegativeWithMultipleStatements(t *testing.T) {
	src := `
try:
    something()
except ValueError:
    pass
    log.info("ignored")
`
	findings := runPyCheckerOnSource(t, &PyExceptPass{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for body with multiple statements, got %d", len(findings))
	}
}

func TestPyExceptPass_NegativeWithRaise(t *testing.T) {
	src := `
try:
    something()
except ValueError as e:
    raise RuntimeError("wrapped") from e
`
	findings := runPyCheckerOnSource(t, &PyExceptPass{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPyExceptPass_MultipleExcepts(t *testing.T) {
	src := `
try:
    something()
except ValueError:
    pass
except TypeError:
    handle()
`
	findings := runPyCheckerOnSource(t, &PyExceptPass{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
