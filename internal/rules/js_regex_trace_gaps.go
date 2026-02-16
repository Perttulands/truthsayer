package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSNoUnhandledRejection detects entry points without unhandledRejection handler.
type JSNoUnhandledRejection struct{}

func (n *JSNoUnhandledRejection) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.no-unhandled-rejection",
		Category:    "trace-gaps",
		Name:        "No unhandledRejection handler",
		Description: "Entry point without process.on('unhandledRejection') — crashes silently on unhandled promises",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".ts", ".mjs"},
		ScanType:    ScanTypeRegex,
	}
}

var unhandledRejectionPattern = regexp.MustCompile(`process\.on\s*\(\s*['"]unhandledRejection['"]`)
var entryPointIndicator = regexp.MustCompile(`(?:\.listen\s*\(|app\.use\s*\(|createServer\s*\(|main\s*\(\s*\))`)

func (n *JSNoUnhandledRejection) CheckLines(path string, lines []string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	// Only check files that look like entry points
	isEntryPoint := false
	hasHandler := false

	for _, line := range lines {
		if entryPointIndicator.MatchString(line) {
			isEntryPoint = true
		}
		if unhandledRejectionPattern.MatchString(line) {
			hasHandler = true
		}
	}

	if !isEntryPoint || hasHandler {
		return nil
	}

	return []finding.Finding{{
		Rule:       n.Meta().ID,
		Severity:   n.Meta().Severity,
		File:       path,
		Line:       1,
		Code:       lines[0],
		Message:    "Entry point without unhandledRejection handler",
		Suggestion: "Add process.on('unhandledRejection', (err) => { ... }) to handle unhandled promise rejections",
	}}
}

// JSConsoleLogInProduction detects console.log in non-test source.
type JSConsoleLogInProduction struct{}

func (c *JSConsoleLogInProduction) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.console-log-in-production",
		Category:    "trace-gaps",
		Name:        "console.log in production",
		Description: "console.log in production code — use a structured logger instead",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeRegex,
	}
}

var consoleLogPattern = regexp.MustCompile(`\bconsole\.log\s*\(`)

func (c *JSConsoleLogInProduction) CheckLines(path string, lines []string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}
		if consoleLogPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       c.Meta().ID,
				Severity:   c.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "console.log in production code",
				Suggestion: "Use a structured logger (winston, pino, bunyan) or remove debug logging",
			})
		}
	}
	return findings
}
