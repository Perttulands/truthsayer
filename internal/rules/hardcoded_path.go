package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// HardcodedPath detects hardcoded absolute paths like /home/username/ in scripts and configs.
type HardcodedPath struct{}

func (h *HardcodedPath) Meta() Rule {
	return Rule{
		ID:          "config-smells.hardcoded-path",
		Category:    "config-smells",
		Name:        "Hardcoded path",
		Description: "Hardcoded absolute path that should use a variable or config value",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".sh", ".bash", ".toml", ".yaml", ".yml", ".json", ".env"},
		ScanType:    ScanTypeRegex,
	}
}

var hardcodedPathPattern = regexp.MustCompile(`/(?:home|Users)/[a-zA-Z0-9_.-]+/`)

func (h *HardcodedPath) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}
		if hardcodedPathPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       h.Meta().ID,
				Severity:   h.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Hardcoded user-specific path detected",
				Suggestion: "Use $HOME, environment variable, or config file instead of hardcoded path",
			})
		}
	}
	return findings
}
