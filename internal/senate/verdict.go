package senate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Verdict represents a Senate ruling payload.
type Verdict struct {
	ID         string      `json:"id"`
	Status     string      `json:"status"`
	DecidedAt  time.Time   `json:"decided_at,omitempty"`
	Amendments []Amendment `json:"amendments"`
}

// Amendment represents one approved rule amendment.
type Amendment struct {
	RuleID    string `json:"rule_id"`
	Action    string `json:"action"`
	Severity  string `json:"severity,omitempty"`
	Exception string `json:"exception,omitempty"`
	Rationale string `json:"rationale,omitempty"`
}

const (
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusKeepAsIs = "keep_as_is"
)

const (
	ActionSetSeverity  = "set_severity"
	ActionDisableRule  = "disable_rule"
	ActionEnableRule   = "enable_rule"
	ActionAddException = "add_exception"
)

var fencedJSONPattern = regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")

// ParseVerdictFile parses and validates a Senate verdict file.
func ParseVerdictFile(path string) (Verdict, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Verdict{}, fmt.Errorf("senate: read verdict file: %w", err)
	}
	return ParseVerdict(data)
}

// ParseVerdict parses and validates a Senate verdict document.
func ParseVerdict(data []byte) (Verdict, error) {
	raw := strings.TrimSpace(string(data))
	if raw == "" {
		return Verdict{}, errors.New("senate: empty verdict payload")
	}

	payload := raw
	if !strings.HasPrefix(raw, "{") {
		match := fencedJSONPattern.FindStringSubmatch(raw)
		if len(match) < 2 {
			return Verdict{}, errors.New("senate: expected JSON object or fenced ```json block")
		}
		payload = strings.TrimSpace(match[1])
	}

	var verdict Verdict
	if err := json.Unmarshal([]byte(payload), &verdict); err != nil {
		return Verdict{}, fmt.Errorf("senate: decode verdict json: %w", err)
	}
	if err := verdict.Validate(); err != nil {
		return Verdict{}, err
	}
	return verdict, nil
}

// Validate ensures verdict and amendment schema correctness.
func (v Verdict) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return errors.New("senate: verdict id is required")
	}
	switch v.Status {
	case StatusApproved, StatusRejected, StatusKeepAsIs:
	default:
		return fmt.Errorf("senate: invalid status %q", v.Status)
	}
	for i, a := range v.Amendments {
		if err := a.Validate(); err != nil {
			return fmt.Errorf("senate: amendment %d: %w", i, err)
		}
	}
	if v.Status == StatusApproved && len(v.Amendments) == 0 {
		return errors.New("senate: approved verdict must include amendments")
	}
	return nil
}

// Validate ensures amendment fields are valid for the action type.
func (a Amendment) Validate() error {
	if strings.TrimSpace(a.RuleID) == "" {
		return errors.New("rule_id is required")
	}
	switch a.Action {
	case ActionSetSeverity:
		switch a.Severity {
		case "error", "warning", "info":
		default:
			return fmt.Errorf("set_severity requires severity error|warning|info, got %q", a.Severity)
		}
	case ActionDisableRule, ActionEnableRule:
		// no extra fields required
	case ActionAddException:
		if strings.TrimSpace(a.Exception) == "" {
			return errors.New("add_exception requires exception")
		}
	default:
		return fmt.Errorf("invalid action %q", a.Action)
	}
	return nil
}
