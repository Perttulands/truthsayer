package rules

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSTestImportInSrc detects imports from testing libraries (@testing-library/*,
// vitest, jest) in non-test files.
type JSTestImportInSrc struct{}

func (j *JSTestImportInSrc) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.js-test-import-in-src",
		Category:    "mock-leakage",
		Name:        "Test library import in source",
		Description: "Import from test library (@testing-library, vitest, jest) in non-test file",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

// testLibPrefixes are module path prefixes that indicate test-only libraries.
// Entries without trailing "/" match both exact and subpackage imports.
var testLibPrefixes = []string{
	"@testing-library",
	"vitest",
	"jest",
	"@jest",
	"@vitest",
}

func (j *JSTestImportInSrc) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	root := tree.RootNode()
	var findings []finding.Finding

	// Check import_statement: import { render } from '@testing-library/react'
	for _, node := range jsFindNodesByType(root, "import_statement") {
		srcNode := node.ChildByFieldName("source")
		if srcNode == nil {
			continue
		}
		modPath := jsUnquote(jsNodeText(srcNode, source))
		if isTestLibrary(modPath) {
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Test library '" + modPath + "' imported in non-test file",
				Suggestion: "Move this import and its usage to a test file",
			})
		}
	}

	// Check require() calls: const { render } = require('@testing-library/react')
	for _, node := range jsFindNodesByType(root, "call_expression") {
		fn := node.ChildByFieldName("function")
		if fn == nil || jsNodeText(fn, source) != "require" {
			continue
		}
		args := node.ChildByFieldName("arguments")
		if args == nil || args.NamedChildCount() == 0 {
			continue
		}
		arg := args.NamedChild(0)
		if arg == nil || arg.Type() != "string" {
			continue
		}
		modPath := jsUnquote(jsNodeText(arg, source))
		if isTestLibrary(modPath) {
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Test library '" + modPath + "' required in non-test file",
				Suggestion: "Move this require and its usage to a test file",
			})
		}
	}

	return findings
}

func isTestLibrary(modPath string) bool {
	for _, prefix := range testLibPrefixes {
		if modPath == prefix || strings.HasPrefix(modPath, prefix+"/") {
			return true
		}
	}
	return false
}
