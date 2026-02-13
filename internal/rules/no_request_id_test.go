package rules

import "testing"

func TestNoRequestID_FindsHandlerWithoutExtraction(t *testing.T) {
	checker := &NoRequestID{}
	src := `package p
import "net/http"

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Fatalf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestNoRequestID_HeaderExtraction(t *testing.T) {
	checker := &NoRequestID{}
	src := `package p
import "net/http"

func handler(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	_ = reqID
	w.WriteHeader(http.StatusOK)
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestNoRequestID_ContextExtraction(t *testing.T) {
	checker := &NoRequestID{}
	src := `package p
import "net/http"

func handler(w http.ResponseWriter, r *http.Request) {
	_ = r.Context().Value("request_id")
	w.WriteHeader(http.StatusOK)
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
