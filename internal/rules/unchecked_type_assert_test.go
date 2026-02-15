package rules

import (
	"testing"
)

func TestUncheckedTypeAssert_Flagged(t *testing.T) {
	checker := &UncheckedTypeAssert{}
	src := `package main

type MyInterface interface{ Do() }
type MyStruct struct{}
func (MyStruct) Do() {}

func process(i interface{}) {
	s := i.(MyStruct)
	_ = s
}
`
	findings := runASTCheckerOnSource(t, checker, "test.go", src)
	if len(findings) == 0 {
		t.Fatal("expected finding for unchecked type assertion, got none")
	}
	if findings[0].Rule != "silent-fallback.unchecked-type-assert" {
		t.Errorf("expected rule silent-fallback.unchecked-type-assert, got %s", findings[0].Rule)
	}
}

func TestUncheckedTypeAssert_CommaOk(t *testing.T) {
	checker := &UncheckedTypeAssert{}
	src := `package main

func process(i interface{}) {
	s, ok := i.(string)
	if ok {
		_ = s
	}
}
`
	findings := runASTCheckerOnSource(t, checker, "test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected no findings for comma-ok assertion, got %d", len(findings))
	}
}

func TestUncheckedTypeAssert_TypeSwitch(t *testing.T) {
	checker := &UncheckedTypeAssert{}
	src := `package main

func process(i interface{}) {
	switch i.(type) {
	case string:
	case int:
	}
}
`
	findings := runASTCheckerOnSource(t, checker, "test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected no findings for type switch, got %d", len(findings))
	}
}

func TestUncheckedTypeAssert_SkipTestFiles(t *testing.T) {
	checker := &UncheckedTypeAssert{}
	src := `package main

func process(i interface{}) {
	s := i.(string)
	_ = s
}
`
	findings := runASTCheckerOnSource(t, checker, "handler_test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected no findings for test file, got %d", len(findings))
	}
}

func TestUncheckedTypeAssert_InExpression(t *testing.T) {
	checker := &UncheckedTypeAssert{}
	src := `package main

import "fmt"

func process(i interface{}) {
	fmt.Println(i.(string))
}
`
	findings := runASTCheckerOnSource(t, checker, "test.go", src)
	if len(findings) == 0 {
		t.Fatal("expected finding for unchecked type assertion in expression, got none")
	}
}
