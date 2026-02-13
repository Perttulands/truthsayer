package finding

import (
	"fmt"
	"sort"
)

// Severity represents the severity level of a finding.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

func (s Severity) rank() int {
	switch s {
	case SeverityError:
		return 0
	case SeverityWarning:
		return 1
	case SeverityInfo:
		return 2
	default:
		return 3
	}
}

// Finding represents a single anti-pattern detection result.
type Finding struct {
	Rule       string   `json:"rule"`
	Severity   Severity `json:"severity"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion"`
}

func (f Finding) key() string {
	return fmt.Sprintf("%s:%s:%d", f.Rule, f.File, f.Line)
}

// Dedup removes duplicate findings (same rule, file, and line).
func Dedup(findings []Finding) []Finding {
	seen := make(map[string]struct{})
	result := make([]Finding, 0, len(findings))
	for _, f := range findings {
		k := f.key()
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			result = append(result, f)
		}
	}
	return result
}

// Sort sorts findings by severity (error first), then file path, then line number.
func Sort(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Severity != b.Severity {
			return a.Severity.rank() < b.Severity.rank()
		}
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Line < b.Line
	})
}
