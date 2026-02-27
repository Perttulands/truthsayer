package rules

import "testing"

func TestNoTimeout_ClientWithoutTimeout(t *testing.T) {
	checker := &NoTimeout{}
	src := `package p

import "net/http"

func fetch() {
	client := &http.Client{}
	_ = client
}
`
	findings := runASTCheckerOnSource(t, checker, "client.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestNoTimeout_ClientWithTimeout(t *testing.T) {
	checker := &NoTimeout{}
	src := `package p

import (
	"net/http"
	"time"
)

func fetch() {
	client := &http.Client{Timeout: 30 * time.Second}
	_ = client
}
`
	findings := runASTCheckerOnSource(t, checker, "client.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings with Timeout set, got %d", len(findings))
	}
}

func TestNoTimeout_HttpGet(t *testing.T) {
	checker := &NoTimeout{}
	src := `package p

import "net/http"

func fetch() {
	resp, err := http.Get("https://example.com")
	_ = resp
	_ = err
}
`
	findings := runASTCheckerOnSource(t, checker, "client.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for http.Get, got %d", len(findings))
	}
}

func TestNoTimeout_HttpPost(t *testing.T) {
	checker := &NoTimeout{}
	src := `package p

import "net/http"

func fetch() {
	resp, err := http.Post("https://example.com", "application/json", nil)
	_ = resp
	_ = err
}
`
	findings := runASTCheckerOnSource(t, checker, "client.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for http.Post, got %d", len(findings))
	}
}

func TestNoTimeout_NoHttp(t *testing.T) {
	checker := &NoTimeout{}
	src := `package p

func foo() {
	x := 1
	_ = x
}
`
	findings := runASTCheckerOnSource(t, checker, "main.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-http code, got %d", len(findings))
	}
}
