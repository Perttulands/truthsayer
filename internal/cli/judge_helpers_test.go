package cli

import (
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/precedent"
)

func TestFilterMatchesByConfidence_EmptyList(t *testing.T) {
	result := filterMatchesByConfidence(nil, 0.5)
	if result != nil {
		t.Fatalf("expected nil for empty list, got %v", result)
	}
}

func TestFilterMatchesByConfidence_ZeroMinReturnsAll(t *testing.T) {
	matches := []precedent.Precedent{
		{Confidence: 0.1, RuleID: "r1", ViolationHash: "v1", Decision: precedent.DecisionAllow, Rationale: "ok", CreatedAt: time.Now()},
		{Confidence: 0.9, RuleID: "r2", ViolationHash: "v2", Decision: precedent.DecisionAllow, Rationale: "ok", CreatedAt: time.Now()},
	}
	result := filterMatchesByConfidence(matches, 0)
	if len(result) != 2 {
		t.Fatalf("expected all 2 matches returned for min=0, got %d", len(result))
	}
}

func TestFilterMatchesByConfidence_FiltersLowConfidence(t *testing.T) {
	matches := []precedent.Precedent{
		{Confidence: 0.3, RuleID: "r1", ViolationHash: "v1", Decision: precedent.DecisionAllow, Rationale: "ok", CreatedAt: time.Now()},
		{Confidence: 0.8, RuleID: "r2", ViolationHash: "v2", Decision: precedent.DecisionAllow, Rationale: "ok", CreatedAt: time.Now()},
		{Confidence: 0.95, RuleID: "r3", ViolationHash: "v3", Decision: precedent.DecisionDeny, Rationale: "bad", CreatedAt: time.Now()},
	}
	result := filterMatchesByConfidence(matches, 0.8)
	if len(result) != 2 {
		t.Fatalf("expected 2 matches above 0.8, got %d", len(result))
	}
}

func TestFilterMatchesByConfidence_ZeroConfidenceDefaultsToHalf(t *testing.T) {
	matches := []precedent.Precedent{
		{Confidence: 0, RuleID: "r1", ViolationHash: "v1", Decision: precedent.DecisionAllow, Rationale: "ok", CreatedAt: time.Now()},
	}
	// effectiveConfidence returns 0.5 for zero confidence
	result := filterMatchesByConfidence(matches, 0.5)
	if len(result) != 1 {
		t.Fatalf("expected 1 match (zero confidence defaults to 0.5), got %d", len(result))
	}
}

func TestStrongestMatch_Empty(t *testing.T) {
	_, ok := strongestMatch(nil)
	if ok {
		t.Fatal("expected false for empty list")
	}
}

func TestStrongestMatch_ReturnsFirst(t *testing.T) {
	matches := []precedent.Precedent{
		{Confidence: 0.95, ViolationHash: "first"},
		{Confidence: 0.80, ViolationHash: "second"},
	}
	p, ok := strongestMatch(matches)
	if !ok {
		t.Fatal("expected true")
	}
	if p.ViolationHash != "first" {
		t.Errorf("expected first match, got %q", p.ViolationHash)
	}
}

func TestEffectiveConfidence_ZeroDefaultsToHalf(t *testing.T) {
	p := precedent.Precedent{Confidence: 0}
	if got := effectiveConfidence(p); got != 0.5 {
		t.Errorf("expected 0.5 for zero confidence, got %f", got)
	}
}

func TestEffectiveConfidence_NegativeDefaultsToHalf(t *testing.T) {
	p := precedent.Precedent{Confidence: -0.1}
	if got := effectiveConfidence(p); got != 0.5 {
		t.Errorf("expected 0.5 for negative confidence, got %f", got)
	}
}

func TestEffectiveConfidence_PositiveReturnsSame(t *testing.T) {
	p := precedent.Precedent{Confidence: 0.87}
	if got := effectiveConfidence(p); got != 0.87 {
		t.Errorf("expected 0.87, got %f", got)
	}
}

func TestParseJudgeOptions_MissingInputPath(t *testing.T) {
	_, err := parseJudgeOptions(nil)
	if err == nil {
		t.Fatal("expected error for missing input path")
	}
}

func TestParseJudgeOptions_BadBudgetValue(t *testing.T) {
	_, err := parseJudgeOptions([]string{"--budget", "abc", "f.json"})
	if err == nil {
		t.Fatal("expected error for invalid budget value")
	}
}

func TestParseJudgeOptions_BadLawThreshold(t *testing.T) {
	_, err := parseJudgeOptions([]string{"--law-threshold", "xyz", "f.json"})
	if err == nil {
		t.Fatal("expected error for invalid law-threshold value")
	}
}

func TestParseJudgeOptions_BadMinConfidence(t *testing.T) {
	_, err := parseJudgeOptions([]string{"--min-confidence", "bad", "f.json"})
	if err == nil {
		t.Fatal("expected error for invalid min-confidence value")
	}
}

func TestParseJudgeOptions_BadAutoApplyThreshold(t *testing.T) {
	_, err := parseJudgeOptions([]string{"--auto-apply-threshold", "bad", "f.json"})
	if err == nil {
		t.Fatal("expected error for invalid auto-apply-threshold value")
	}
}
