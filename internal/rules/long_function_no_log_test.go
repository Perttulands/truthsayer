package rules

import (
	"strings"
	"testing"
)

func TestLongFunctionNoLog_LongNoLog(t *testing.T) {
	checker := &LongFunctionNoLog{}
	// Build a function with >30 lines and no log calls
	var lines []string
	lines = append(lines, "package p")
	lines = append(lines, "func Process() error {")
	for i := 0; i < 35; i++ {
		lines = append(lines, "\tx := "+strings.Repeat("a", 5))
	}
	lines = append(lines, "\treturn nil")
	lines = append(lines, "}")
	src := strings.Join(lines, "\n")

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestLongFunctionNoLog_LongWithLog(t *testing.T) {
	checker := &LongFunctionNoLog{}
	var lines []string
	lines = append(lines, "package p")
	lines = append(lines, `import "log"`)
	lines = append(lines, "func Process() error {")
	for i := 0; i < 35; i++ {
		lines = append(lines, "\tx := 1")
	}
	lines = append(lines, `	log.Println("processing")`)
	lines = append(lines, "\treturn nil")
	lines = append(lines, "}")
	src := strings.Join(lines, "\n")

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when log call present, got %d", len(findings))
	}
}

func TestLongFunctionNoLog_ShortFunction(t *testing.T) {
	checker := &LongFunctionNoLog{}
	src := `package p

func Short() {
	x := 1
	y := 2
	_ = x + y
}
`
	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for short function, got %d", len(findings))
	}
}

func TestLongFunctionNoLog_SkipsTestFile(t *testing.T) {
	checker := &LongFunctionNoLog{}
	var lines []string
	lines = append(lines, "package p")
	lines = append(lines, "func TestBigThing() {")
	for i := 0; i < 35; i++ {
		lines = append(lines, "\tx := 1")
	}
	lines = append(lines, "}")
	src := strings.Join(lines, "\n")

	findings := runASTCheckerOnSource(t, checker, "handler_test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestLongFunctionNoLog_UnexportedSkipped(t *testing.T) {
	checker := &LongFunctionNoLog{}
	var lines []string
	lines = append(lines, "package p")
	lines = append(lines, "func helperFunc() error {")
	for i := 0; i < 35; i++ {
		lines = append(lines, "\tx := 1")
	}
	lines = append(lines, "\treturn nil")
	lines = append(lines, "}")
	src := strings.Join(lines, "\n")

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for unexported helper, got %d", len(findings))
	}
}
