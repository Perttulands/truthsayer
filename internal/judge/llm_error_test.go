package judge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/llm"
	"github.com/perttulands/truthsayer/internal/precedent"
)

func TestNewLLMJudge_NilCompleter(t *testing.T) {
	_, err := NewLLMJudge(nil)
	if err == nil {
		t.Fatal("expected error for nil completer")
	}
}

func TestJudgeFinding_CompletionError(t *testing.T) {
	j, err := NewLLMJudge(fakeCompleter{err: errors.New("connection refused")})
	if err != nil {
		t.Fatalf("new judge: %v", err)
	}
	_, err = j.JudgeFinding(context.Background(), PromptInput{
		Finding: finding.Finding{
			Rule:    "test-rule",
			Message: "test message",
			Code:    "x",
		},
	})
	if err == nil {
		t.Fatal("expected error from LLM failure")
	}
	if got := err.Error(); !contains(got, "llm completion") {
		t.Fatalf("expected 'llm completion' in error, got %q", got)
	}
}

func TestJudgeFinding_BuildPromptError(t *testing.T) {
	j, err := NewLLMJudge(fakeCompleter{})
	if err != nil {
		t.Fatalf("new judge: %v", err)
	}
	// Empty rule triggers BuildPrompt validation error
	_, err = j.JudgeFinding(context.Background(), PromptInput{
		Finding: finding.Finding{Rule: "", Message: ""},
	})
	if err == nil {
		t.Fatal("expected build prompt error")
	}
	if got := err.Error(); !contains(got, "build prompt") {
		t.Fatalf("expected 'build prompt' in error, got %q", got)
	}
}

func TestParseVerdict_EmptyResponse(t *testing.T) {
	_, err := ParseVerdict("")
	if err == nil {
		t.Fatal("expected error for empty response")
	}
	if got := err.Error(); !contains(got, "empty response") {
		t.Fatalf("expected 'empty response' in error, got %q", got)
	}
}

func TestParseVerdict_NoJSONObject(t *testing.T) {
	_, err := ParseVerdict("just some text without braces")
	if err == nil {
		t.Fatal("expected error for missing JSON object")
	}
	if got := err.Error(); !contains(got, "no json object found") {
		t.Fatalf("expected 'no json object found' in error, got %q", got)
	}
}

func TestParseVerdict_UnterminatedJSON(t *testing.T) {
	_, err := ParseVerdict(`{"verdict":"guilty","reasoning":"bad`)
	if err == nil {
		t.Fatal("expected error for unterminated JSON")
	}
}

func TestParseVerdict_InvalidVerdictValue(t *testing.T) {
	_, err := ParseVerdict(`{"verdict":"maybe","reasoning":"dunno","confidence":0.5,"precedent_decision":"allow","precedent_rationale":"reason"}`)
	if err == nil {
		t.Fatal("expected error for invalid verdict value")
	}
	if got := err.Error(); !contains(got, "invalid verdict") {
		t.Fatalf("expected 'invalid verdict' in error, got %q", got)
	}
}

func TestParseVerdict_EmptyReasoning(t *testing.T) {
	_, err := ParseVerdict(`{"verdict":"guilty","reasoning":"","confidence":0.5,"precedent_decision":"allow","precedent_rationale":"reason"}`)
	if err == nil {
		t.Fatal("expected error for empty reasoning")
	}
	if got := err.Error(); !contains(got, "reasoning is required") {
		t.Fatalf("expected 'reasoning is required' in error, got %q", got)
	}
}

func TestParseVerdict_ConfidenceOutOfRange(t *testing.T) {
	_, err := ParseVerdict(`{"verdict":"guilty","reasoning":"bad code","confidence":1.5,"precedent_decision":"deny","precedent_rationale":"reason"}`)
	if err == nil {
		t.Fatal("expected error for confidence out of range")
	}
	if got := err.Error(); !contains(got, "confidence out of range") {
		t.Fatalf("expected 'confidence out of range' in error, got %q", got)
	}
}

