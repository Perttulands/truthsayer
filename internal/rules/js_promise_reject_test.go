package rules

import (
	"testing"
)

func TestJSPromiseReject_Positive_PromiseReject(t *testing.T) {
	src := `
Promise.reject("something failed");
Promise.reject(42);
Promise.reject(null);
Promise.reject(undefined);
Promise.reject(true);
`
	checker := &JSPromiseReject{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 5 {
		t.Fatalf("expected 5 findings, got %d", len(findings))
	}
	for _, f := range findings {
		if f.Rule != "error-context.js-promise-reject-non-error" {
			t.Errorf("unexpected rule ID: %s", f.Rule)
		}
	}
}

func TestJSPromiseReject_Positive_BareReject(t *testing.T) {
	src := `
new Promise((resolve, reject) => {
  reject("bad request");
});
`
	checker := &JSPromiseReject{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSPromiseReject_Negative_ErrorObject(t *testing.T) {
	src := `
Promise.reject(new Error("something failed"));
Promise.reject(new TypeError("invalid input"));
`
	checker := &JSPromiseReject{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSPromiseReject_Negative_Variable(t *testing.T) {
	src := `
Promise.reject(err);
Promise.reject(error);
`
	checker := &JSPromiseReject{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for variable args, got %d", len(findings))
	}
}

func TestJSPromiseReject_Negative_NoArgs(t *testing.T) {
	src := `Promise.reject();`
	checker := &JSPromiseReject{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for no args, got %d", len(findings))
	}
}

func TestJSPromiseReject_Negative_TestFile(t *testing.T) {
	src := `Promise.reject("test error");`
	checker := &JSPromiseReject{}
	findings := runJSCheckerOnSource(t, checker, "app.test.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSPromiseReject_TemplateLiteral(t *testing.T) {
	src := "Promise.reject(`error: ${msg}`);"
	checker := &JSPromiseReject{}
	findings := runJSCheckerOnSource(t, checker, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for template literal reject, got %d", len(findings))
	}
}
