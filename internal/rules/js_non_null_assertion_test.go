package rules

import (
	"testing"
)

func TestJSNonNullAssertion(t *testing.T) {
	checker := &JSNonNullAssertion{}

	t.Run("triggers on non-null assertion", func(t *testing.T) {
		src := `const x = obj!;`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "bad-defaults.non-null-assertion" {
			t.Errorf("wrong rule: %s", findings[0].Rule)
		}
	})

	t.Run("triggers on property access after assertion", func(t *testing.T) {
		src := `const x = obj!.prop;`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("triggers on subscript after assertion", func(t *testing.T) {
		src := `const x = arr![0];`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("clean on regular property access", func(t *testing.T) {
		src := `const x = obj.prop;`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on optional chaining", func(t *testing.T) {
		src := `const x = obj?.prop;`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("skips JS files", func(t *testing.T) {
		src := `const x = obj;`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for .js file, got %d", len(findings))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		src := `const x = obj!;`
		findings := runTSCheckerOnSource(t, checker, "app.test.ts", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings in test file, got %d", len(findings))
		}
	})

	t.Run("triggers on multiple assertions", func(t *testing.T) {
		src := `
const x = a!;
const y = b!.c;
const z = d;
`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("tsx file", func(t *testing.T) {
		src := `const x = ref!;`
		findings := runTSCheckerOnSource(t, checker, "component.tsx", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding for .tsx, got %d", len(findings))
		}
	})
}
