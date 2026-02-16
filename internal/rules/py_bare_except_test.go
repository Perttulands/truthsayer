package rules

import "testing"

func TestPyBareExcept_Positive(t *testing.T) {
	src := `
try:
    something()
except:
    handle()
`
	findings := runPyCheckerOnSource(t, &PyBareExcept{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.py-bare-except" {
		t.Errorf("wrong rule: %s", findings[0].Rule)
	}
}

func TestPyBareExcept_MultipleBare(t *testing.T) {
	src := `
try:
    a()
except:
    pass

try:
    b()
except:
    log(e)
`
	findings := runPyCheckerOnSource(t, &PyBareExcept{}, "app.py", src)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestPyBareExcept_NegativeWithType(t *testing.T) {
	src := `
try:
    something()
except ValueError:
    handle()
`
	findings := runPyCheckerOnSource(t, &PyBareExcept{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPyBareExcept_NegativeWithAsPattern(t *testing.T) {
	src := `
try:
    something()
except Exception as e:
    handle(e)
`
	findings := runPyCheckerOnSource(t, &PyBareExcept{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPyBareExcept_NegativeTupleException(t *testing.T) {
	src := `
try:
    something()
except (ValueError, TypeError):
    handle()
`
	findings := runPyCheckerOnSource(t, &PyBareExcept{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPyBareExcept_MixedBareAndTyped(t *testing.T) {
	src := `
try:
    a()
except ValueError:
    handle()
except:
    fallback()
`
	findings := runPyCheckerOnSource(t, &PyBareExcept{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
