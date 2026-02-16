package rules

import "testing"

func TestJSEmptyCatch_DetectsEmptyBody(t *testing.T) {
	src := `
try {
  doSomething();
} catch (e) {
}
`
	findings := runJSCheckerOnSource(t, &JSEmptyCatch{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.js-empty-catch" {
		t.Errorf("expected rule silent-fallback.js-empty-catch, got %s", findings[0].Rule)
	}
}

func TestJSEmptyCatch_SkipsNonEmpty(t *testing.T) {
	src := `
try {
  doSomething();
} catch (e) {
  console.error(e);
}
`
	findings := runJSCheckerOnSource(t, &JSEmptyCatch{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSEmptyCatch_SkipsCommentOnly(t *testing.T) {
	// A catch with only a comment is intentional silencing — skip it
	src := `
try {
  doSomething();
} catch (e) {
  // intentionally ignored
}
`
	findings := runJSCheckerOnSource(t, &JSEmptyCatch{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for comment-only catch, got %d", len(findings))
	}
}

func TestJSEmptyCatch_DetectsMultiple(t *testing.T) {
	src := `
try { a(); } catch (e) {}
try { b(); } catch (err) {}
`
	findings := runJSCheckerOnSource(t, &JSEmptyCatch{}, "app.js", src)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestJSEmptyCatch_SkipsTestFiles(t *testing.T) {
	src := `
try { doSomething(); } catch (e) {}
`
	findings := runJSCheckerOnSource(t, &JSEmptyCatch{}, "app.test.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSEmptyCatch_DetectsNoParamCatch(t *testing.T) {
	// catch without parameter (ES2019+)
	src := `
try {
  doSomething();
} catch {
}
`
	findings := runJSCheckerOnSource(t, &JSEmptyCatch{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for catch without param, got %d", len(findings))
	}
}
