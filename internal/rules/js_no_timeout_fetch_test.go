package rules

import (
	"testing"
)

func TestJSNoTimeoutFetch(t *testing.T) {
	checker := &JSNoTimeoutFetch{}

	t.Run("triggers on bare fetch", func(t *testing.T) {
		src := `const data = await fetch("/api/users");`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "bad-defaults.no-timeout-fetch" {
			t.Errorf("wrong rule: %s", findings[0].Rule)
		}
	})

	t.Run("triggers on fetch with options but no signal", func(t *testing.T) {
		src := `
const response = await fetch("/api/data", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(data),
});
`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("clean with signal option", func(t *testing.T) {
		src := `
const controller = new AbortController();
const response = await fetch("/api/data", { signal: controller.signal });
`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean with AbortSignal.timeout", func(t *testing.T) {
		src := `const response = await fetch("/api/data", { signal: AbortSignal.timeout(5000) });`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean with spread options", func(t *testing.T) {
		src := `const response = await fetch("/api/data", { ...opts });`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean with shorthand signal", func(t *testing.T) {
		src := `const response = await fetch("/api/data", { signal });`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		src := `const data = await fetch("/api/users");`
		findings := runJSCheckerOnSource(t, checker, "app.test.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings in test file, got %d", len(findings))
		}
	})

	t.Run("triggers on multiple fetch calls", func(t *testing.T) {
		src := `
const a = await fetch("/api/a");
const b = await fetch("/api/b");
`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("clean with spread argument", func(t *testing.T) {
		src := `const data = await fetch("/api/data", ...args);`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings with spread arg, got %d", len(findings))
		}
	})
}
