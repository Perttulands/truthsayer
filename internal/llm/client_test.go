package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClaudeClient_RequiresAPIKey(t *testing.T) {
	_, err := NewClaudeClient(ClientOptions{})
	if err == nil {
		t.Fatal("expected error when API key is missing")
	}
}

func testAPIKey(t *testing.T) string {
	t.Helper()
	t.Setenv("TRUTHSAYER_TEST_API_KEY", "token")
	return os.Getenv("TRUTHSAYER_TEST_API_KEY")
}

func TestComplete_SendsClaudeRequestAndParsesResponse(t *testing.T) {
	var gotKey string
	var gotVersion string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"msg_1",
			"stop_reason":"end_turn",
			"content":[{"type":"text","text":"{\"verdict\":\"guilty\"}"}],
			"usage":{"input_tokens":10,"output_tokens":20}
		}`))
	}))
	defer srv.Close()

	client, err := NewClaudeClient(ClientOptions{
		APIKey:             testAPIKey(t),
		BaseURL:            srv.URL,
		MinRequestInterval: 1 * time.Millisecond,
		MaxRetries:         0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	got, err := client.Complete(context.Background(), "sys", "user")
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	if got.Text != `{"verdict":"guilty"}` {
		t.Fatalf("unexpected completion text: %q", got.Text)
	}
	if got.InputTokens != 10 || got.OutputTokens != 20 {
		t.Fatalf("unexpected token usage: in=%d out=%d", got.InputTokens, got.OutputTokens)
	}
	if gotKey != testAPIKey(t) {
		t.Fatalf("missing api key header, got %q", gotKey)
	}
	if gotVersion == "" {
		t.Fatal("expected anthropic-version header")
	}
	if gotBody["model"] == "" {
		t.Fatal("expected model in request body")
	}
}

func TestComplete_RetriesOnServerError(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n == 1 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{
			"id":"msg_ok",
			"stop_reason":"end_turn",
			"content":[{"type":"text","text":"ok"}],
			"usage":{"input_tokens":1,"output_tokens":1}
		}`))
	}))
	defer srv.Close()

	client, err := NewClaudeClient(ClientOptions{
		APIKey:             testAPIKey(t),
		BaseURL:            srv.URL,
		MinRequestInterval: 1 * time.Millisecond,
		MaxRetries:         2,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.sleep = func(time.Duration) {}

	got, err := client.Complete(context.Background(), "", "hello")
	if err != nil {
		t.Fatalf("complete failed after retry: %v", err)
	}
	if got.Text != "ok" {
		t.Fatalf("expected ok response, got %q", got.Text)
	}
	if atomic.LoadInt32(&hits) < 2 {
		t.Fatalf("expected retry attempt, got %d hits", hits)
	}
}

func TestComplete_NonRetryableHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"bad request"}}`))
	}))
	defer srv.Close()

	client, err := NewClaudeClient(ClientOptions{
		APIKey:             testAPIKey(t),
		BaseURL:            srv.URL,
		MinRequestInterval: 1 * time.Millisecond,
		MaxRetries:         2,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.sleep = func(time.Duration) {}

	_, err = client.Complete(context.Background(), "", "hello")
	if err == nil {
		t.Fatal("expected non-retryable error")
	}
	if !strings.Contains(err.Error(), "status=400") {
		t.Fatalf("expected status in error, got %v", err)
	}
}

func TestComplete_RespectsRateLimitInterval(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"id":"msg_ok",
			"stop_reason":"end_turn",
			"content":[{"type":"text","text":"ok"}],
			"usage":{"input_tokens":1,"output_tokens":1}
		}`))
	}))
	defer srv.Close()

	client, err := NewClaudeClient(ClientOptions{
		APIKey:             testAPIKey(t),
		BaseURL:            srv.URL,
		MinRequestInterval: 25 * time.Millisecond,
		MaxRetries:         0,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	start := time.Now()
	_, err = client.Complete(context.Background(), "", "first")
	if err != nil {
		t.Fatalf("first complete: %v", err)
	}
	_, err = client.Complete(context.Background(), "", "second")
	if err != nil {
		t.Fatalf("second complete: %v", err)
	}

	if elapsed := time.Since(start); elapsed < 25*time.Millisecond {
		t.Fatalf("expected at least min interval delay, got %s", elapsed)
	}
}
