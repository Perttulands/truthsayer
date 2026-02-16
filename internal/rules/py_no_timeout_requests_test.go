package rules

import (
	"testing"
)

func TestPyNoTimeoutRequests(t *testing.T) {
	checker := &PyNoTimeoutRequests{}

	t.Run("triggers on requests.get without timeout", func(t *testing.T) {
		src := `import requests
response = requests.get("https://api.example.com/data")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "bad-defaults.py-no-timeout-requests" {
			t.Errorf("expected rule bad-defaults.py-no-timeout-requests, got %s", findings[0].Rule)
		}
	})

	t.Run("triggers on requests.post without timeout", func(t *testing.T) {
		src := `import requests
response = requests.post("https://api.example.com/submit", json={"key": "value"})
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("triggers on multiple methods", func(t *testing.T) {
		src := `import requests
requests.get("https://api.example.com")
requests.put("https://api.example.com")
requests.delete("https://api.example.com")
requests.patch("https://api.example.com")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 4 {
			t.Fatalf("expected 4 findings, got %d", len(findings))
		}
	})

	t.Run("clean on requests.get with timeout", func(t *testing.T) {
		src := `import requests
response = requests.get("https://api.example.com/data", timeout=30)
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on requests.post with timeout tuple", func(t *testing.T) {
		src := `import requests
response = requests.post("https://api.example.com/submit", json={"key": "value"}, timeout=(5, 30))
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on kwargs spread", func(t *testing.T) {
		src := `import requests
config = {"timeout": 30}
response = requests.get("https://api.example.com/data", **config)
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("ignores non-requests calls", func(t *testing.T) {
		src := `import other
response = other.get("https://api.example.com/data")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})
}
