package precedent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
)

func TestNewStore_EmptyPathUsesDefault(t *testing.T) {
	store := NewStore("")
	if store.Path() != DefaultPath {
		t.Errorf("expected default path %q, got %q", DefaultPath, store.Path())
	}
}

func TestNewStore_WhitespacePathUsesDefault(t *testing.T) {
	store := NewStore("   ")
	if store.Path() != DefaultPath {
		t.Errorf("expected default path %q, got %q", DefaultPath, store.Path())
	}
}

func TestNewStore_CustomPath(t *testing.T) {
	store := NewStore("/custom/path.json")
	if store.Path() != "/custom/path.json" {
		t.Errorf("expected custom path, got %q", store.Path())
	}
}

func TestStorePath(t *testing.T) {
	store := NewStore("test.json")
	if store.Path() != "test.json" {
		t.Errorf("expected test.json, got %q", store.Path())
	}
}

func TestValidate_EmptyRuleID(t *testing.T) {
	p := Precedent{
		ViolationHash: "abc",
		Decision:      DecisionAllow,
		Rationale:     "ok",
		CreatedAt:     time.Now(),
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for empty rule_id")
	}
}

func TestValidate_EmptyViolationHash(t *testing.T) {
	p := Precedent{
		RuleID:    "rule-1",
		Decision:  DecisionAllow,
		Rationale: "ok",
		CreatedAt: time.Now(),
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for empty violation_hash")
	}
}

func TestValidate_NegativeConfidence(t *testing.T) {
	p := Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		Confidence:    -0.1,
		Decision:      DecisionAllow,
		Rationale:     "ok",
		CreatedAt:     time.Now(),
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for negative confidence")
	}
}

func TestValidate_NegativeSeenCount(t *testing.T) {
	p := Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		SeenCount:     -1,
		Decision:      DecisionAllow,
		Rationale:     "ok",
		CreatedAt:     time.Now(),
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for negative seen_count")
	}
}

func TestValidate_EmptyDecision(t *testing.T) {
	p := Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		Decision:      "",
		Rationale:     "ok",
		CreatedAt:     time.Now(),
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for empty decision")
	}
}

func TestValidate_EmptyRationale(t *testing.T) {
	p := Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		Decision:      DecisionAllow,
		Rationale:     "",
		CreatedAt:     time.Now(),
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for empty rationale")
	}
}

func TestValidate_ZeroCreatedAt(t *testing.T) {
	p := Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		Decision:      DecisionAllow,
		Rationale:     "ok",
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for zero created_at")
	}
}

func TestValidate_ValidPrecedent(t *testing.T) {
	p := Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		Decision:      DecisionDeny,
		Rationale:     "bad pattern",
		CreatedAt:     time.Now(),
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected no error for valid precedent, got %v", err)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	if err := os.WriteFile(path, []byte("  \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	records, err := store.Load()
	if err != nil {
		t.Fatalf("expected no error for empty file, got %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected empty list, got %d", len(records))
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadInvalidRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	// Valid JSON but invalid precedent (missing required fields)
	data := `[{"rule_id":"","violation_hash":"abc","decision":"allow","rationale":"ok","created_at":"2026-02-20T00:00:00Z"}]`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for invalid record in file")
	}
}

func TestSave_CreatesParentDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "dir", "precedents.json")
	store := NewStore(path)
	err := store.Save([]Precedent{
		{
			RuleID:        "rule-1",
			ViolationHash: "abc",
			Decision:      DecisionAllow,
			Rationale:     "ok",
			CreatedAt:     time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("expected save to create dirs, got %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist after save: %v", err)
	}
}

func TestAdd_SetsCreatedAtIfZero(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	store := NewStore(path)
	err := store.Add(Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		Decision:      DecisionAllow,
		Rationale:     "ok",
		// CreatedAt is zero
	})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	records, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].CreatedAt.IsZero() {
		t.Fatal("expected created_at to be set automatically")
	}
}

func TestQuery_NoMatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	store := NewStore(path)

	_, ok, err := store.Query("nonexistent", "hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected no match for empty store")
	}
}

func TestNormalizeMatchOptions_ClampsBounds(t *testing.T) {
	opts := normalizeMatchOptions(MatchOptions{MinConfidence: -0.5})
	if opts.MinConfidence != 0 {
		t.Errorf("expected min 0, got %f", opts.MinConfidence)
	}

	opts = normalizeMatchOptions(MatchOptions{MinConfidence: 1.5})
	if opts.MinConfidence != 1 {
		t.Errorf("expected max 1, got %f", opts.MinConfidence)
	}
}

func TestPrecedenceConfidence_ZeroDefaultsToHalf(t *testing.T) {
	p := Precedent{Confidence: 0}
	if got := precedenceConfidence(p); got != 0.5 {
		t.Errorf("expected 0.5, got %f", got)
	}
}

func TestPrecedenceConfidence_NegativeDefaultsToHalf(t *testing.T) {
	p := Precedent{Confidence: -1}
	if got := precedenceConfidence(p); got != 0.5 {
		t.Errorf("expected 0.5, got %f", got)
	}
}

func TestPrecedenceConfidence_PositiveReturnsSame(t *testing.T) {
	p := Precedent{Confidence: 0.75}
	if got := precedenceConfidence(p); got != 0.75 {
		t.Errorf("expected 0.75, got %f", got)
	}
}

func TestMatch_EmptyPrecedents(t *testing.T) {
	f := finding.Finding{Rule: "test", Code: "x"}
	result := Match(nil, f, MatchOptions{})
	if result != nil {
		t.Fatalf("expected nil for empty precedents, got %v", result)
	}
}

func TestMatch_EmptyPatternHash(t *testing.T) {
	f := finding.Finding{Rule: "test", Code: ""}
	result := Match([]Precedent{{RuleID: "test"}}, f, MatchOptions{})
	if result != nil {
		t.Fatalf("expected nil for empty pattern hash, got %v", result)
	}
}

func TestStoreMatch_LoadError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	if err := os.WriteFile(path, []byte("broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	_, err := store.Match(finding.Finding{Rule: "test", Code: "x"}, MatchOptions{})
	if err == nil {
		t.Fatal("expected error from broken file")
	}
}

func TestAddOrUpdateJudgment_LoadError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	if err := os.WriteFile(path, []byte("broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	_, err := store.AddOrUpdateJudgment(Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		PatternHash:   "pat",
		Decision:      DecisionAllow,
		Rationale:     "ok",
	})
	if err == nil {
		t.Fatal("expected error from broken file")
	}
}

func TestAddOrUpdateJudgment_SetsLastSeenIfZero(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	store := NewStore(path)
	result, err := store.AddOrUpdateJudgment(Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		PatternHash:   "pat",
		Decision:      DecisionAllow,
		Rationale:     "ok",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LastSeen.IsZero() {
		t.Fatal("expected last_seen to be set")
	}
}

func TestAddOrUpdateJudgment_InitialConfidenceForNew(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	store := NewStore(path)
	result, err := store.AddOrUpdateJudgment(Precedent{
		RuleID:        "rule-1",
		ViolationHash: "abc",
		PatternHash:   "pat",
		Decision:      DecisionAllow,
		Rationale:     "ok",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Confidence != initialConfidence {
		t.Errorf("expected initial confidence %f, got %f", initialConfidence, result.Confidence)
	}
	if result.SeenCount != 1 {
		t.Errorf("expected seen_count 1, got %d", result.SeenCount)
	}
}
