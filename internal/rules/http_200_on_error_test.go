package rules

import "testing"

func TestHTTP200OnError_WriteHeaderAfterErrCheck(t *testing.T) {
	checker := &HTTP200OnError{}
	src := `package p
import "net/http"

func handler(w http.ResponseWriter, r *http.Request) {
	_, err := doThing()
	if err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
	}
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

func TestHTTP200OnError_JSONEncodeAfterErrCheck(t *testing.T) {
	checker := &HTTP200OnError{}
	src := `package p
import (
	"encoding/json"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	_, err := doThing()
	if err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"ok": "1"})
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestHTTP200OnError_NoFindingWhenReturning(t *testing.T) {
	checker := &HTTP200OnError{}
	src := `package p
import "net/http"

func handler(w http.ResponseWriter, r *http.Request) {
	_, err := doThing()
	if err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
