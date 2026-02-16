package rules

import (
	"testing"
)

func TestPyGlobalState(t *testing.T) {
	checker := &PyGlobalState{}

	t.Run("triggers on module-level list", func(t *testing.T) {
		src := `ITEMS = []
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "bad-defaults.py-global-state" {
			t.Errorf("expected rule bad-defaults.py-global-state, got %s", findings[0].Rule)
		}
	})

	t.Run("triggers on module-level dict", func(t *testing.T) {
		src := `CACHE = {}
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("triggers on module-level set", func(t *testing.T) {
		src := `SEEN = set()
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("triggers on multiple globals", func(t *testing.T) {
		src := `ITEMS = []
CACHE = {}
SEEN = set()
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 3 {
			t.Fatalf("expected 3 findings, got %d", len(findings))
		}
	})

	t.Run("clean on immutable constants", func(t *testing.T) {
		src := `MAX_RETRIES = 3
API_URL = "https://api.example.com"
DEBUG = False
VERSION = "1.0.0"
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on tuple constant", func(t *testing.T) {
		src := `VALID_STATUSES = (200, 201, 204)
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("ignores function-level state", func(t *testing.T) {
		src := `def get_cache():
    cache = {}
    return cache
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("ignores class-level state", func(t *testing.T) {
		src := `class Registry:
    _handlers = {}
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		src := `FIXTURES = []
TEST_DATA = {}
`
		findings := runPyCheckerOnSource(t, checker, "test_app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for test file, got %d", len(findings))
		}
	})
}
