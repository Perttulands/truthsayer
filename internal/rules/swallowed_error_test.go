package rules

import "testing"

func TestSwallowedError_DetectsLogWithoutReturn(t *testing.T) {
	src := `package main

import "log"

func process() error {
	err := doWork()
	if err != nil {
		log.Println(err)
	}
	return nil
}`
	findings := runASTCheckerOnSource(t, &SwallowedError{}, "main.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "error-context.swallowed-error" {
		t.Fatalf("expected rule error-context.swallowed-error, got %s", findings[0].Rule)
	}
}

func TestSwallowedError_SkipsWhenReturnPresent(t *testing.T) {
	src := `package main

import "log"

func process() error {
	err := doWork()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}`
	findings := runASTCheckerOnSource(t, &SwallowedError{}, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestSwallowedError_SkipsTestFiles(t *testing.T) {
	src := `package main

import "log"

func TestThing() {
	err := doWork()
	if err != nil {
		log.Println(err)
	}
}`
	findings := runASTCheckerOnSource(t, &SwallowedError{}, "main_test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
