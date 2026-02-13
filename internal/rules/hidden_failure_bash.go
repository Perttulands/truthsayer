package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// HiddenFailureBash detects || true, 2>/dev/null, || : patterns in bash scripts.
type HiddenFailureBash struct{}

func (h *HiddenFailureBash) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.hidden-failure-bash",
		Category:    "silent-fallback",
		Name:        "Hidden failure in bash",
		Description: "Command failure silently suppressed with || true, 2>/dev/null, or || :",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".sh", ".bash"},
		ScanType:    ScanTypeRegex,
	}
}

var hiddenFailurePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\|\|\s*true`),
	regexp.MustCompile(`\|\|\s*:`),
	regexp.MustCompile(`2>\s*/dev/null`),
}

var hiddenFailureMessages = []string{
	"'|| true' silently swallows command failure",
	"'|| :' silently swallows command failure",
	"'2>/dev/null' hides error output",
}

var hiddenFailureSuggestions = []string{
	"Handle the error explicitly or log it: cmd || { echo 'cmd failed' >&2; exit 1; }",
	"Handle the error explicitly or log it: cmd || { echo 'cmd failed' >&2; exit 1; }",
	"Capture stderr to a variable or log file instead of discarding it",
}

func (h *HiddenFailureBash) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		for j, pat := range hiddenFailurePatterns {
			if pat.MatchString(line) {
				findings = append(findings, finding.Finding{
					Rule:       h.Meta().ID,
					Severity:   h.Meta().Severity,
					File:       path,
					Line:       i + 1,
					Code:       line,
					Message:    hiddenFailureMessages[j],
					Suggestion: hiddenFailureSuggestions[j],
				})
			}
		}
	}
	return findings
}
