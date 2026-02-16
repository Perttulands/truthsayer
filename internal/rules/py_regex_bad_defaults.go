package rules

import (
	"regexp"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyTypeIgnoreBare detects # type: ignore without specific error code.
type PyTypeIgnoreBare struct{}

func (t *PyTypeIgnoreBare) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.type-ignore-bare",
		Category:    "bad-defaults",
		Name:        "Bare type: ignore",
		Description: "# type: ignore without specific error code — blanket type suppression",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeRegex,
	}
}

// Matches "# type: ignore" at end of line, but NOT "# type: ignore[error-code]".
var typeIgnoreBarePattern = regexp.MustCompile(`#\s*type:\s*ignore\s*$`)

func (t *PyTypeIgnoreBare) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		if typeIgnoreBarePattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       t.Meta().ID,
				Severity:   t.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Bare # type: ignore without specific error code",
				Suggestion: "Specify the error code: # type: ignore[assignment] or # type: ignore[return-value]",
			})
		}
	}
	return findings
}

// PyNoqaBare detects # noqa without specific code.
type PyNoqaBare struct{}

func (n *PyNoqaBare) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.noqa-bare",
		Category:    "bad-defaults",
		Name:        "Bare noqa comment",
		Description: "# noqa without specific code — blanket linter suppression",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeRegex,
	}
}

// Matches "# noqa" at end of line, but NOT "# noqa: E501" (with specific code).
var noqaBarePattern = regexp.MustCompile(`#\s*noqa\s*$`)

func (n *PyNoqaBare) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		if noqaBarePattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       n.Meta().ID,
				Severity:   n.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Bare # noqa without specific code",
				Suggestion: "Specify the code: # noqa: E501 or # noqa: F401",
			})
		}
	}
	return findings
}
