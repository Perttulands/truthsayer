package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// NoErrTrap detects bash scripts with set -e but no trap ... ERR handler.
type NoErrTrap struct{}

func (n *NoErrTrap) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.no-err-trap",
		Category:    "silent-fallback",
		Name:        "No ERR trap",
		Description: "Script uses set -e but has no trap for ERR — failures may go unnoticed",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".sh", ".bash"},
		ScanType:    ScanTypeRegex,
	}
}

var setEPattern = regexp.MustCompile(`set\s+-[a-zA-Z]*e`)
var trapErrPattern = regexp.MustCompile(`trap\s+.*\bERR\b`)

func (n *NoErrTrap) CheckLines(path string, lines []string) []finding.Finding {
	hasSetE := false
	hasTrapErr := false
	setELine := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if setEPattern.MatchString(line) && !hasSetE {
			hasSetE = true
			setELine = i
		}
		if trapErrPattern.MatchString(line) {
			hasTrapErr = true
		}
	}

	if hasSetE && !hasTrapErr {
		return []finding.Finding{
			{
				Rule:       n.Meta().ID,
				Severity:   n.Meta().Severity,
				File:       path,
				Line:       setELine + 1,
				Code:       lines[setELine],
				Message:    "set -e without trap ERR — errors exit silently without cleanup or logging",
				Suggestion: "Add: trap 'echo \"Error at line $LINENO\" >&2' ERR",
			},
		}
	}
	return nil
}
