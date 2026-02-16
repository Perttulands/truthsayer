package rules

import (
	"testing"
)

func TestPyMutableDefault(t *testing.T) {
	checker := &PyMutableDefault{}

	t.Run("triggers on list default", func(t *testing.T) {
		src := `def append_to(items=[]):
    items.append(1)
    return items
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "bad-defaults.py-mutable-default-arg" {
			t.Errorf("expected rule bad-defaults.py-mutable-default-arg, got %s", findings[0].Rule)
		}
	})

	t.Run("triggers on dict default", func(t *testing.T) {
		src := `def make_config(options={}):
    return options
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("triggers on set default", func(t *testing.T) {
		src := `def collect(values=set()):
    values.add(42)
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("triggers on method in class", func(t *testing.T) {
		src := `class MyClass:
    def method(self, data=[1, 2, 3]):
        return data
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("multiple mutable defaults", func(t *testing.T) {
		src := `def f(a=[], b={}, c=set()):
    pass
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 3 {
			t.Fatalf("expected 3 findings, got %d", len(findings))
		}
	})

	t.Run("clean on None default", func(t *testing.T) {
		src := `def append_to(items=None):
    if items is None:
        items = []
    return items
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on immutable defaults", func(t *testing.T) {
		src := `def greet(name="world"):
    return f"Hello, {name}!"

def count(start=0, end=10):
    return range(start, end)

def toggle(flag=True):
    return not flag

def choose(value=(1, 2, 3)):
    return value
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})
}
