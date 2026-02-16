package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSJestMockInSrc detects jest.mock/fn/spyOn in non-test files.
type JSJestMockInSrc struct{}

func (j *JSJestMockInSrc) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.jest-mock-in-src",
		Category:    "mock-leakage",
		Name:        "Jest mock in source",
		Description: "jest.mock/fn/spyOn found in non-test file — mocks should stay in test files",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeRegex,
	}
}

var jestMockPattern = regexp.MustCompile(`\bjest\.(mock|fn|spyOn)\s*\(`)

func (j *JSJestMockInSrc) CheckLines(path string, lines []string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}
		if jestMockPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "jest.mock/fn/spyOn found in non-test file",
				Suggestion: "Move mock setup to test files or __mocks__/ directory",
			})
		}
	}
	return findings
}

// JSStorybookInSrc detects .stories. imports in non-story source files.
type JSStorybookInSrc struct{}

func (s *JSStorybookInSrc) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.storybook-in-src",
		Category:    "mock-leakage",
		Name:        "Storybook import in source",
		Description: ".stories. import found in non-story file — storybook artifacts leaking into production",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeRegex,
	}
}

var storybookImportPattern = regexp.MustCompile(`(?:import\s+.*from\s+|require\s*\().*\.stories\.`)

func (s *JSStorybookInSrc) CheckLines(path string, lines []string) []finding.Finding {
	if jsIsTestFile(path) || isStoryFile(path) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}
		if storybookImportPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       s.Meta().ID,
				Severity:   s.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    ".stories. import found in non-story file",
				Suggestion: "Remove storybook import from production code",
			})
		}
	}
	return findings
}

func isStoryFile(path string) bool {
	return strings.Contains(path, ".stories.")
}
