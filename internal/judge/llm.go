package judge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/llm"
	"github.com/perttulands/truthsayer/internal/precedent"
)

// VerdictType is the judgment class emitted by the LLM.
type VerdictType string

const (
	VerdictGuilty    VerdictType = "guilty"
	VerdictNotGuilty VerdictType = "not_guilty"
	VerdictAdvisory  VerdictType = "advisory"
)

// Verdict is a structured judgment result for a finding.
type Verdict struct {
	Verdict           VerdictType         `json:"verdict"`
	Reasoning         string              `json:"reasoning"`
	Confidence        float64             `json:"confidence"`
	PrecedentDecision precedent.Decision  `json:"precedent_decision"`
	PrecedentRationale string             `json:"precedent_rationale"`
	Source            string              `json:"source"` // llm or precedent
	InputTokens       int                 `json:"input_tokens,omitempty"`
	OutputTokens      int                 `json:"output_tokens,omitempty"`
}

// ToPrecedent converts a verdict + finding into a precedent record.
func (v Verdict) ToPrecedent(f finding.Finding, at time.Time) PrecedentRecord {
	decision := v.PrecedentDecision
	if decision == "" {
		if v.Verdict == VerdictNotGuilty {
			decision = precedent.DecisionAllow
		} else {
			decision = precedent.DecisionDeny
		}
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	return PrecedentRecord{
		RuleID:        f.Rule,
		ViolationHash: precedent.HashViolation(f.Rule, f.File, f.Line, f.Code, ""),
		PatternHash:   precedent.HashFindingPattern(f),
		Decision:      decision,
		Rationale:     strings.TrimSpace(v.PrecedentRationale),
		Confidence:    v.Confidence,
		SeenCount:     1,
		CreatedAt:     at.UTC(),
	}
}

// PrecedentRecord is a small alias-like projection used by judge pipeline.
type PrecedentRecord struct {
	RuleID        string
	ViolationHash string
	PatternHash   string
	Decision      precedent.Decision
	Rationale     string
	Confidence    float64
	SeenCount     int
	CreatedAt     time.Time
}

// AsPrecedent converts to storage format.
func (p PrecedentRecord) AsPrecedent() precedent.Precedent {
	return precedent.Precedent{
		RuleID:        p.RuleID,
		ViolationHash: p.ViolationHash,
		PatternHash:   p.PatternHash,
		Decision:      p.Decision,
		Rationale:     p.Rationale,
		Confidence:    p.Confidence,
		SeenCount:     p.SeenCount,
		CreatedAt:     p.CreatedAt,
	}
}

// Completer abstracts the LLM client for testability.
type Completer interface {
	Complete(ctx context.Context, systemPrompt, userPrompt string) (llm.Completion, error)
}

// LLMJudge performs judgment calls via LLM completions.
type LLMJudge struct {
	completer Completer
}

// NewLLMJudge builds a judge over a completion client.
func NewLLMJudge(completer Completer) (*LLMJudge, error) {
	if completer == nil {
		return nil, errors.New("judge: completer is required")
	}
	return &LLMJudge{completer: completer}, nil
}

// JudgeFinding calls the LLM and parses a structured verdict.
func (j *LLMJudge) JudgeFinding(ctx context.Context, input PromptInput) (Verdict, error) {
	systemPrompt, userPrompt, err := BuildPrompt(input)
	if err != nil {
		return Verdict{}, fmt.Errorf("judge: build prompt: %w", err)
	}

	completion, err := j.completer.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		return Verdict{}, fmt.Errorf("judge: llm completion: %w", err)
	}

	verdict, err := ParseVerdict(completion.Text)
	if err != nil {
		return Verdict{}, fmt.Errorf("judge: parse verdict: %w", err)
	}
	verdict.Source = "llm"
	verdict.InputTokens = completion.InputTokens
	verdict.OutputTokens = completion.OutputTokens
	return verdict, nil
}

// ParseVerdict parses a JSON verdict from raw LLM text.
func ParseVerdict(raw string) (Verdict, error) {
	payload, err := extractJSONObject(raw)
	if err != nil {
		return Verdict{}, fmt.Errorf("extract verdict JSON: %w", err)
	}

	type verdictPayload struct {
		Verdict            string  `json:"verdict"`
		Reasoning          string  `json:"reasoning"`
		Confidence         float64 `json:"confidence"`
		PrecedentDecision  string  `json:"precedent_decision"`
		PrecedentRationale string  `json:"precedent_rationale"`
	}
	var parsed verdictPayload
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return Verdict{}, fmt.Errorf("invalid verdict json: %w", err)
	}

	v := Verdict{
		Verdict:            VerdictType(strings.TrimSpace(parsed.Verdict)),
		Reasoning:          strings.TrimSpace(parsed.Reasoning),
		Confidence:         parsed.Confidence,
		PrecedentDecision:  precedent.Decision(strings.TrimSpace(parsed.PrecedentDecision)),
		PrecedentRationale: strings.TrimSpace(parsed.PrecedentRationale),
	}
	if err := validateVerdict(v); err != nil {
		return Verdict{}, fmt.Errorf("validate verdict: %w", err)
	}
	return v, nil
}

func validateVerdict(v Verdict) error {
	switch v.Verdict {
	case VerdictGuilty, VerdictNotGuilty, VerdictAdvisory:
	default:
		return fmt.Errorf("invalid verdict %q", v.Verdict)
	}
	if strings.TrimSpace(v.Reasoning) == "" {
		return errors.New("reasoning is required")
	}
	if v.Confidence < 0 || v.Confidence > 1 {
		return fmt.Errorf("confidence out of range [0,1]: %f", v.Confidence)
	}
	if v.PrecedentDecision != precedent.DecisionAllow && v.PrecedentDecision != precedent.DecisionDeny {
		return fmt.Errorf("invalid precedent_decision %q", v.PrecedentDecision)
	}
	if strings.TrimSpace(v.PrecedentRationale) == "" {
		return errors.New("precedent_rationale is required")
	}
	return nil
}

func extractJSONObject(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("empty response")
	}
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return "", errors.New("no json object found")
	}

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], nil
			}
		}
	}
	return "", errors.New("unterminated json object")
}
