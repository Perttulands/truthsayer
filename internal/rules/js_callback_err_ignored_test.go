package rules

import "testing"

func TestJSCallbackErrIgnored_DetectsIgnoredErr(t *testing.T) {
	src := `
const fs = require('fs');
fs.readFile('x', (err, data) => {
  console.log(data);
});
`
	findings := runJSCheckerOnSource(t, &JSCallbackErrIgnored{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.js-callback-err-ignored" {
		t.Errorf("expected rule silent-fallback.js-callback-err-ignored, got %s", findings[0].Rule)
	}
}

func TestJSCallbackErrIgnored_SkipsWhenErrUsed(t *testing.T) {
	src := `
fs.readFile('x', (err, data) => {
  if (err) throw err;
  console.log(data);
});
`
	findings := runJSCheckerOnSource(t, &JSCallbackErrIgnored{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSCallbackErrIgnored_DetectsErrorParam(t *testing.T) {
	src := `
doWork((error, result) => {
  return result;
});
`
	findings := runJSCheckerOnSource(t, &JSCallbackErrIgnored{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSCallbackErrIgnored_SkipsSingleParam(t *testing.T) {
	// Single-param callbacks are not err-first pattern
	src := `
items.forEach((item) => {
  console.log(item);
});
`
	findings := runJSCheckerOnSource(t, &JSCallbackErrIgnored{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for single-param callback, got %d", len(findings))
	}
}

func TestJSCallbackErrIgnored_SkipsNonErrName(t *testing.T) {
	// First param not named err/error/e
	src := `
doWork((response, data) => {
  console.log(data);
});
`
	findings := runJSCheckerOnSource(t, &JSCallbackErrIgnored{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-err name, got %d", len(findings))
	}
}

func TestJSCallbackErrIgnored_SkipsNonCallback(t *testing.T) {
	// Arrow function not in arguments — regular assignment
	src := `
const handler = (err, data) => {
  console.log(data);
};
`
	findings := runJSCheckerOnSource(t, &JSCallbackErrIgnored{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-callback arrow, got %d", len(findings))
	}
}

func TestJSCallbackErrIgnored_DetectsRegularFunction(t *testing.T) {
	src := `
fs.readFile('x', function(err, data) {
  console.log(data);
});
`
	findings := runJSCheckerOnSource(t, &JSCallbackErrIgnored{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for function callback, got %d", len(findings))
	}
}

func TestJSCallbackErrIgnored_SkipsTestFiles(t *testing.T) {
	src := `
fs.readFile('x', (err, data) => {
  console.log(data);
});
`
	findings := runJSCheckerOnSource(t, &JSCallbackErrIgnored{}, "app.test.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSCallbackErrIgnored_DetectsEShortName(t *testing.T) {
	src := `
doWork((e, result) => {
  return result;
});
`
	findings := runJSCheckerOnSource(t, &JSCallbackErrIgnored{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for 'e' param, got %d", len(findings))
	}
}
