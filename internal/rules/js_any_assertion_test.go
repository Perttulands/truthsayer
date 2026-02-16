package rules

import (
	"testing"
)

func TestJSAnyAssertion(t *testing.T) {
	checker := &JSAnyAssertion{}

	t.Run("triggers on as any", func(t *testing.T) {
		src := `const x = value as any;`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "bad-defaults.any-type-assertion" {
			t.Errorf("wrong rule: %s", findings[0].Rule)
		}
	})

	t.Run("triggers on double assertion with any", func(t *testing.T) {
		src := `const x = value as unknown as any;`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("clean on specific type assertion", func(t *testing.T) {
		src := `const x = value as string;`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on as unknown", func(t *testing.T) {
		src := `const x = value as unknown;`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on custom type", func(t *testing.T) {
		src := `const x = value as SomeType;`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("skips JS files", func(t *testing.T) {
		// as any is TS-only syntax, but even if parsed, skip .js files
		src := `const x = value;`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for .js file, got %d", len(findings))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		src := `const x = value as any;`
		findings := runTSCheckerOnSource(t, checker, "app.test.ts", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings in test file, got %d", len(findings))
		}
	})

	t.Run("triggers on multiple", func(t *testing.T) {
		src := `
const x = a as any;
const y = b as any;
const z = c as string;
`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("tsx file", func(t *testing.T) {
		src := `const x = value as any;`
		findings := runTSCheckerOnSource(t, checker, "component.tsx", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding for .tsx, got %d", len(findings))
		}
	})
}
