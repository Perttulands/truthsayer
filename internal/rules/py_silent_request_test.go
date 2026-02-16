package rules

import (
	"testing"
)

func TestPySilentRequest(t *testing.T) {
	checker := &PySilentRequest{}

	t.Run("detects requests.get without status check", func(t *testing.T) {
		src := `import requests
response = requests.get("https://api.example.com/data")
data = response.json()
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "trace-gaps.py-silent-request" {
			t.Errorf("expected rule trace-gaps.py-silent-request, got %s", findings[0].Rule)
		}
	})

	t.Run("detects requests.post without status check", func(t *testing.T) {
		src := `import requests
requests.post("https://api.example.com/submit", json={"key": "value"})
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("clean with raise_for_status", func(t *testing.T) {
		src := `import requests
response = requests.get("https://api.example.com/data")
response.raise_for_status()
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean with status_code check", func(t *testing.T) {
		src := `import requests
resp = requests.post("https://api.example.com/submit", json={"key": "value"})
if resp.status_code != 200:
    raise Exception("Request failed")
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean with chained raise_for_status", func(t *testing.T) {
		src := `import requests
requests.get("https://api.example.com/health").raise_for_status()
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean with .ok check", func(t *testing.T) {
		src := `import requests
r = requests.get("https://api.example.com/check")
if not r.ok:
    raise Exception("Not ok")
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("detects multiple unchecked requests", func(t *testing.T) {
		src := `import requests
requests.get("https://api.example.com/one")
requests.post("https://api.example.com/two")
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("no false positive on non-requests module", func(t *testing.T) {
		src := `import urllib
response = urllib.request.urlopen("https://example.com")
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})
}