func TestParseVerdict_NegativeConfidence(t *testing.T) {
	_, err := ParseVerdict(`{"verdict":"guilty","reasoning":"bad code","confidence":-0.1,"precedent_decision":"deny","precedent_rationale":"reason"}`)
	if err == nil {
		t.Fatal("expected error for negative confidence")
	}
}

func TestParseVerdict_InvalidPrecedentDecision(t *testing.T) {
	_, err := ParseVerdict(`{"verdict":"guilty","reasoning":"bad code","confidence":0.8,"precedent_decision":"unknown","precedent_rationale":"reason"}`)
	if err == nil {
		t.Fatal("expected error for invalid precedent_decision")
	}
	if got := err.Error(); !contains(got, "invalid precedent_decision") {
		t.Fatalf("expected 'invalid precedent_decision' in error, got %q", got)
	}
}

func TestParseVerdict_EmptyPrecedentRationale(t *testing.T) {
	_, err := ParseVerdict(`{"verdict":"guilty","reasoning":"bad code","confidence":0.8,"precedent_decision":"deny","precedent_rationale":""}`)
	if err == nil {
		t.Fatal("expected error for empty precedent_rationale")
	}
	if got := err.Error(); !contains(got, "precedent_rationale is required") {
		t.Fatalf("expected 'precedent_rationale is required' in error, got %q", got)
	}
}

