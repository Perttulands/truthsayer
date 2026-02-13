package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalk_SkipsExcludedDirs(t *testing.T) {
	// Create a temp directory structure
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "src"), 0o755)
	os.MkdirAll(filepath.Join(tmp, "vendor", "lib"), 0o755)
	os.MkdirAll(filepath.Join(tmp, ".git", "objects"), 0o755)
	os.WriteFile(filepath.Join(tmp, "src", "main.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(tmp, "vendor", "lib", "dep.go"), []byte("package dep"), 0o644)
	os.WriteFile(filepath.Join(tmp, ".git", "objects", "file"), []byte("data"), 0o644)

	files, err := Walk(tmp, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		rel, _ := filepath.Rel(tmp, f)
		if filepath.HasPrefix(rel, "vendor") || filepath.HasPrefix(rel, ".git") {
			t.Errorf("should not include file from excluded dir: %s", rel)
		}
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file (src/main.go), got %d: %v", len(files), files)
	}
}

func TestWalk_FindsSupportedExtensions(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(tmp, "deploy.sh"), []byte("#!/bin/bash"), 0o644)
	os.WriteFile(filepath.Join(tmp, "readme.md"), []byte("# readme"), 0o644) // unsupported
	os.WriteFile(filepath.Join(tmp, "config.toml"), []byte("[x]"), 0o644)

	files, err := Walk(tmp, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 supported files, got %d: %v", len(files), files)
	}
}
