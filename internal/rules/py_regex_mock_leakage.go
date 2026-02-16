package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyPytestFixtureInSrc detects @pytest.fixture decorator in non-test files.
type PyPytestFixtureInSrc struct{}

func (p *PyPytestFixtureInSrc) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.pytest-fixture-in-src",
		Category:    "mock-leakage",
		Name:        "pytest fixture in source",
		Description: "@pytest.fixture in non-test file — test fixtures should stay in test/conftest files",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeRegex,
	}
}

var pytestFixturePattern = regexp.MustCompile(`@pytest\.fixture\b`)

func (p *PyPytestFixtureInSrc) CheckLines(path string, lines []string) []finding.Finding {
	if pyIsTestFile(path) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if pytestFixturePattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "@pytest.fixture found in non-test file",
				Suggestion: "Move fixture to a conftest.py or test file",
			})
		}
	}
	return findings
}
