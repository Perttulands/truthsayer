package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyPrintDebug detects print() used for debugging in non-test/non-script files.
type PyPrintDebug struct{}

func (p *PyPrintDebug) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.print-debug",
		Category:    "trace-gaps",
		Name:        "print() used for debugging",
		Description: "print() in production code — use logging module instead",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeRegex,
	}
}

var printPattern = regexp.MustCompile(`\bprint\s*\(`)

func (p *PyPrintDebug) CheckLines(path string, lines []string) []finding.Finding {
	if pyIsTestFile(path) || pyIsScript(path) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if printPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "print() used for debugging in production code",
				Suggestion: "Use logging.info() or logging.debug() instead of print()",
			})
		}
	}
	return findings
}

// pyIsScript returns true if the file is in a scripts/ or bin/ directory.
func pyIsScript(path string) bool {
	parts := strings.Split(strings.ReplaceAll(path, "\\", "/"), "/")
	for _, p := range parts {
		if p == "scripts" || p == "bin" {
			return true
		}
	}
	return false
}

// PyNoLoggingConfig detects Python packages without logging configuration.
type PyNoLoggingConfig struct{}

func (n *PyNoLoggingConfig) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.no-logging-config",
		Category:    "trace-gaps",
		Name:        "No logging configuration",
		Description: "Entry point without logging.basicConfig() or logging.getLogger() — logs may be silently lost",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeRegex,
	}
}

var loggingConfigPattern = regexp.MustCompile(`logging\.(basicConfig|getLogger)\s*\(`)
var pyEntryPointIndicator = regexp.MustCompile(`if\s+__name__\s*==\s*['"]__main__['"]`)

func (n *PyNoLoggingConfig) CheckLines(path string, lines []string) []finding.Finding {
	if pyIsTestFile(path) {
		return nil
	}

	isEntryPoint := false
	hasLogging := false

	for _, line := range lines {
		if pyEntryPointIndicator.MatchString(line) {
			isEntryPoint = true
		}
		if loggingConfigPattern.MatchString(line) {
			hasLogging = true
		}
	}

	if !isEntryPoint || hasLogging {
		return nil
	}

	code := ""
	if len(lines) > 0 {
		code = lines[0]
	}

	return []finding.Finding{{
		Rule:       n.Meta().ID,
		Severity:   n.Meta().Severity,
		File:       path,
		Line:       1,
		Code:       code,
		Message:    "Entry point without logging configuration",
		Suggestion: "Add logging.basicConfig() or logging.getLogger(__name__) to configure logging",
	}}
}
