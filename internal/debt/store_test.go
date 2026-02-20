package debt

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStore_LoadMissingReturnsEmpty(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), DefaultPath))
	entries, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(entries))
	}
}

func TestStore_AddAndLoad(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), DefaultPath))
	err := store.Add(Entry{
		RuleID:    "silent-fallback.hidden-failure-bash",
		File:      "script.sh",
		Line:      3,
		Code:      "cmd || true",
		Message:   "hidden failure",
		Reasoning: "requires follow-up",
		CreatedAt: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	entries, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].RuleID != "silent-fallback.hidden-failure-bash" {
		t.Fatalf("unexpected rule id: %q", entries[0].RuleID)
	}
}

func TestEntryValidate(t *testing.T) {
	if err := (Entry{}).Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}
