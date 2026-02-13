package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// SecretInConfig detects inline secrets like password=, token=, secret= in config files.
type SecretInConfig struct{}

func (s *SecretInConfig) Meta() Rule {
	return Rule{
		ID:          "config-smells.secret-in-config",
		Category:    "config-smells",
		Name:        "Secret in config",
		Description: "Potential secret or credential hardcoded in configuration file",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".toml", ".yaml", ".yml", ".json", ".env", ".sh", ".bash"},
		ScanType:    ScanTypeRegex,
	}
}

var secretPattern = regexp.MustCompile(`(?i)(password|passwd|secret|api_key|apikey|api[-_]?token|access[-_]?token|private[-_]?key)\s*[=:]\s*["']?[^\s"']{8,}`)

func (s *SecretInConfig) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}
		if secretPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       s.Meta().ID,
				Severity:   s.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Potential secret or credential in config file",
				Suggestion: "Use environment variable or secrets manager instead of inline credentials",
			})
		}
	}
	return findings
}
