package law

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/precedent"
)

func TestCandidateValidate_MissingRuleID(t *testing.T) {
	c := Candidate{PatternHash: "p", Decision: precedent.DecisionAllow, Count: 1, Threshold: 1,
		FirstSeen: time.Now(), LastSeen: time.Now()}
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "rule_id") {
		t.Fatalf("expected rule_id error, got %v", err)
	}
}

func TestCandidateValidate_MissingPatternHash(t *testing.T) {
	c := Candidate{RuleID: "r", Decision: precedent.DecisionAllow, Count: 1, Threshold: 1,
		FirstSeen: time.Now(), LastSeen: time.Now()}
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "pattern_hash") {
		t.Fatalf("expected pattern_hash error, got %v", err)
	}
}

func TestCandidateValidate_InvalidDecision(t *testing.T) {
	c := Candidate{RuleID: "r", PatternHash: "p", Decision: "unknown", Count: 1, Threshold: 1,
		FirstSeen: time.Now(), LastSeen: time.Now()}
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid decision") {
		t.Fatalf("expected invalid decision error, got %v", err)
	}
}

func TestCandidateValidate_ZeroCount(t *testing.T) {
	c := Candidate{RuleID: "r", PatternHash: "p", Decision: precedent.DecisionAllow, Count: 0, Threshold: 1,
		FirstSeen: time.Now(), LastSeen: time.Now()}
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "count") {
		t.Fatalf("expected count error, got %v", err)
	}
}

func TestCandidateValidate_ZeroThreshold(t *testing.T) {
	c := Candidate{RuleID: "r", PatternHash: "p", Decision: precedent.DecisionAllow, Count: 1, Threshold: 0,
		FirstSeen: time.Now(), LastSeen: time.Now()}
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "threshold") {
		t.Fatalf("expected threshold error, got %v", err)
	}
}

func TestCandidateValidate_ZeroTimes(t *testing.T) {
	c := Candidate{RuleID: "r", PatternHash: "p", Decision: precedent.DecisionAllow, Count: 1, Threshold: 1}
	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "first_seen") {
		t.Fatalf("expected time error, got %v", err)
	}
}

func TestCandidateValidate_Valid(t *testing.T) {
	c := Candidate{RuleID: "r", PatternHash: "p", Decision: precedent.DecisionDeny, Count: 1, Threshold: 1,
		FirstSeen: time.Now(), LastSeen: time.Now()}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDetectCandidates_DefaultThreshold(t *testing.T) {
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	records := make([]precedent.Precedent, 0, 10)
	for i := 0; i < 10; i++ {
		records = append(records, precedent.Precedent{
			RuleID:      "rule-a",
			PatternHash: "ph",
			Decision:    precedent.DecisionAllow,
			CreatedAt:   now.Add(time.Duration(i) * time.Hour),
		})
	}
	// threshold <= 0 defaults to 10
	candidates := DetectCandidates(records, 0)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate with default threshold, got %d", len(candidates))
	}
}

func TestDetectCandidates_SkipsEmptyRuleID(t *testing.T) {
	now := time.Now()
	records := []precedent.Precedent{
		{RuleID: "", PatternHash: "ph", Decision: precedent.DecisionAllow, CreatedAt: now},
	}
	candidates := DetectCandidates(records, 1)
	if len(candidates) != 0 {
		t.Fatalf("expected 0 candidates for empty rule_id, got %d", len(candidates))
	}
}

func TestDetectCandidates_SkipsEmptyPatternHash(t *testing.T) {
	now := time.Now()
	records := []precedent.Precedent{
		{RuleID: "r", PatternHash: "", Decision: precedent.DecisionAllow, CreatedAt: now},
	}
	candidates := DetectCandidates(records, 1)
	if len(candidates) != 0 {
		t.Fatalf("expected 0 candidates for empty pattern_hash, got %d", len(candidates))
	}
}

func TestDetectCandidates_SkipsInvalidDecision(t *testing.T) {
	now := time.Now()
	records := []precedent.Precedent{
		{RuleID: "r", PatternHash: "ph", Decision: "maybe", CreatedAt: now},
	}
	candidates := DetectCandidates(records, 1)
	if len(candidates) != 0 {
		t.Fatalf("expected 0 candidates for invalid decision, got %d", len(candidates))
	}
}

func TestDetectCandidates_SkipsZeroCreatedAt(t *testing.T) {
	records := []precedent.Precedent{
		{RuleID: "r", PatternHash: "ph", Decision: precedent.DecisionAllow},
	}
	candidates := DetectCandidates(records, 1)
	if len(candidates) != 0 {
		t.Fatalf("expected 0 candidates for zero created_at, got %d", len(candidates))
	}
}

func TestDetectCandidates_BelowThreshold(t *testing.T) {
	now := time.Now()
	records := []precedent.Precedent{
		{RuleID: "r", PatternHash: "ph", Decision: precedent.DecisionAllow, CreatedAt: now},
		{RuleID: "r", PatternHash: "ph", Decision: precedent.DecisionAllow, CreatedAt: now.Add(time.Hour)},
	}
	candidates := DetectCandidates(records, 3)
	if len(candidates) != 0 {
		t.Fatalf("expected 0 candidates below threshold, got %d", len(candidates))
	}
}

