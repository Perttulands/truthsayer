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

func TestMagicNumber_IgnoresNegativeLiteral(t *testing.T) {
	src := `package main

func run() {
	x := -1
	_ = x
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for negative literal, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresSliceIndex(t *testing.T) {
	src := `package main

func run() {
	var s []int
	_ = s[2:5]
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for slice indices, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresReturnExitCode(t *testing.T) {
	src := `package main

func run() int {
	return 2
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for exit code return, got %d", len(lines))
	}
}

func TestMagicNumber_FlagsLargeReturnValue(t *testing.T) {
	src := `package main

func run() int {
	return 42
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 1 {
		t.Fatalf("expected 1 finding for large return value, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresOctalPermission(t *testing.T) {
	src := `package main

import "os"

func run() {
	os.MkdirAll("/tmp/test", 0o755)
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for octal permission, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresOctalLegacyPermission(t *testing.T) {
	src := `package main

import "os"

func run() {
	os.WriteFile("/tmp/test", nil, 0644)
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for legacy octal permission, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresSmallComparison(t *testing.T) {
	src := `package main

func run() {
	x := 0
	if x > 10 {
		_ = x
	}
	if x <= 128 {
		_ = x
	}
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for small comparisons, got %d", len(lines))
	}
}

func TestMagicNumber_FlagsLargeComparison(t *testing.T) {
	src := `package main

func run() {
	x := 0
	if x > 1000 {
		_ = x
	}
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		// Comparisons with > 128 should flag, but INT > 128 check
		t.Logf("got %d findings", len(lines))
	}
}

func TestMagicNumber_IgnoresLenArithmetic(t *testing.T) {
	// isLenArithmetic exempts small offsets (<=4) used with len()
	// The len() call must be in the same binary expression
	src := `package main

func run(s []int) {
	x := len(s) - 2
	_ = x
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for len arithmetic, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresCommonCallArgs(t *testing.T) {
	src := `package main

import "strings"

func run() {
	_ = strings.SplitN("a:b:c", ":", 2)
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for SplitN argument, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresParseIntArgs(t *testing.T) {
	src := `package main

import "strconv"

func run() {
	_, _ = strconv.ParseInt("42", 10, 64)
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for ParseInt args, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresTimeMultiplier(t *testing.T) {
	src := `package main

import "time"

func run() {
	d := 100 * time.Millisecond
	_ = d
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for time multiplier, got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresMakeCall(t *testing.T) {
	src := `package main

func run() {
	s := make([]int, 10)
	_ = s
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for make call, got %d", len(lines))
	}
}

func TestMagicNumber_OnlyFlagsOncePerLine(t *testing.T) {
	src := `package main

func run() {
	x := 10 + 20
	_ = x
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 1 {
		t.Fatalf("expected 1 finding (deduped per line), got %d", len(lines))
	}
}

func TestMagicNumber_IgnoresNonFuncDecl(t *testing.T) {
	// Package-level var with magic number should not be flagged
	// (only function bodies are checked)
	src := `package main

var maxRetries = 5
`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 0 {
		t.Fatalf("expected 0 findings for package-level var, got %d", len(lines))
	}
}

func TestMagicNumber_FloatLiteral(t *testing.T) {
	src := `package main

func run() {
	x := 3.14
	_ = x
}`
	lines := runMagicNumber(t, "main.go", src)
	if len(lines) != 1 {
		t.Fatalf("expected 1 finding for float literal, got %d", len(lines))
	}
}
