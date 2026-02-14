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
		Severity:    finding.SeverityError,
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

// hasReasonComment checks if the line has an inline # REASON: comment justifying the suppression,
// or if the immediately preceding line is a # REASON: comment.
func hasReasonComment(line string, lines []string, lineIdx int) bool {
	if strings.Contains(line, "# REASON:") {
		return true
	}
	if lineIdx > 0 && strings.Contains(lines[lineIdx-1], "# REASON:") {
		return true
	}
	return false
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
				severity := finding.SeverityError
				msg := hiddenFailureMessages[j]
				suggestion := hiddenFailureSuggestions[j]

				if hasReasonComment(line, lines, i) {
					severity = finding.SeverityInfo
					msg = msg + " (justified with REASON comment)"
					suggestion = ""
				} else {
					msg = msg + " — add '# REASON: ...' to justify or fix the suppression"
				}

				findings = append(findings, finding.Finding{
					Rule:       h.Meta().ID,
					Severity:   severity,
					File:       path,
					Line:       i + 1,
					Code:       line,
					Message:    msg,
					Suggestion: suggestion,
				})
			}
		}
	}
	return findings
}
