package rules

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSTestOnlyImport detects imports of test-only modules in production source files.
type JSTestOnlyImport struct{}

func (j *JSTestOnlyImport) Meta() Rule {
	return Rule{
		ID:          "test-isolation.test-only-import",
		Category:    "test-isolation",
		Name:        "Test-only import in source",
		Description: "Module used exclusively for test setup imported in production source",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSTestOnlyImport) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	// Only flag non-test files — test files can import whatever they want
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding

	// Check ES module imports: import ... from '...'
	for _, node := range jsFindNodesByType(tree.RootNode(), "import_statement") {
		srcNode := node.ChildByFieldName("source")
		if srcNode == nil {
			continue
		}
		modPath := jsUnquote(jsNodeText(srcNode, source))
		if isTestOnlyModule(modPath) {
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Import of test-only module '" + modPath + "' in production source",
				Suggestion: "Move test helpers to test files or extract shared utilities to a non-test module",
			})
		}
	}

	// Check CommonJS require: const x = require('...')
	for _, node := range jsFindNodesByType(tree.RootNode(), "call_expression") {
		fn := node.ChildByFieldName("function")
		if fn == nil || jsNodeText(fn, source) != "require" {
			continue
		}
		args := node.ChildByFieldName("arguments")
		if args == nil || args.NamedChildCount() == 0 {
			continue
		}
		arg := args.NamedChild(0)
		if arg.Type() != "string" {
			continue
		}
		modPath := jsUnquote(jsNodeText(arg, source))
		if isTestOnlyModule(modPath) {
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Import of test-only module '" + modPath + "' in production source",
				Suggestion: "Move test helpers to test files or extract shared utilities to a non-test module",
			})
		}
	}

	return findings
}

// isTestOnlyModule checks if a module path indicates a test-only module.
func isTestOnlyModule(modPath string) bool {
	lower := strings.ToLower(modPath)

	// Check path segments for test-related directories
	parts := strings.Split(lower, "/")
	for _, part := range parts {
		switch part {
		case "__tests__", "__mocks__", "__fixtures__", "testutils", "test-utils":
			return true
		}
	}

	// Check the final segment (the actual module name) for test patterns
	last := parts[len(parts)-1]

	// Check suffixes on the raw segment (handles ./user.test, ./user.spec)
	testSuffixes := []string{".test", ".spec", "_test", "_spec", "-test", "-spec"}
	for _, suffix := range testSuffixes {
		if strings.HasSuffix(last, suffix) {
			return true
		}
	}

	// Strip file extension and recheck (handles ./user.test.js → user.test)
	if idx := strings.LastIndex(last, "."); idx > 0 {
		stripped := last[:idx]
		for _, suffix := range testSuffixes {
			if strings.HasSuffix(stripped, suffix) {
				return true
			}
		}
		last = stripped
	}

	testPrefixes := []string{"test-helper", "test_helper", "testhelper", "mock-", "mock_", "fake-", "fake_", "stub-", "stub_"}
	for _, prefix := range testPrefixes {
		if strings.HasPrefix(last, prefix) {
			return true
		}
	}

	// Exact matches
	testNames := map[string]bool{
		"test-helpers": true, "test_helpers": true, "testhelpers": true,
		"test-utils": true, "test_utils": true, "testutils": true,
		"test-setup": true, "test_setup": true, "testsetup": true,
		"mocks": true, "fixtures": true, "stubs": true, "fakes": true,
	}
	return testNames[last]
}
