package debt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewStore_EmptyPath(t *testing.T) {
	store := NewStore("")
	if store.path != DefaultPath {
		t.Fatalf("expected default path %q, got %q", DefaultPath, store.path)
	}
}

func TestNewStore_WhitespacePath(t *testing.T) {
	store := NewStore("   ")
	if store.path != DefaultPath {
		t.Fatalf("expected default path %q, got %q", DefaultPath, store.path)
	}
}

func TestEntryValidate_MissingRuleID(t *testing.T) {
	e := Entry{File: "f", Line: 1, Reasoning: "r", CreatedAt: time.Now()}
	err := e.Validate()
	if err == nil || !strings.Contains(err.Error(), "rule_id") {
		t.Fatalf("expected rule_id error, got %v", err)
	}
}

func TestEntryValidate_MissingFile(t *testing.T) {
	e := Entry{RuleID: "r", Line: 1, Reasoning: "r", CreatedAt: time.Now()}
	err := e.Validate()
	if err == nil || !strings.Contains(err.Error(), "file") {
		t.Fatalf("expected file error, got %v", err)
	}
}

func TestEntryValidate_ZeroLine(t *testing.T) {
	e := Entry{RuleID: "r", File: "f", Line: 0, Reasoning: "r", CreatedAt: time.Now()}
	err := e.Validate()
	if err == nil || !strings.Contains(err.Error(), "line") {
		t.Fatalf("expected line error, got %v", err)
	}
}

func TestEntryValidate_NegativeLine(t *testing.T) {
	e := Entry{RuleID: "r", File: "f", Line: -1, Reasoning: "r", CreatedAt: time.Now()}
	err := e.Validate()
	if err == nil || !strings.Contains(err.Error(), "line") {
		t.Fatalf("expected line error, got %v", err)
	}
}

func TestEntryValidate_MissingReasoning(t *testing.T) {
	e := Entry{RuleID: "r", File: "f", Line: 1, Reasoning: "", CreatedAt: time.Now()}
	err := e.Validate()
	if err == nil || !strings.Contains(err.Error(), "reasoning") {
		t.Fatalf("expected reasoning error, got %v", err)
	}
}

func TestEntryValidate_ZeroCreatedAt(t *testing.T) {
	e := Entry{RuleID: "r", File: "f", Line: 1, Reasoning: "r"}
	err := e.Validate()
	if err == nil || !strings.Contains(err.Error(), "created_at") {
		t.Fatalf("expected created_at error, got %v", err)
	}
}

func TestEntryValidate_Valid(t *testing.T) {
	e := Entry{RuleID: "r", File: "f", Line: 1, Reasoning: "r", CreatedAt: time.Now()}
	if err := e.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStore_LoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "debt.json")
	if err := os.WriteFile(path, []byte("  \n"), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	entries, err := store.Load()
	if err != nil {
		t.Fatalf("load empty file: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestStore_LoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "debt.json")
	if err := os.WriteFile(path, []byte("{broken json"), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestStore_LoadInvalidEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "debt.json")
	// Valid JSON array but invalid entry (missing required fields)
	if err := os.WriteFile(path, []byte(`[{"rule_id":"","file":"","line":0,"reasoning":"","created_at":"0001-01-01T00:00:00Z"}]`), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid entry") {
		t.Fatalf("expected 'invalid entry' in error, got %v", err)
	}
}

func TestStore_SaveInvalidEntry(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "debt.json"))
	err := store.Save([]Entry{{RuleID: "", File: "f", Line: 1, Reasoning: "r", CreatedAt: time.Now()}})
	if err == nil {
		t.Fatal("expected validation error on save")
	}
}

func TestStore_SaveSubdir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "dir", "debt.json")
	store := NewStore(path)
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	err := store.Save([]Entry{{RuleID: "r", File: "f", Line: 1, Reasoning: "r", CreatedAt: now}})
	if err != nil {
		t.Fatalf("save to subdir failed: %v", err)
	}
	entries, err := store.Load()
	if err != nil {
		t.Fatalf("load from subdir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestStore_Add_DefaultsTime(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "debt.json"))
	before := time.Now().UTC()
	err := store.Add(Entry{
		RuleID:    "r",
		File:      "f",
		Line:      1,
		Reasoning: "r",
		// CreatedAt intentionally zero
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
	after := time.Now().UTC()
	if entries[0].CreatedAt.Before(before) || entries[0].CreatedAt.After(after) {
		t.Fatalf("expected CreatedAt near now, got %v", entries[0].CreatedAt)
	}
}

func TestStore_Add_LoadError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "debt.json")
	// Write corrupt data so Load fails
	if err := os.WriteFile(path, []byte("{broken"), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	err := store.Add(Entry{RuleID: "r", File: "f", Line: 1, Reasoning: "r", CreatedAt: time.Now()})
	if err == nil {
		t.Fatal("expected error when load fails during add")
	}
}

func TestStore_SaveAndLoadRoundtrip(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "debt.json"))
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	entries := []Entry{
		{RuleID: "r1", File: "a.go", Line: 1, Code: "x", Message: "m", Reasoning: "reason1", CreatedAt: now},
		{RuleID: "r2", File: "b.go", Line: 2, Code: "y", Message: "n", Reasoning: "reason2", CreatedAt: now},
	}
	if err := store.Save(entries); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].RuleID != "r1" || got[1].RuleID != "r2" {
		t.Fatalf("unexpected entries: %+v", got)
	}
}
