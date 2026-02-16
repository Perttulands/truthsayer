package rules

import (
	"testing"
)

func TestJSEvalUsage(t *testing.T) {
	checker := &JSEvalUsage{}

	t.Run("triggers on eval", func(t *testing.T) {
		src := `const result = eval("1 + 2");`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "bad-defaults.eval-usage" {
			t.Errorf("wrong rule: %s", findings[0].Rule)
		}
		if findings[0].Message != "eval() executes arbitrary code — security and debuggability risk" {
			t.Errorf("wrong message: %s", findings[0].Message)
		}
	})

	t.Run("triggers on new Function", func(t *testing.T) {
		src := `const fn = new Function("a", "b", "return a + b");`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Message != "new Function() executes arbitrary code — security and debuggability risk" {
			t.Errorf("wrong message: %s", findings[0].Message)
		}
	})

	t.Run("triggers on eval without assignment", func(t *testing.T) {
		src := `eval(code);`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("clean on regular function call", func(t *testing.T) {
		src := `const result = calculate(1, 2);`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on new RegExp", func(t *testing.T) {
		src := `const re = new RegExp("test");`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("does not trigger on method named eval", func(t *testing.T) {
		src := `const result = obj.eval("expr");`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for obj.eval, got %d", len(findings))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		src := `eval("test code");`
		findings := runJSCheckerOnSource(t, checker, "app.test.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings in test file, got %d", len(findings))
		}
	})

	t.Run("triggers on both eval and new Function", func(t *testing.T) {
		src := `
eval("code");
const fn = new Function("return 42");
`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("works with TypeScript", func(t *testing.T) {
		src := `eval("code");`
		findings := runTSCheckerOnSource(t, checker, "app.ts", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding for TS file, got %d", len(findings))
		}
	})
}
