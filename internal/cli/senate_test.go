package cli

import (
	"strings"
	"testing"
)

func TestRunSenateParse(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "verdict.json", `{
  "id":"quick-1771535739",
  "status":"approved",
  "amendments":[{"rule_id":"silent-fallback.hidden-failure-bash","action":"disable_rule"}]
}`)

	out := captureStdout(t, func() {
		code := runSenate([]string{"parse", path})
		if code != 0 {
			t.Fatalf("expected exit code 0, got %d", code)
		}
	})
	if !strings.Contains(out, `"id": "quick-1771535739"`) {
		t.Fatalf("expected parsed verdict output, got:\n%s", out)
	}
}

func TestRunSenateParse_Invalid(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "verdict.json", `{"id":"v","status":"approved","amendments":[{"rule_id":"r","action":"bad"}]}`)
	code := runSenate([]string{"parse", path})
	if code != 2 {
		t.Fatalf("expected exit code 2 for invalid verdict, got %d", code)
	}
}

