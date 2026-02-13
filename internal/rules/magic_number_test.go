package rules

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func runMagicNumber(t *testing.T, filename, src string) []int {
	t.Helper()
	checker := &MagicNumber{}
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

func TestMagicNumber_FindsLiteralGreaterThanOne(t *testing.T) {
	src := `package main

func run() {
	timeout := 30
	_ = timeout
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(lines))
	}
	if lines[0] != 4 {
		t.Errorf("expected finding on line 4, got %d", lines[0])
	}
}

func TestMagicNumber_IgnoresZeroAndOne(t *testing.T) {
	src := `package main

func run() {
	x := 0
	y := 1
	_, _ = x, y
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresArrayIndex(t *testing.T) {
	src := `package main

func run() {
	var arr []int
	_ = arr[2]
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresConstDeclaration(t *testing.T) {
	src := `package main

func run() {
	const retries = 3
	_ = retries
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(lines))
	}
}

func TestMagicNumber_SkipsTestFiles(t *testing.T) {
	src := `package main

func TestRun() {
	x := 42
	_ = x
}`
	lines := runMagicNumber(t, "main_test.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(lines))
	}
}
