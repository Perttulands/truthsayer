package rules

import "testing"

func TestPyGetattrSilentDefault_None(t *testing.T) {
	src := `
x = getattr(obj, 'name', None)
`
	findings := runPyCheckerOnSource(t, &PyGetattrSilentDefault{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.py-getattr-silent-default" {
		t.Errorf("wrong rule: %s", findings[0].Rule)
	}
}

func TestPyGetattrSilentDefault_NegativeExplicitDefault(t *testing.T) {
	src := `
x = getattr(obj, 'name', 'default_value')
`
	findings := runPyCheckerOnSource(t, &PyGetattrSilentDefault{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPyGetattrSilentDefault_NegativeTwoArgs(t *testing.T) {
	src := `
x = getattr(obj, 'name')
`
	findings := runPyCheckerOnSource(t, &PyGetattrSilentDefault{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for 2-arg getattr (raises AttributeError), got %d", len(findings))
	}
}

func TestPyGetattrSilentDefault_NegativeFalseDefault(t *testing.T) {
	src := `
x = getattr(obj, 'enabled', False)
`
	findings := runPyCheckerOnSource(t, &PyGetattrSilentDefault{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for False default, got %d", len(findings))
	}
}

func TestPyGetattrSilentDefault_NegativeEmptyListDefault(t *testing.T) {
	src := `
x = getattr(obj, 'items', [])
`
	findings := runPyCheckerOnSource(t, &PyGetattrSilentDefault{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for [] default, got %d", len(findings))
	}
}

func TestPyGetattrSilentDefault_Multiple(t *testing.T) {
	src := `
x = getattr(obj, 'name', None)
y = getattr(obj, 'age', None)
z = getattr(obj, 'email', 'unknown')
`
	findings := runPyCheckerOnSource(t, &PyGetattrSilentDefault{}, "app.py", src)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}
