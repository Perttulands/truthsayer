package precedent

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
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

func TestStoreSaveRejectsInvalidConfidence(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "precedents.json"))
	err := store.Save([]Precedent{
		{
			RuleID:        "mock-leakage.test-import-in-src",
			ViolationHash: "deadbeef",
			PatternHash:   "beefdead",
			Decision:      DecisionDeny,
			Rationale:     "Invalid confidence should be rejected.",
			Confidence:    1.2,
			CreatedAt:     time.Now().UTC(),
		},
	})
	if err == nil {
		t.Fatal("expected confidence validation error, got nil")
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

func TestMatch_FiltersByRuleAndPatternHash(t *testing.T) {
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	targetFinding := finding.Finding{
		Rule: "silent-fallback.empty-error-check",
		Code: "if err != nil { return nil }",
	}
	targetPattern := HashFindingPattern(targetFinding)

	precedents := []Precedent{
		{
			RuleID:        targetFinding.Rule,
			ViolationHash: "v1",
			PatternHash:   targetPattern,
			Decision:      DecisionDeny,
			Rationale:     "still guilty",
			Confidence:    0.95,
			SeenCount:     4,
			CreatedAt:     now,
		},
		{
			RuleID:        targetFinding.Rule,
			ViolationHash: "v2",
			PatternHash:   HashPattern(targetFinding.Rule, "if other != nil { return nil }"), // same normalized pattern
			Decision:      DecisionDeny,
			Rationale:     "historical",
			Confidence:    0.80,
			SeenCount:     8,
			CreatedAt:     now.Add(-time.Hour),
		},
		{
			RuleID:        "bad-defaults.no-timeout",
			ViolationHash: "v3",
			PatternHash:   targetPattern,
			Decision:      DecisionDeny,
			Rationale:     "other rule",
			Confidence:    1.0,
			SeenCount:     12,
			CreatedAt:     now,
		},
	}

	matches := Match(precedents, targetFinding, MatchOptions{})
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Confidence < matches[1].Confidence {
		t.Fatalf("expected confidence-sorted matches, got %.2f then %.2f", matches[0].Confidence, matches[1].Confidence)
	}
}

func TestMatch_RespectsConfidenceThresholdAndLimit(t *testing.T) {
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	f := finding.Finding{
		Rule: "silent-fallback.hidden-failure-bash",
		Code: "cmd || true",
	}
	pattern := HashFindingPattern(f)
	precedents := []Precedent{
		{
			RuleID:        f.Rule,
			ViolationHash: "a",
			PatternHash:   pattern,
			Decision:      DecisionAllow,
			Rationale:     "cleanup context",
			Confidence:    0.91,
			SeenCount:     11,
			CreatedAt:     now,
		},
		{
			RuleID:        f.Rule,
			ViolationHash: "b",
			PatternHash:   pattern,
			Decision:      DecisionAllow,
			Rationale:     "old",
			Confidence:    0.40,
			SeenCount:     2,
			CreatedAt:     now.Add(-time.Hour),
		},
	}

	matches := Match(precedents, f, MatchOptions{MinConfidence: 0.9, Limit: 1})
	if len(matches) != 1 {
		t.Fatalf("expected 1 high-confidence match, got %d", len(matches))
	}
	if matches[0].ViolationHash != "a" {
		t.Fatalf("expected highest confidence match 'a', got %q", matches[0].ViolationHash)
	}
}

func TestStoreMatch_LoadsFromDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "precedents.json")
	store := NewStore(path)
	f := finding.Finding{
		Rule: "silent-fallback.hidden-failure-bash",
		Code: "cmd || true",
	}
	rec := Precedent{
		RuleID:        f.Rule,
		ViolationHash: "x1",
		PatternHash:   HashFindingPattern(f),
		Decision:      DecisionAllow,
		Rationale:     "trusted precedent",
		Confidence:    0.93,
		SeenCount:     7,
		CreatedAt:     time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	}
	if err := store.Save([]Precedent{rec}); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	matches, err := store.Match(f, MatchOptions{MinConfidence: 0.9})
	if err != nil {
		t.Fatalf("store match failed: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].ViolationHash != rec.ViolationHash {
		t.Fatalf("expected %q, got %q", rec.ViolationHash, matches[0].ViolationHash)
	}
}

func BenchmarkMatch(b *testing.B) {
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	f := finding.Finding{
		Rule: "silent-fallback.empty-error-check",
		Code: "if err != nil { return nil }",
	}
	targetPattern := HashFindingPattern(f)

	precedents := make([]Precedent, 0, 2000)
	for i := 0; i < 2000; i++ {
		rule := "silent-fallback.empty-error-check"
		if i%3 == 0 {
			rule = "bad-defaults.no-timeout"
		}
		pattern := targetPattern
		if i%5 == 0 {
			pattern = HashPattern(rule, "x = 1")
		}
		precedents = append(precedents, Precedent{
			RuleID:        rule,
			ViolationHash: "v" + time.Unix(int64(i), 0).UTC().Format(time.RFC3339),
			PatternHash:   pattern,
			Decision:      DecisionDeny,
			Rationale:     "benchmark",
			Confidence:    0.5 + float64(i%50)/100,
			SeenCount:     i % 20,
			CreatedAt:     now.Add(-time.Duration(i) * time.Minute),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Match(precedents, f, MatchOptions{MinConfidence: 0.6, Limit: 10})
	}
}
