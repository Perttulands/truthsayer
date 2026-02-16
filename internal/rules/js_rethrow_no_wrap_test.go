package rules

import (
	"testing"
)

func TestJSRethrowNoWrap_Positive(t *testing.T) {
	src := `
try {
  doSomething();
} catch (err) {
  throw err;
}

try {
  doSomethingElse();
} catch (e) {
  throw e;
}
`
	checker := &JSRethrowNoWrap{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	for _, f := range findings {
		if f.Rule != "error-context.js-rethrow-no-wrap" {
			t.Errorf("unexpected rule ID: %s", f.Rule)
		}
	}
}

func TestJSRethrowNoWrap_RethrowAfterStatements(t *testing.T) {
	src := `
try {
  doWork();
} catch (error) {
  console.log("oops");
  throw error;
}
`
	checker := &JSRethrowNoWrap{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSRethrowNoWrap_Negative_WrappedError(t *testing.T) {
	src := `
try {
  doSomething();
} catch (err) {
  throw new Error("failed", { cause: err });
}
`
	checker := &JSRethrowNoWrap{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSRethrowNoWrap_Negative_DifferentError(t *testing.T) {
	src := `
try {
  parse();
} catch (err) {
  throw new TypeError("invalid input");
}
`
	checker := &JSRethrowNoWrap{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSRethrowNoWrap_Negative_NoThrow(t *testing.T) {
	src := `
try {
  load();
} catch (err) {
  console.error(err);
  return null;
}
`
	checker := &JSRethrowNoWrap{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSRethrowNoWrap_Negative_TestFile(t *testing.T) {
	src := `
try {
  doSomething();
} catch (err) {
  throw err;
}
`
	checker := &JSRethrowNoWrap{}
	findings := runJSCheckerOnSource(t, checker, "app.test.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSRethrowNoWrap_ParameterlessCatch(t *testing.T) {
	src := `
try {
  doSomething();
} catch {
  throw new Error("something failed");
}
`
	checker := &JSRethrowNoWrap{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for parameterless catch, got %d", len(findings))
	}
}