func TestDetectCandidates_SortByCountDescThenRuleAsc(t *testing.T) {
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	records := make([]precedent.Precedent, 0)
	// rule-b: 3 occurrences
	for i := 0; i < 3; i++ {
		records = append(records, precedent.Precedent{
			RuleID: "rule-b", PatternHash: "ph", Decision: precedent.DecisionAllow,
			CreatedAt: now.Add(time.Duration(i) * time.Hour),
		})
	}
	// rule-a: 5 occurrences (should sort first by count desc)
	for i := 0; i < 5; i++ {
		records = append(records, precedent.Precedent{
			RuleID: "rule-a", PatternHash: "ph", Decision: precedent.DecisionAllow,
			CreatedAt: now.Add(time.Duration(i) * time.Hour),
		})
	}
	candidates := DetectCandidates(records, 2)
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
	if candidates[0].RuleID != "rule-a" {
		t.Fatalf("expected rule-a first (higher count), got %q", candidates[0].RuleID)
	}
	if candidates[1].RuleID != "rule-b" {
		t.Fatalf("expected rule-b second, got %q", candidates[1].RuleID)
	}
}

func TestDetectCandidates_TracksFirstLastSeen(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	records := []precedent.Precedent{
		{RuleID: "r", PatternHash: "ph", Decision: precedent.DecisionAllow, CreatedAt: t2},
		{RuleID: "r", PatternHash: "ph", Decision: precedent.DecisionAllow, CreatedAt: t1},
		{RuleID: "r", PatternHash: "ph", Decision: precedent.DecisionAllow, CreatedAt: t3},
	}
	candidates := DetectCandidates(records, 2)
	if len(candidates) != 1 {
		t.Fatalf("expected 1, got %d", len(candidates))
	}
	if !candidates[0].FirstSeen.Equal(t1) {
		t.Fatalf("expected first_seen %v, got %v", t1, candidates[0].FirstSeen)
	}
	if !candidates[0].LastSeen.Equal(t3) {
		t.Fatalf("expected last_seen %v, got %v", t3, candidates[0].LastSeen)
	}
}

func TestNewCandidateStore_EmptyPath(t *testing.T) {
	store := NewCandidateStore("")
	if store.path != DefaultCandidatesPath {
		t.Fatalf("expected default path %q, got %q", DefaultCandidatesPath, store.path)
	}
}

func TestCandidateStore_LoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "candidates.json")
	if err := os.WriteFile(path, []byte("  "), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewCandidateStore(path)
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load empty file: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 candidates, got %d", len(got))
	}
}

func TestCandidateStore_LoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "candidates.json")
	if err := os.WriteFile(path, []byte("{broken"), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewCandidateStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestCandidateStore_LoadInvalidEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "candidates.json")
	// Valid JSON but invalid candidate (missing fields)
	if err := os.WriteFile(path, []byte(`[{"rule_id":"","pattern_hash":"","decision":"","count":0,"threshold":0}]`), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewCandidateStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid candidate") {
		t.Fatalf("expected 'invalid candidate' in error, got %v", err)
	}
}

func TestCandidateStore_SaveInvalidEntry(t *testing.T) {
	store := NewCandidateStore(filepath.Join(t.TempDir(), "candidates.json"))
	err := store.Save([]Candidate{{RuleID: ""}})
	if err == nil {
		t.Fatal("expected validation error on save")
	}
}

func TestCandidateStore_SaveSubdir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "dir", "candidates.json")
	store := NewCandidateStore(path)
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	c := Candidate{RuleID: "r", PatternHash: "p", Decision: precedent.DecisionDeny,
		Count: 1, Threshold: 1, FirstSeen: now, LastSeen: now}
	if err := store.Save([]Candidate{c}); err != nil {
		t.Fatalf("save to subdir: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load from subdir: %v", err)
	}
	if len(got) != 1 || got[0].RuleID != "r" {
		t.Fatalf("unexpected load result: %+v", got)
	}
}

func TestSuggestAmendment_Allow(t *testing.T) {
	s := SuggestAmendment(precedent.DecisionAllow)
	if !strings.Contains(s, "exception") {
		t.Fatalf("expected 'exception' in allow suggestion, got %q", s)
	}
}

func TestSuggestAmendment_Deny(t *testing.T) {
	s := SuggestAmendment(precedent.DecisionDeny)
	if !strings.Contains(s, "Tighten") {
		t.Fatalf("expected 'Tighten' in deny suggestion, got %q", s)
	}
}

func TestRenderProposalsMarkdown_NoCandidates(t *testing.T) {
	md := RenderProposalsMarkdown(nil, nil, time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(md, "No proposal candidates") {
		t.Fatalf("expected 'No proposal candidates' in output, got %q", md)
	}
}

func TestRenderProposalsMarkdown_MissingDescription(t *testing.T) {
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	md := RenderProposalsMarkdown([]Candidate{
		{RuleID: "missing-rule", PatternHash: "ph", Decision: precedent.DecisionAllow,
			Count: 5, Threshold: 3, FirstSeen: now, LastSeen: now},
	}, map[string]string{}, now)
	if !strings.Contains(md, "(rule description unavailable)") {
		t.Fatalf("expected unavailable description, got %q", md)
	}
}

func TestRenderProposalsMarkdown_ZeroTime(t *testing.T) {
	before := time.Now().UTC()
	md := RenderProposalsMarkdown(nil, nil, time.Time{})
	_ = before
	if !strings.Contains(md, "Generated:") {
		t.Fatalf("expected Generated timestamp, got %q", md)
	}
}

func TestWriteProposals_Subdir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "proposals.md")
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	err := WriteProposals(path, nil, nil, now)
	if err != nil {
		t.Fatalf("write proposals to subdir: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read proposals: %v", err)
	}
	if !strings.Contains(string(data), "Law Update Proposals") {
		t.Fatalf("expected proposals content, got %q", string(data))
	}
}
