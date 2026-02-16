package rules

import "testing"

func TestJSCatchReturnNull_DetectsCatchNull(t *testing.T) {
	src := `
const data = fetch("/api").catch(() => null);
`
	findings := runJSCheckerOnSource(t, &JSCatchReturnNull{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.js-catch-return-null" {
		t.Errorf("expected rule silent-fallback.js-catch-return-null, got %s", findings[0].Rule)
	}
}

func TestJSCatchReturnNull_DetectsCatchUndefined(t *testing.T) {
	src := `
const data = fetch("/api").catch(() => undefined);
`
	findings := runJSCheckerOnSource(t, &JSCatchReturnNull{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSCatchReturnNull_DetectsBlockReturnNull(t *testing.T) {
	src := `
const data = fetch("/api").catch((err) => { return null; });
`
	findings := runJSCheckerOnSource(t, &JSCatchReturnNull{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSCatchReturnNull_SkipsProperHandling(t *testing.T) {
	src := `
const data = fetch("/api").catch((err) => {
  console.error(err);
  return defaultValue;
});
`
	findings := runJSCheckerOnSource(t, &JSCatchReturnNull{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSCatchReturnNull_SkipsCatchWithLogging(t *testing.T) {
	src := `
const data = fetch("/api").catch((err) => {
  logger.error(err);
  return null;
});
`
	// This still returns null but at least logs — could be intentional.
	// We still flag it since the return value is null.
	findings := runJSCheckerOnSource(t, &JSCatchReturnNull{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSCatchReturnNull_SkipsTestFiles(t *testing.T) {
	src := `
const data = fetch("/api").catch(() => null);
`
	findings := runJSCheckerOnSource(t, &JSCatchReturnNull{}, "app.test.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSCatchReturnNull_DetectsThenCatchNull(t *testing.T) {
	src := `
promise.then(handleSuccess).catch(() => null);
`
	findings := runJSCheckerOnSource(t, &JSCatchReturnNull{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
