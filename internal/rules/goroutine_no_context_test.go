package rules

import "testing"

func TestGoroutineNoContext_NoContext(t *testing.T) {
	checker := &GoroutineNoContext{}
	src := `package p

func work() {
	go func() {
		doSomething()
	}()
}
`
	findings := runASTCheckerOnSource(t, checker, "server.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestGoroutineNoContext_WithCtx(t *testing.T) {
	checker := &GoroutineNoContext{}
	src := `package p

import "context"

func work(ctx context.Context) {
	go func(ctx context.Context) {
		doSomething(ctx)
	}(ctx)
}
`
	findings := runASTCheckerOnSource(t, checker, "server.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings with context param, got %d", len(findings))
	}
}

func TestGoroutineNoContext_NamedFuncWithCtxArg(t *testing.T) {
	checker := &GoroutineNoContext{}
	src := `package p

import "context"

func work(ctx context.Context) {
	go handleRequest(ctx)
}
`
	findings := runASTCheckerOnSource(t, checker, "server.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for named func with ctx arg, got %d", len(findings))
	}
}

func TestGoroutineNoContext_NamedFuncNoCtx(t *testing.T) {
	checker := &GoroutineNoContext{}
	src := `package p

func work() {
	go doBackground()
}
`
	findings := runASTCheckerOnSource(t, checker, "server.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for named func without ctx, got %d", len(findings))
	}
}

func TestGoroutineNoContext_SkipsTestFiles(t *testing.T) {
	checker := &GoroutineNoContext{}
	src := `package p

func TestSomething() {
	go func() {
		doSomething()
	}()
}
`
	findings := runASTCheckerOnSource(t, checker, "server_test.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestGoroutineNoContext_BodyReferencesCtx(t *testing.T) {
	checker := &GoroutineNoContext{}
	src := `package p

import "context"

func work() {
	ctx := context.Background()
	go func() {
		doSomething(ctx)
	}()
}
`
	findings := runASTCheckerOnSource(t, checker, "server.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when body references ctx, got %d", len(findings))
	}
}
