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

	files, err := Walk(tmp, nil, nil)
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

	files, err := Walk(tmp, nil, nil)
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

	files, err := Walk(tmp, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file (app/server.go), got %d: %v", len(files), files)
	}
}

func TestWalk_CustomExcludeDirs(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "src"), 0o755)
	os.MkdirAll(filepath.Join(tmp, "generated"), 0o755)
	os.WriteFile(filepath.Join(tmp, "src", "main.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(tmp, "generated", "code.go"), []byte("package gen"), 0o644)

	// Default excludes + custom "generated" dir
	excludes := map[string]bool{
		".git":      true,
		"vendor":    true,
		"generated": true,
	}
	files, err := Walk(tmp, excludes, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(files), files)
	}
}

func TestWalk_ExcludePatterns(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(tmp, "main_generated.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(tmp, "api.pb.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(tmp, "utils.go"), []byte("package main"), 0o644)

	patterns := []string{"*_generated.go", "*.pb.go"}
	files, err := Walk(tmp, nil, patterns)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files (main.go, utils.go), got %d: %v", len(files), files)
	}
	for _, f := range files {
		base := filepath.Base(f)
		if base == "main_generated.go" || base == "api.pb.go" {
			t.Errorf("should not include pattern-excluded file: %s", base)
		}
	}
}

func TestWalk_ExcludePatternsSubdir(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "pkg"), 0o755)
	os.WriteFile(filepath.Join(tmp, "pkg", "service.go"), []byte("package pkg"), 0o644)
	os.WriteFile(filepath.Join(tmp, "pkg", "service_generated.go"), []byte("package pkg"), 0o644)

	patterns := []string{"*_generated.go"}
	files, err := Walk(tmp, nil, patterns)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file (pkg/service.go), got %d: %v", len(files), files)
	}
}

func TestWalk_NilPatterns_NoFiltering(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(tmp, "gen.pb.go"), []byte("package main"), 0o644)

	files, err := Walk(tmp, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files (no pattern filtering), got %d: %v", len(files), files)
	}
}

func TestWalk_FindsSupportedExtensions(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(tmp, "app.test.js"), []byte("describe('x', () => {})"), 0o644)
	os.WriteFile(filepath.Join(tmp, "types.spec.ts"), []byte("describe('x', () => {})"), 0o644)
	os.WriteFile(filepath.Join(tmp, "deploy.sh"), []byte("#!/bin/bash"), 0o644)
	os.WriteFile(filepath.Join(tmp, "readme.md"), []byte("# readme"), 0o644) // unsupported
	os.WriteFile(filepath.Join(tmp, "config.toml"), []byte("[x]"), 0o644)
	os.WriteFile(filepath.Join(tmp, "script.py"), []byte("print('hi')"), 0o644)
	os.WriteFile(filepath.Join(tmp, "stubs.pyi"), []byte("def f() -> int: ..."), 0o644)

	files, err := Walk(tmp, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 7 {
		t.Fatalf("expected 7 supported files, got %d: %v", len(files), files)
	}
}

func TestWalk_SkipsPycacheAndVenv(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "src"), 0o755)
	os.WriteFile(filepath.Join(tmp, "src", "app.py"), []byte("print('hi')"), 0o644)

	for _, dir := range []string{"__pycache__", ".venv", "dist", "build"} {
		os.MkdirAll(filepath.Join(tmp, dir), 0o755)
		os.WriteFile(filepath.Join(tmp, dir, "cached.py"), []byte("x=1"), 0o644)
	}

	files, err := Walk(tmp, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file (src/app.py), got %d: %v", len(files), files)
	}
}

func TestWalk_DefaultExcludePatterns(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "app.js"), []byte("var x = 1;"), 0o644)
	os.WriteFile(filepath.Join(tmp, "app.min.js"), []byte("var x=1;"), 0o644)
	os.WriteFile(filepath.Join(tmp, "vendor.bundle.js"), []byte("var x=1;"), 0o644)

	files, err := Walk(tmp, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file (app.js only), got %d: %v", len(files), files)
	}
	if filepath.Base(files[0]) != "app.js" {
		t.Errorf("expected app.js, got %s", filepath.Base(files[0]))
	}
}
