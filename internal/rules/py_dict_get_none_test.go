package rules

import "testing"

func TestPyDictGetNone_NoDefault(t *testing.T) {
	src := `
data = {"name": "Alice"}
x = data.get("name")
`
	findings := runPyCheckerOnSource(t, &PyDictGetNone{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.py-dict-get-none" {
		t.Errorf("wrong rule: %s", findings[0].Rule)
	}
}

func TestPyDictGetNone_NegativeWithDefault(t *testing.T) {
	src := `
data = {"name": "Alice"}
x = data.get("name", "unknown")
`
	findings := runPyCheckerOnSource(t, &PyDictGetNone{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPyDictGetNone_NegativeWithNoneDefault(t *testing.T) {
	src := `
data = {"name": "Alice"}
x = data.get("name", None)
`
	findings := runPyCheckerOnSource(t, &PyDictGetNone{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for explicit None default (intentional), got %d", len(findings))
	}
}

func TestPyDictGetNone_NegativeDirectAccess(t *testing.T) {
	src := `
data = {"name": "Alice"}
x = data["name"]
`
	findings := runPyCheckerOnSource(t, &PyDictGetNone{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for direct access, got %d", len(findings))
	}
}

func TestPyDictGetNone_NegativeNonDictGet(t *testing.T) {
	src := `
import requests
response = requests.get("https://example.com")
`
	findings := runPyCheckerOnSource(t, &PyDictGetNone{}, "app.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for requests.get, got %d", len(findings))
	}
}

func TestPyDictGetNone_Multiple(t *testing.T) {
	src := `
config = {"host": "localhost", "port": 8080}
host = config.get("host")
port = config.get("port")
debug = config.get("debug", False)
`
	findings := runPyCheckerOnSource(t, &PyDictGetNone{}, "app.py", src)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestPyDictGetNone_ChainedCall(t *testing.T) {
	src := `
data = {"items": [1, 2, 3]}
items = data.get("items")
`
	findings := runPyCheckerOnSource(t, &PyDictGetNone{}, "app.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
