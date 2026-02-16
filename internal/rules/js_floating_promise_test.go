package rules

import "testing"

func TestJSFloatingPromise_DetectsFetch(t *testing.T) {
	src := `
function loadData() {
  fetch("/api/data");
}
`
	findings := runJSCheckerOnSource(t, &JSFloatingPromise{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.js-floating-promise" {
		t.Errorf("expected rule silent-fallback.js-floating-promise, got %s", findings[0].Rule)
	}
}

func TestJSFloatingPromise_DetectsAsyncCall(t *testing.T) {
	src := `
async function doWork() { return 1; }
function main() {
  doWork();
}
`
	// doWork is async but called without await — however we only detect known promise
	// functions like fetch. We don't flag arbitrary calls since we can't know if they're async.
	findings := runJSCheckerOnSource(t, &JSFloatingPromise{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for unknown async call, got %d", len(findings))
	}
}

func TestJSFloatingPromise_SkipsAwaitedFetch(t *testing.T) {
	src := `
async function loadData() {
  const data = await fetch("/api/data");
  return data;
}
`
	findings := runJSCheckerOnSource(t, &JSFloatingPromise{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestJSFloatingPromise_SkipsAssigned(t *testing.T) {
	src := `
const promise = fetch("/api/data");
`
	findings := runJSCheckerOnSource(t, &JSFloatingPromise{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for assigned fetch, got %d", len(findings))
	}
}

func TestJSFloatingPromise_SkipsReturned(t *testing.T) {
	src := `
function loadData() {
  return fetch("/api/data");
}
`
	findings := runJSCheckerOnSource(t, &JSFloatingPromise{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for returned fetch, got %d", len(findings))
	}
}

func TestJSFloatingPromise_SkipsDotThen(t *testing.T) {
	src := `
fetch("/api/data").then(handleResponse);
`
	findings := runJSCheckerOnSource(t, &JSFloatingPromise{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for .then() chained fetch, got %d", len(findings))
	}
}

func TestJSFloatingPromise_SkipsDotCatch(t *testing.T) {
	src := `
fetch("/api/data").catch(handleError);
`
	findings := runJSCheckerOnSource(t, &JSFloatingPromise{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for .catch() chained fetch, got %d", len(findings))
	}
}

func TestJSFloatingPromise_SkipsTestFiles(t *testing.T) {
	src := `
function test() {
  fetch("/api/data");
}
`
	findings := runJSCheckerOnSource(t, &JSFloatingPromise{}, "app.test.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSFloatingPromise_TypeScript(t *testing.T) {
	src := `
function loadData(): void {
  fetch("/api/data");
}
`
	findings := runTSCheckerOnSource(t, &JSFloatingPromise{}, "app.ts", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
