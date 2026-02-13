package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// NoStderrCapture detects exec.Command usage without stderr capture.
type NoStderrCapture struct{}

func (n *NoStderrCapture) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.no-stderr-capture",
		Category:    "trace-gaps",
		Name:        "No stderr capture",
		Description: "exec.Command usage without capturing stderr output",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeRegex,
	}
}

var execCommandPattern = regexp.MustCompile(`\bexec\.(?:Command|CommandContext)\s*\(`)
var stderrPipePattern = regexp.MustCompile(`\.StderrPipe\s*\(`)
var combinedOutputPattern = regexp.MustCompile(`\.CombinedOutput\s*\(`)
var stderrAssignPattern = regexp.MustCompile(`\.Stderr\s*=`)

func (n *NoStderrCapture) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if !execCommandPattern.MatchString(line) {
			continue
		}
		if hasNearbyStderrCapture(lines, i) {
			continue
		}

		findings = append(findings, finding.Finding{
			Rule:       n.Meta().ID,
			Severity:   n.Meta().Severity,
			File:       path,
			Line:       i + 1,
			Code:       line,
			Message:    "exec.Command used without stderr capture",
			Suggestion: "Capture stderr with cmd.StderrPipe(), cmd.Stderr = ..., or CombinedOutput()",
		})
	}

	return findings
}

func hasNearbyStderrCapture(lines []string, start int) bool {
	end := start + 6
	if end >= len(lines) {
		end = len(lines) - 1
	}

	for i := start; i <= end; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if i > start && execCommandPattern.MatchString(line) {
			break
		}
		if stderrPipePattern.MatchString(line) || combinedOutputPattern.MatchString(line) || stderrAssignPattern.MatchString(line) {
			return true
		}
	}

	return false
}
