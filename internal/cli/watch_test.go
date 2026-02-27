package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunWatch_NoArgs(t *testing.T) {
	code := runWatch(nil)
	if code != 2 {
		t.Errorf("expected exit code 2 for no args, got %d", code)
	}
}

func TestRunWatch_NonexistentPath(t *testing.T) {
	code := runWatch([]string{"/nonexistent/path/does/not/exist"})
	if code != 2 {
		t.Errorf("expected exit code 2 for nonexistent path, got %d", code)
	}
}

func TestRunWatch_FileNotDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.go")
	if err := os.WriteFile(path, []byte("package p\n"), 0644); err != nil {
		t.Fatal(err)
	}

	code := runWatch([]string{path})
	if code != 2 {
		t.Errorf("expected exit code 2 for file (not dir), got %d", code)
	}
}

func TestExitCode_True(t *testing.T) {
	if exitCode(true) != 1 {
		t.Errorf("expected 1 for sawError=true")
	}
}

func TestExitCode_False(t *testing.T) {
	if exitCode(false) != 0 {
		t.Errorf("expected 0 for sawError=false")
	}
}

func TestRunWatch_ConfigFlag(t *testing.T) {
	code := runWatch([]string{"--config", "/nonexistent/config.toml", t.TempDir()})
	if code != 2 {
		t.Errorf("expected exit code 2 for bad config, got %d", code)
	}
}
