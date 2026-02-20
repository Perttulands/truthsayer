package judge

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/llm"
	"github.com/perttulands/truthsayer/internal/precedent"
)

type fakeCompleter struct {
	response llm.Completion
	err      error
}

func (f fakeCompleter) Complete(context.Context, string, string) (llm.Completion, error) {
	return f.response, f.err
}

func TestJudgeFinding_Success(t *testing.T) {
	j, err := NewLLMJudge(fakeCompleter{
		response: llm.Completion{
			Text: `{"verdict":"not_guilty","reasoning":"cleanup trap context","confidence":0.93,"precedent_decision":"allow","precedent_rationale":"intentional cleanup fallback"}`,
			InputTokens:  42,
			OutputTokens: 17,
		},
	})
	if err != nil {
		t.Fatalf("new judge: %v", err)
	}

	v, err := j.JudgeFinding(context.Background(), PromptInput{
		Finding: finding.Finding{
			Rule:    "silent-fallback.hidden-failure-bash",
			Message: "|| true hidden failure",
			Code:    "cmd || true",
			File:    "script.sh",
			Line:    3,
		},
		RuleDescription: "Detect hidden failure suppression in bash.",
	})
	if err != nil {
		t.Fatalf("judge finding failed: %v", err)
	}
	if v.Verdict != VerdictNotGuilty {
		t.Fatalf("expected not_guilty, got %q", v.Verdict)
	}
	if v.PrecedentDecision != precedent.DecisionAllow {
		t.Fatalf("expected allow precedent decision, got %q", v.PrecedentDecision)
	}
	if v.Source != "llm" {
		t.Fatalf("expected source llm, got %q", v.Source)
	}
	if v.InputTokens != 42 || v.OutputTokens != 17 {
		t.Fatalf("unexpected token values in verdict: in=%d out=%d", v.InputTokens, v.OutputTokens)
	}
}

func TestParseVerdict_ExtractsJSONFromWrappedText(t *testing.T) {
	raw := "Here is the result:\n```json\n{\"verdict\":\"guilty\",\"reasoning\":\"returns nil on error\",\"confidence\":0.88,\"precedent_decision\":\"deny\",\"precedent_rationale\":\"error swallowed\"}\n```"
	got, err := ParseVerdict(raw)
	if err != nil {
		t.Fatalf("parse verdict failed: %v", err)
	}
	if got.Verdict != VerdictGuilty {
		t.Fatalf("expected guilty, got %q", got.Verdict)
	}
}

func TestParseVerdict_RejectsMalformedResponse(t *testing.T) {
	_, err := ParseVerdict("not json")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestJudgeFinding_HandlesMalformedJSON(t *testing.T) {
	j, err := NewLLMJudge(fakeCompleter{
		response: llm.Completion{
			Text: "```json\n{\"verdict\":\"maybe\"}\n```",
		},
	})
	if err != nil {
		t.Fatalf("new judge: %v", err)
	}
	_, err = j.JudgeFinding(context.Background(), PromptInput{
		Finding: finding.Finding{
			Rule:    "silent-fallback.empty-error-check",
			Message: "bad return nil",
			Code:    "if err != nil { return nil }",
		},
	})
	if err == nil {
		t.Fatal("expected malformed verdict error")
	}
	if !strings.Contains(err.Error(), "parse verdict") {
		t.Fatalf("expected parse verdict context in error, got %v", err)
	}
}

func TestVerdictToPrecedent(t *testing.T) {
	v := Verdict{
		Verdict:            VerdictAdvisory,
		Reasoning:          "could be intentional but risky",
		Confidence:         0.72,
		PrecedentDecision:  precedent.DecisionDeny,
		PrecedentRationale: "needs human follow-up",
	}
	f := finding.Finding{
		Rule: "trace-gaps.error-path-no-log",
		File: "service.go",
		Line: 28,
		Code: "if err != nil { return err }",
	}
	at := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)

	rec := v.ToPrecedent(f, at)
	if rec.RuleID != f.Rule {
		t.Fatalf("expected rule_id %q, got %q", f.Rule, rec.RuleID)
	}
	if rec.PatternHash == "" || rec.ViolationHash == "" {
		t.Fatal("expected hashes to be populated")
	}
	if rec.Confidence != v.Confidence {
		t.Fatalf("expected confidence %.2f, got %.2f", v.Confidence, rec.Confidence)
	}
}
