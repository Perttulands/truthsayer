package rules

import (
	"testing"
)

func TestJSConsoleErrorNoThrow_Positive(t *testing.T) {
	src := `
try {
  doSomething();
} catch (err) {
  console.error(err);
}
`
	checker := &JSConsoleErrorNoThrow{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "error-context.js-console-error-no-throw" {
		t.Errorf("unexpected rule ID: %s", findings[0].Rule)
	}
}

func TestJSConsoleErrorNoThrow_Positive_WithReturn(t *testing.T) {
	src := `
try {
  doWork();
} catch (e) {
  console.error("Error:", e.message);
  return null;
}
`
	checker := &JSConsoleErrorNoThrow{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSConsoleErrorNoThrow_Negative_WithThrow(t *testing.T) {
	src := `
try {
  doSomething();
} catch (err) {
  console.error(err);
  throw err;
}
`
	checker := &JSConsoleErrorNoThrow{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSConsoleErrorNoThrow_Negative_WithWrappedThrow(t *testing.T) {
	src := `
try {
  doWork();
} catch (e) {
  console.error("Error:", e.message);
  throw new Error("failed", { cause: e });
}
`
	checker := &JSConsoleErrorNoThrow{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSConsoleErrorNoThrow_Negative_NoConsoleError(t *testing.T) {
	src := `
try {
  parse();
} catch (err) {
  throw err;
}
`
	checker := &JSConsoleErrorNoThrow{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSConsoleErrorNoThrow_Negative_ConsoleLog(t *testing.T) {
	src := `
try {
  load();
} catch (err) {
  console.log(err);
}
`
	checker := &JSConsoleErrorNoThrow{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for console.log, got %d", len(findings))
	}
}

func TestJSConsoleErrorNoThrow_Negative_TestFile(t *testing.T) {
	src := `
try {
  doSomething();
} catch (err) {
  console.error(err);
}
`
	checker := &JSConsoleErrorNoThrow{}
	findings := runJSCheckerOnSource(t, checker, "app.spec.ts", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSConsoleErrorNoThrow_MultipleCatches(t *testing.T) {
	src := `
try {
  a();
} catch (err) {
  console.error(err);
}

try {
  b();
} catch (err) {
  console.error(err);
  throw err;
}

try {
  c();
} catch (err) {
  console.error(err);
}
`
	checker := &JSConsoleErrorNoThrow{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}
