package cli

import (
	"os"
	"strings"
	"testing"
)

func TestVersion_PrintsVersionString(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Version = "1.2.3"
	code := runVersion()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "truthsayer version 1.2.3") {
		t.Fatalf("expected version string, got: %s", output)
	}
}

func TestVersion_DefaultVersion(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Version = ""
	code := runVersion()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "dev") {
		t.Fatalf("expected 'dev' default version, got: %s", output)
	}
}

func TestVersion_CLIDispatch(t *testing.T) {
	Version = "0.5.0"

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"truthsayer", "--version"}
	code := Run()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "truthsayer version 0.5.0") {
		t.Fatalf("expected version in CLI dispatch output, got: %s", output)
	}
}
