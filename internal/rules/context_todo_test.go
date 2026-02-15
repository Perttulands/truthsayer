package rules

import "testing"

func TestContextTodo_DetectsContextTODO(t *testing.T) {
	src := `package main

import "context"

func serve() {
	ctx := context.TODO()
	_ = ctx
}`
	findings := runASTCheckerOnSource(t, &ContextTodo{}, "server.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "bad-defaults.context-todo" {
		t.Fatalf("expected rule bad-defaults.context-todo, got %s", findings[0].Rule)
	}
}

func TestContextTodo_DetectsContextBackground(t *testing.T) {
	src := `package main

import "context"

func serve() {
	ctx := context.Background()
	_ = ctx
}`
	findings := runASTCheckerOnSource(t, &ContextTodo{}, "server.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestContextTodo_SkipsMainFunction(t *testing.T) {
	src := `package main

import "context"

func main() {
	ctx := context.Background()
	_ = ctx
}`
	findings := runASTCheckerOnSource(t, &ContextTodo{}, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestContextTodo_SkipsInitFunction(t *testing.T) {
	src := `package main

import "context"

func init() {
	ctx := context.Background()
	_ = ctx
}`
	findings := runASTCheckerOnSource(t, &ContextTodo{}, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestContextTodo_SkipsTestFiles(t *testing.T) {
	src := `package main

import "context"

func serve() {
	ctx := context.TODO()
	_ = ctx
}`
	findings := runASTCheckerOnSource(t, &ContextTodo{}, "main_test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
