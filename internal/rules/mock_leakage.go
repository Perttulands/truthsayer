package rules

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// MockImportNonTest detects mock/testify imports in non-test Go files.
type MockImportNonTest struct{}

func (m *MockImportNonTest) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.mock-import-non-test",
		Category:    "mock-leakage",
		Name:        "Mock import in non-test file",
		Description: "Test/mock dependency imported in production code",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

var mockImportPatterns = []string{
	"testify",
	"mock",
	"gomock",
	"httptest",
	"testing",
}

func (m *MockImportNonTest) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	fname := fset.File(file.Pos()).Name()
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	var findings []finding.Finding
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		for _, pat := range mockImportPatterns {
			if strings.Contains(path, pat) {
				// Allow "testing" only if the file is a test helper
				if pat == "testing" && strings.Contains(fname, "test_helper") {
					continue
				}
				pos := fset.Position(imp.Pos())
				findings = append(findings, finding.Finding{
					Rule:       m.Meta().ID,
					Severity:   m.Meta().Severity,
					File:       fname,
					Line:       pos.Line,
					Code:       sourceLine(lines, pos.Line),
					Message:    "Test/mock package imported in production code: " + path,
					Suggestion: "Move test dependencies to _test.go files",
				})
				break
			}
		}
	}
	return findings
}

// TestFixtureRef detects testdata/ or fixture path references in non-test files.
type TestFixtureRef struct{}

func (t *TestFixtureRef) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.test-fixture-ref",
		Category:    "mock-leakage",
		Name:        "Test fixture reference in non-test file",
		Description: "Reference to testdata/ or fixture path in production code",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeRegex,
	}
}

var fixtureRefPattern = regexp.MustCompile(`(?:testdata/|fixture[s]?/|test_data/)`)

func (t *TestFixtureRef) CheckLines(path string, lines []string) []finding.Finding {
	if strings.HasSuffix(path, "_test.go") {
		return nil
	}
	if strings.HasSuffix(filepath.ToSlash(path), "internal/rules/mock_leakage.go") {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if strings.Contains(line, "regexp.MustCompile") || strings.Contains(line, "regexp.Compile") {
			continue
		}
		if fixtureRefPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       t.Meta().ID,
				Severity:   t.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Reference to test fixture path in production code",
				Suggestion: "Move fixture references to _test.go files or use a config-driven path",
			})
		}
	}
	return findings
}
