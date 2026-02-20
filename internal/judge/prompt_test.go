package judge

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/precedent"
)

func TestBuildPrompt_IncludesFindingContextAndSchema(t *testing.T) {
	system, user, err := BuildPrompt(PromptInput{
		Finding: finding.Finding{
			Rule:     "silent-fallback.hidden-failure-bash",
			Severity: finding.SeverityError,
			File:     "scripts/deploy.sh",
			Line:     19,
			Code:     "cmd || true",
			Context:  ">> 19 | cmd || true",
			Message:  "'|| true' silently swallows command failure",
		},
		RuleDescription: "Detect hidden failure suppression in bash.",
		Precedents: []precedent.Precedent{
			{
				RuleID:      "silent-fallback.hidden-failure-bash",
				PatternHash: "abc123",
				Decision:    precedent.DecisionAllow,
				Rationale:   "trap cleanup context",
				Confidence:  0.94,
				SeenCount:   7,
				CreatedAt:   time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
			},
		},
	})
	if err != nil {
		t.Fatalf("build prompt failed: %v", err)
	}
	if !strings.Contains(system, "Return exactly one JSON object") {
		t.Fatalf("system prompt missing strict JSON instruction:\n%s", system)
	}
	if !strings.Contains(user, `"response_schema"`) {
		t.Fatalf("user prompt missing response schema:\n%s", user)
	}
	if !strings.Contains(user, `"context": "\u003e\u003e 19 | cmd || true"`) {
		t.Fatalf("user prompt missing finding context:\n%s", user)
	}
	if !strings.Contains(user, `"verdict"`) || !strings.Contains(user, `"precedent_decision"`) {
		t.Fatalf("user prompt missing required fields:\n%s", user)
	}
}

func TestBuildPrompt_HasParseablePayload(t *testing.T) {
	_, user, err := BuildPrompt(PromptInput{
		Finding: finding.Finding{
			Rule:    "error-context.http-200-on-error",
			Message: "Returning 200 from catch hides failures",
			Code:    "return 200",
		},
		RuleDescription: "Detect HTTP success status on error paths.",
	})
	if err != nil {
		t.Fatalf("build prompt failed: %v", err)
	}
	idx := strings.Index(user, "{")
	if idx < 0 {
		t.Fatalf("expected JSON payload in prompt:\n%s", user)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(user[idx:]), &payload); err != nil {
		t.Fatalf("prompt payload should be valid JSON: %v\npayload:\n%s", err, user[idx:])
	}
	if _, ok := payload["response_schema"]; !ok {
		t.Fatal("expected response_schema in prompt payload")
	}
}

func TestBuildPrompt_ValidatesRequiredFindingFields(t *testing.T) {
	_, _, err := BuildPrompt(PromptInput{
		Finding: finding.Finding{
			Rule: "",
		},
	})
	if err == nil {
		t.Fatal("expected validation error for missing finding rule")
	}
}
