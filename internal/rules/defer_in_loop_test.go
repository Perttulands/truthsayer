package rules

import (
	"testing"
)

func TestDeferInLoop_ForLoop(t *testing.T) {
	checker := &DeferInLoop{}
	src := `package main

import "os"

func processFiles(paths []string) {
	for _, p := range paths {
		f, _ := os.Open(p)
		defer f.Close()
	}
}
`
	findings := runASTCheckerOnSource(t, checker, "test.go", src)
	if len(findings) == 0 {
		t.Fatal("expected finding for defer in loop, got none")
	}
	if findings[0].Rule != "bad-defaults.defer-in-loop" {
		t.Errorf("expected rule bad-defaults.defer-in-loop, got %s", findings[0].Rule)
	}
}

func TestDeferInLoop_NotInLoop(t *testing.T) {
	checker := &DeferInLoop{}
	src := `package main

import "os"

func processFile(path string) {
	f, _ := os.Open(path)
	defer f.Close()
}
`
	findings := runASTCheckerOnSource(t, checker, "test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected no findings for defer outside loop, got %d", len(findings))
	}
}

func TestDeferInLoop_NestedLoop(t *testing.T) {
	checker := &DeferInLoop{}
	src := `package main

import "os"

func processNested(matrix [][]string) {
	for _, row := range matrix {
		for _, p := range row {
			f, _ := os.Open(p)
			defer f.Close()
		}
	}
}
`
	findings := runASTCheckerOnSource(t, checker, "test.go", src)
	if len(findings) == 0 {
		t.Fatal("expected finding for defer in nested loop, got none")
	}
}

func TestDeferInLoop_DeferInIfInsideLoop(t *testing.T) {
	checker := &DeferInLoop{}
	src := `package main

import "os"

func conditionalDefer(paths []string) {
	for _, p := range paths {
		f, err := os.Open(p)
		if err == nil {
			defer f.Close()
		}
	}
}
`
	findings := runASTCheckerOnSource(t, checker, "test.go", src)
	if len(findings) == 0 {
		t.Fatal("expected finding for defer in if inside loop, got none")
	}
}

func TestDeferInLoop_SkipTestFiles(t *testing.T) {
	checker := &DeferInLoop{}
	src := `package main

import "os"

func processFiles(paths []string) {
	for _, p := range paths {
		f, _ := os.Open(p)
		defer f.Close()
	}
}
`
	findings := runASTCheckerOnSource(t, checker, "handler_test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected no findings for test file, got %d", len(findings))
	}
}