func TestParseVerdict_AdvisoryVerdict(t *testing.T) {
	v, err := ParseVerdict(`{"verdict":"advisory","reasoning":"risky but intentional","confidence":0.7,"precedent_decision":"deny","precedent_rationale":"needs review"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Verdict != VerdictAdvisory {
		t.Fatalf("expected advisory, got %q", v.Verdict)
	}
}

func TestParseVerdict_EscapedStringsInJSON(t *testing.T) {
	v, err := ParseVerdict(`{"verdict":"guilty","reasoning":"contains \"quoted\" text","confidence":0.85,"precedent_decision":"deny","precedent_rationale":"has \\backslash"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Verdict != VerdictGuilty {
		t.Fatalf("expected guilty, got %q", v.Verdict)
	}
}

func TestParseVerdict_NestedJSON(t *testing.T) {
	// extractJSONObject must handle nested braces
	raw := `Some preamble {"verdict":"not_guilty","reasoning":"nested {object} ok","confidence":0.9,"precedent_decision":"allow","precedent_rationale":"safe pattern"}`
	v, err := ParseVerdict(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Verdict != VerdictNotGuilty {
		t.Fatalf("expected not_guilty, got %q", v.Verdict)
	}
}

func TestAsPrecedent(t *testing.T) {
	rec := PrecedentRecord{
		RuleID:        "test-rule",
		ViolationHash: "vh",
		PatternHash:   "ph",
		Decision:      precedent.DecisionAllow,
		Rationale:     "known pattern",
		Confidence:    0.85,
		SeenCount:     3,
		CreatedAt:     time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	}
	p := rec.AsPrecedent()
	if p.RuleID != rec.RuleID {
		t.Fatalf("expected rule_id %q, got %q", rec.RuleID, p.RuleID)
	}
	if p.ViolationHash != rec.ViolationHash {
		t.Fatalf("expected violation_hash %q, got %q", rec.ViolationHash, p.ViolationHash)
	}
	if p.PatternHash != rec.PatternHash {
		t.Fatalf("expected pattern_hash %q, got %q", rec.PatternHash, p.PatternHash)
	}
	if p.Decision != rec.Decision {
		t.Fatalf("expected decision %q, got %q", rec.Decision, p.Decision)
	}
	if p.Confidence != rec.Confidence {
		t.Fatalf("expected confidence %.2f, got %.2f", rec.Confidence, p.Confidence)
	}
	if p.SeenCount != rec.SeenCount {
		t.Fatalf("expected seen_count %d, got %d", rec.SeenCount, p.SeenCount)
	}
}

func TestToPrecedent_GuiltyDefaultsDeny(t *testing.T) {
	v := Verdict{
		Verdict:            VerdictGuilty,
		Reasoning:          "clearly bad",
		Confidence:         0.95,
		PrecedentDecision:  "",
		PrecedentRationale: "swallowed error",
	}
	f := finding.Finding{
		Rule: "swallowed-error",
		File: "handler.go",
		Line: 10,
		Code: "_ = err",
	}
	rec := v.ToPrecedent(f, time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC))
	if rec.Decision != precedent.DecisionDeny {
		t.Fatalf("expected deny for guilty verdict, got %q", rec.Decision)
	}
}

func TestToPrecedent_NotGuiltyDefaultsAllow(t *testing.T) {
	v := Verdict{
		Verdict:            VerdictNotGuilty,
		Reasoning:          "intentional",
		Confidence:         0.9,
		PrecedentDecision:  "",
		PrecedentRationale: "safe cleanup",
	}
	f := finding.Finding{
		Rule: "cleanup-pattern",
		File: "cleanup.go",
		Line: 5,
		Code: "defer f.Close()",
	}
	rec := v.ToPrecedent(f, time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC))
	if rec.Decision != precedent.DecisionAllow {
		t.Fatalf("expected allow for not_guilty verdict, got %q", rec.Decision)
	}
}

func TestToPrecedent_ZeroTimeDefaultsToNow(t *testing.T) {
	v := Verdict{
		Verdict:            VerdictGuilty,
		Reasoning:          "bad",
		Confidence:         0.8,
		PrecedentDecision:  precedent.DecisionDeny,
		PrecedentRationale: "swallowed",
	}
	f := finding.Finding{Rule: "r", File: "f.go", Line: 1, Code: "x"}
	before := time.Now().UTC()
	rec := v.ToPrecedent(f, time.Time{})
	after := time.Now().UTC()
	if rec.CreatedAt.Before(before) || rec.CreatedAt.After(after) {
		t.Fatalf("expected CreatedAt near now, got %v", rec.CreatedAt)
	}
}

func TestToPrecedent_PreservesExplicitDecision(t *testing.T) {
	v := Verdict{
		Verdict:            VerdictNotGuilty,
		Reasoning:          "safe",
		Confidence:         0.8,
		PrecedentDecision:  precedent.DecisionDeny,
		PrecedentRationale: "override",
	}
	f := finding.Finding{Rule: "r", File: "f.go", Line: 1, Code: "x"}
	rec := v.ToPrecedent(f, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if rec.Decision != precedent.DecisionDeny {
		t.Fatalf("expected explicit deny to be preserved, got %q", rec.Decision)
	}
}

func TestJudgeFinding_SetsSourceAndTokens(t *testing.T) {
	j, err := NewLLMJudge(fakeCompleter{
		response: llm.Completion{
			Text:         `{"verdict":"guilty","reasoning":"error swallowed","confidence":0.9,"precedent_decision":"deny","precedent_rationale":"known bad"}`,
			InputTokens:  100,
			OutputTokens: 50,
		},
	})
	if err != nil {
		t.Fatalf("new judge: %v", err)
	}
	v, err := j.JudgeFinding(context.Background(), PromptInput{
		Finding: finding.Finding{Rule: "r", Message: "m", Code: "c"},
	})
	if err != nil {
		t.Fatalf("judge finding: %v", err)
	}
	if v.Source != "llm" {
		t.Fatalf("expected source 'llm', got %q", v.Source)
	}
	if v.InputTokens != 100 || v.OutputTokens != 50 {
		t.Fatalf("expected tokens 100/50, got %d/%d", v.InputTokens, v.OutputTokens)
	}
}

func TestBuildPrompt_MissingMessage(t *testing.T) {
	_, _, err := BuildPrompt(PromptInput{
		Finding: finding.Finding{Rule: "r", Message: ""},
	})
	if err == nil {
		t.Fatal("expected error for missing message")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
