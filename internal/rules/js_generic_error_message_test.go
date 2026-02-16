package rules

import (
	"testing"
)

func TestJSGenericErrorMessage_Positive(t *testing.T) {
	src := `
throw new Error("failed");
throw new Error("error");
throw new Error("something went wrong");
throw new TypeError("unknown error");
`
	checker := &JSGenericErrorMessage{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 4 {
		t.Fatalf("expected 4 findings, got %d", len(findings))
	}
	for _, f := range findings {
		if f.Rule != "error-context.js-generic-error-message" {
			t.Errorf("unexpected rule ID: %s", f.Rule)
		}
	}
}

func TestJSGenericErrorMessage_Negative_Descriptive(t *testing.T) {
	src := `
throw new Error("Failed to parse config file");
throw new TypeError("Expected a number but got a string");
throw new Error("Connection to redis timed out after 30s");
`
	checker := &JSGenericErrorMessage{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSGenericErrorMessage_Negative_TemplateLiteral(t *testing.T) {
	src := "throw new Error(`User ${userId} not found`);"
	checker := &JSGenericErrorMessage{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for template literal, got %d", len(findings))
	}
}

func TestJSGenericErrorMessage_Negative_CustomError(t *testing.T) {
	src := `throw new CustomError("failed");`
	checker := &JSGenericErrorMessage{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for custom error, got %d", len(findings))
	}
}

func TestJSGenericErrorMessage_Negative_NoArgs(t *testing.T) {
	src := `throw new Error();`
	checker := &JSGenericErrorMessage{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for no-arg Error, got %d", len(findings))
	}
}

func TestJSGenericErrorMessage_Negative_TestFile(t *testing.T) {
	src := `throw new Error("failed");`
	checker := &JSGenericErrorMessage{}
	findings := runJSCheckerOnSource(t, checker, "app.test.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSGenericErrorMessage_CaseInsensitive(t *testing.T) {
	src := `
throw new Error("Failed");
throw new Error("ERROR");
throw new Error("Oops");
`
	checker := &JSGenericErrorMessage{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings (case insensitive), got %d", len(findings))
	}
}
