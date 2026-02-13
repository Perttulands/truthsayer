package rules

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func runDebugGuard(t *testing.T, filename, src string) []int {
	t.Helper()
	checker := &DebugGuard{}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	lines := strings.Split(src, "\n")
	findings := checker.CheckAST(fset, file, lines)

	var findingLines []int
	for _, f := range findings {
		findingLines = append(findingLines, f.Line)
	}
	return findingLines
}

func TestDebugGuard_FindsDebugCondition(t *testing.T) {
	src := `package main

func run() {
	if debug {
		return
	}
}`
	lines := runDebugGuard(t, "main.go", src)
	if len(lines) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(lines))
	}
	if lines[0] != 4 {
		t.Errorf("expected finding on line 4, got %d", lines[0])
	}
}

func TestDebugGuard_FindsIsTestCondition(t *testing.T) {
	src := `package main

func run() {
	if isTest {
		return
	}
}`
	lines := runDebugGuard(t, "main.go", src)
	if len(lines) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(lines))
	}
}

func TestDebugGuard_IgnoresNormalCondition(t *testing.T) {
	src := `package main

func run(err error) {
	if err != nil {
		return
	}
}`
	lines := runDebugGuard(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(lines))
	}
}

func TestDebugGuard_SkipsTestFile(t *testing.T) {
	src := `package main

func TestRun() {
	if debug {
		return
	}
}`
	lines := runDebugGuard(t, "main_test.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(lines))
	}
}
