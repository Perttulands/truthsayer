package rules

import "testing"

func TestPyExceptBroad_Exception(t *testing.T) {
	src := `
try:
    something()
except Exception:
    handle()
`
	findings := runPyCheckerOnSource(t, &PyExceptBroad{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.py-except-broad" {
		t.Errorf("wrong rule: %s", findings[0].Rule)
	}
}

func TestPyExceptBroad_BaseException(t *testing.T) {
	src := `
try:
    something()
except BaseException:
    handle()
`
	findings := runPyCheckerOnSource(t, &PyExceptBroad{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestPyExceptBroad_ExceptionAsE(t *testing.T) {
	src := `
try:
    something()
except Exception as e:
    handle(e)
`
	findings := runPyCheckerOnSource(t, &PyExceptBroad{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestPyExceptBroad_NegativeSpecific(t *testing.T) {
	src := `
try:
    something()
except ValueError:
    handle()
`
	findings := runPyCheckerOnSource(t, &PyExceptBroad{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPyExceptBroad_NegativeTuple(t *testing.T) {
	src := `
try:
    something()
except (ValueError, TypeError):
    handle()
`
	findings := runPyCheckerOnSource(t, &PyExceptBroad{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPyExceptBroad_NegativeBare(t *testing.T) {
	src := `
try:
    something()
except:
    handle()
`
	findings := runPyCheckerOnSource(t, &PyExceptBroad{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for bare except (different rule), got %d", len(findings))
	}
}

func TestPyExceptBroad_MultipleExcepts(t *testing.T) {
	src := `
try:
    something()
except ValueError:
    handle_value()
except Exception:
    handle_generic()
`
	findings := runPyCheckerOnSource(t, &PyExceptBroad{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestPyExceptBroad_NegativeSpecificAsE(t *testing.T) {
	src := `
try:
    something()
except KeyError as e:
    handle(e)
`
	findings := runPyCheckerOnSource(t, &PyExceptBroad{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
