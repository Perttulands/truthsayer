package precedent

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreLoadMissingFileReturnsEmpty(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "precedents.json"))
	precedents, err := store.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(precedents) != 0 {
		t.Fatalf("expected empty precedents, got %d", len(precedents))
	}
}

func TestStoreSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	store := NewStore(path)
	now := time.Date(2026, 2, 19, 0, 0, 0, 0, time.UTC)

	input := []Precedent{
		{
			RuleID:        "silent-fallback.empty-error-check",
			ViolationHash: "abc123",
			Decision:      DecisionDeny,
			Rationale:     "This pattern swallowed production errors.",
			CreatedAt:     now,
		},
	}

	if err := store.Save(input); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 precedent, got %d", len(got))
	}
	if got[0].RuleID != input[0].RuleID {
		t.Fatalf("expected rule_id %q, got %q", input[0].RuleID, got[0].RuleID)
	}
	if !got[0].CreatedAt.Equal(now) {
		t.Fatalf("expected created_at %v, got %v", now, got[0].CreatedAt)
	}
}

func TestStoreSaveRejectsInvalidDecision(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "precedents.json"))
	err := store.Save([]Precedent{
		{
			RuleID:        "mock-leakage.test-import-in-src",
			ViolationHash: "deadbeef",
			Decision:      "maybe",
			Rationale:     "Invalid decision should be rejected.",
			CreatedAt:     time.Now().UTC(),
		},
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestStoreAddAndQuery(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "precedents.json"))
	if err := store.Add(Precedent{
		RuleID:        "error-context.http-200-on-error",
		ViolationHash: "hash-1",
		Decision:      DecisionDeny,
		Rationale:     "Hides failures from clients.",
	}); err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	if err := store.Add(Precedent{
		RuleID:        "error-context.http-200-on-error",
		ViolationHash: "hash-1",
		Decision:      DecisionAllow,
		Rationale:     "Legacy endpoint is explicitly whitelisted.",
	}); err != nil {
		t.Fatalf("second add failed: %v", err)
	}

	p, ok, err := store.Query("error-context.http-200-on-error", "hash-1")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if !ok {
		t.Fatal("expected precedent to be found")
	}
	if p.Decision != DecisionAllow {
		t.Fatalf("expected most recent decision allow, got %q", p.Decision)
	}
}

func TestQueryByRule(t *testing.T) {
	now := time.Date(2026, 2, 19, 0, 0, 0, 0, time.UTC)
	precedents := []Precedent{
		{
			RuleID:        "bad-defaults.no-timeout",
			ViolationHash: "a1",
			Decision:      DecisionDeny,
			Rationale:     "No timeout can hang forever.",
			CreatedAt:     now,
		},
		{
			RuleID:        "bad-defaults.no-timeout",
			ViolationHash: "a2",
			Decision:      DecisionAllow,
			Rationale:     "Limited to test code.",
			CreatedAt:     now,
		},
		{
			RuleID:        "mock-leakage.jest-mock-in-src",
			ViolationHash: "b1",
			Decision:      DecisionDeny,
			Rationale:     "Production source uses jest mock.",
			CreatedAt:     now,
		},
	}

	filtered := QueryByRule(precedents, "bad-defaults.no-timeout")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 precedents, got %d", len(filtered))
	}
}
