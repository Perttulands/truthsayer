package cli

import "testing"

func TestScan_AppliesSenateSeverityAmendment(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.go", goWithError)
	writeFile(t, dir, ".truthsayer-amendments.json", `[
  {
    "verdict_id": "quick-1",
    "applied_at": "2026-02-20T00:00:00Z",
    "amendment": {
      "rule_id": "silent-fallback.empty-error-check",
      "action": "set_severity",
      "severity": "warning"
    }
  }
]`)

	code := runScan([]string{dir})
	if code != 0 {
		t.Fatalf("expected exit code 0 after severity amendment, got %d", code)
	}
}
