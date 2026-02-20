package judge

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/precedent"
)

// PromptInput contains data used to build a judgment prompt.
type PromptInput struct {
	Finding         finding.Finding
	RuleDescription string
	Precedents      []precedent.Precedent
}

type promptPayload struct {
	GeneratedAtUTC  string                  `json:"generated_at_utc"`
	Finding         finding.Finding         `json:"finding"`
	RuleDescription string                  `json:"rule_description"`
	Precedents      []promptPrecedentRecord `json:"precedents"`
	ResponseSchema  map[string]any          `json:"response_schema"`
}

type promptPrecedentRecord struct {
	RuleID      string  `json:"rule_id"`
	Decision    string  `json:"decision"`
	Rationale   string  `json:"rationale"`
	Confidence  float64 `json:"confidence"`
	SeenCount   int     `json:"seen_count"`
	PatternHash string  `json:"pattern_hash,omitempty"`
}

// BuildPrompt returns system and user prompts for a judgment call.
func BuildPrompt(input PromptInput) (string, string, error) {
	if strings.TrimSpace(input.Finding.Rule) == "" {
		return "", "", fmt.Errorf("judge prompt: finding rule is required")
	}
	if strings.TrimSpace(input.Finding.Message) == "" {
		return "", "", fmt.Errorf("judge prompt: finding message is required")
	}

	precedents := make([]promptPrecedentRecord, 0, len(input.Precedents))
	for _, p := range input.Precedents {
		precedents = append(precedents, promptPrecedentRecord{
			RuleID:      p.RuleID,
			Decision:    string(p.Decision),
			Rationale:   p.Rationale,
			Confidence:  p.Confidence,
			SeenCount:   p.SeenCount,
			PatternHash: p.PatternHash,
		})
	}

	payload := promptPayload{
		GeneratedAtUTC:  time.Now().UTC().Format(time.RFC3339),
		Finding:         input.Finding,
		RuleDescription: strings.TrimSpace(input.RuleDescription),
		Precedents:      precedents,
		ResponseSchema:  responseSchema(),
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("judge prompt: marshal payload: %w", err)
	}

	system := strings.TrimSpace(`
You are Truthsayer Senate Judge, a strict classifier for anti-pattern findings.
Return exactly one JSON object and no surrounding text.
Apply precedents when relevant, but do not ignore clear harmful behavior.
`)
	user := "Evaluate this finding and produce a judgment JSON that matches response_schema exactly:\n\n" + string(body)
	return system, user, nil
}

func responseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []string{
			"verdict",
			"reasoning",
			"confidence",
			"precedent_decision",
			"precedent_rationale",
		},
		"properties": map[string]any{
			"verdict": map[string]any{
				"type": "string",
				"enum": []string{"guilty", "not_guilty", "advisory"},
			},
			"reasoning": map[string]any{
				"type": "string",
			},
			"confidence": map[string]any{
				"type":    "number",
				"minimum": 0.0,
				"maximum": 1.0,
			},
			"precedent_decision": map[string]any{
				"type": "string",
				"enum": []string{"allow", "deny"},
			},
			"precedent_rationale": map[string]any{
				"type": "string",
			},
		},
	}
}
