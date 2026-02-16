package rules

import (
	"testing"
)

func TestPyNoEncodingOpen(t *testing.T) {
	checker := &PyNoEncodingOpen{}

	t.Run("triggers on open without encoding", func(t *testing.T) {
		src := `f = open("data.txt")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "bad-defaults.py-no-encoding-open" {
			t.Errorf("expected rule bad-defaults.py-no-encoding-open, got %s", findings[0].Rule)
		}
	})

	t.Run("triggers on open with read mode no encoding", func(t *testing.T) {
		src := `f = open("data.txt", "r")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("triggers on open with write mode no encoding", func(t *testing.T) {
		src := `f = open("data.txt", "w")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("triggers with context manager", func(t *testing.T) {
		src := `with open("config.json") as cfg:
    data = cfg.read()
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("clean on open with encoding", func(t *testing.T) {
		src := `f = open("data.txt", encoding="utf-8")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on binary mode rb", func(t *testing.T) {
		src := `f = open("image.png", "rb")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on binary mode wb", func(t *testing.T) {
		src := `f = open("output.bin", "wb")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("clean on keyword binary mode", func(t *testing.T) {
		src := `f = open("data.dat", mode="rb")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("ignores qualified open calls", func(t *testing.T) {
		src := `import io
f = io.open("data.txt")
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for io.open, got %d", len(findings))
		}
	})
}
