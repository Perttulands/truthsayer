package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// UnvalidatedEnvBash detects ${VAR:-default} patterns that silently mask missing required values.
type UnvalidatedEnvBash struct{}

func (u *UnvalidatedEnvBash) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.unvalidated-env-bash",
		Category:    "bad-defaults",
		Name:        "Unvalidated env default in bash",
		Description: "Environment variable with silent default that may mask a missing required value",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".sh", ".bash"},
		ScanType:    ScanTypeRegex,
	}
}

// Matches ${VAR:-default} and ${VAR:=default} patterns.
var unvalidatedEnvPattern = regexp.MustCompile(`\$\{[A-Z_][A-Z0-9_]*:-[^}]+\}`)
var unvalidatedEnvAssignPattern = regexp.MustCompile(`\$\{[A-Z_][A-Z0-9_]*:=[^}]+\}`)

// Lines that are purely echo/printf output — env defaults here are informational, not bugs.
var echoPrintfLine = regexp.MustCompile(`^[[:space:]]*(echo|printf)\s`)

func (u *UnvalidatedEnvBash) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Skip env var defaults that appear only in echo/printf output context.
		if echoPrintfLine.MatchString(line) {
			continue
		}
		matched := false
		if unvalidatedEnvPattern.MatchString(line) {
			matched = true
		}
		if unvalidatedEnvAssignPattern.MatchString(line) {
			matched = true
		}
		if matched {
			findings = append(findings, finding.Finding{
				Rule:       u.Meta().ID,
				Severity:   u.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Environment variable uses silent default — missing value won't be caught",
				Suggestion: "Validate required env vars explicitly: ${VAR:?\"VAR is required\"} or check before use",
			})
		}
	}
	return findings
}
