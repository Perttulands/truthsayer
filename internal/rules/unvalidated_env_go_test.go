package rules

import "testing"

func TestUnvalidatedEnvGo_Flags(t *testing.T) {
	src := `package main

import "os"

func run() {
	key := os.Getenv("API_KEY")
	useKey(key)
}`
	findings := runASTCheckerOnSource(t, &UnvalidatedEnvGo{}, "main.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestUnvalidatedEnvGo_SkipsIfInit(t *testing.T) {
	src := `package main

import "os"

func run() {
	if key := os.Getenv("API_KEY"); key != "" {
		useKey(key)
	}
}`
	findings := runASTCheckerOnSource(t, &UnvalidatedEnvGo{}, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for if-init, got %d", len(findings))
	}
}

func TestUnvalidatedEnvGo_SkipsWhenValidatedNextLine(t *testing.T) {
	src := `package main

import "os"

func run() {
	key := os.Getenv("API_KEY")
	if key != "" {
		useKey(key)
	}
}`
	findings := runASTCheckerOnSource(t, &UnvalidatedEnvGo{}, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for next-line validation, got %d", len(findings))
	}
}

func TestUnvalidatedEnvGo_SkipsWhenValidatedWithTrimSpace(t *testing.T) {
	src := `package main

import (
	"os"
	"strings"
)

func run() {
	key := strings.TrimSpace(os.Getenv("API_KEY"))
	if key != "" {
		useKey(key)
	}
}`
	findings := runASTCheckerOnSource(t, &UnvalidatedEnvGo{}, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for TrimSpace+validation, got %d", len(findings))
	}
}

func TestUnvalidatedEnvGo_SkipsWhenValidatedEqualEmpty(t *testing.T) {
	src := `package main

import "os"

func run() {
	key := os.Getenv("API_KEY")
	if key == "" {
		return
	}
	useKey(key)
}`
	findings := runASTCheckerOnSource(t, &UnvalidatedEnvGo{}, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for == empty validation, got %d", len(findings))
	}
}
