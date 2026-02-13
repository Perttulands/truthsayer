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

func TestWalk_SkipsAllDefaultExcludedDirs(t *testing.T) {
	tmp := t.TempDir()

	// Create source dir with scannable file
	os.MkdirAll(filepath.Join(tmp, "src"), 0o755)
	os.WriteFile(filepath.Join(tmp, "src", "main.go"), []byte("package main"), 0o644)

	// Create all three excluded directories with Go files inside
	for _, dir := range []string{"vendor", "node_modules", ".git"} {
		os.MkdirAll(filepath.Join(tmp, dir), 0o755)
		os.WriteFile(filepath.Join(tmp, dir, "file.go"), []byte("package hidden"), 0o644)
	}

	files, err := Walk(tmp, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Only src/main.go should appear
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(files), files)
	}

	for _, f := range files {
		rel, _ := filepath.Rel(tmp, f)
		for _, excluded := range []string{"vendor", "node_modules", ".git"} {
			if rel == excluded || len(rel) > len(excluded) && rel[:len(excluded)+1] == excluded+string(filepath.Separator) {
				t.Errorf("should not include file from excluded dir %q: %s", excluded, rel)
			}
		}
	}
}

func TestWalk_SkipsNodeModulesNested(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "app"), 0o755)
	os.WriteFile(filepath.Join(tmp, "app", "server.go"), []byte("package app"), 0o644)

	// Nested node_modules with deeply nested files
	os.MkdirAll(filepath.Join(tmp, "node_modules", "pkg", "lib"), 0o755)
	os.WriteFile(filepath.Join(tmp, "node_modules", "pkg", "lib", "index.json"), []byte(`{}`), 0o644)
	os.WriteFile(filepath.Join(tmp, "node_modules", "pkg", "package.json"), []byte(`{}`), 0o644)

	files, err := Walk(tmp, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file (app/server.go), got %d: %v", len(files), files)
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
