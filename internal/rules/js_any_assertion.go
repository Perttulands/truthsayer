package rules

import (
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSAnyAssertion detects `as any` type assertions in TypeScript files.
type JSAnyAssertion struct{}

func (j *JSAnyAssertion) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.any-type-assertion",
		Category:    "bad-defaults",
		Name:        "as any type assertion",
		Description: "`as any` defeats TypeScript's type safety",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".ts", ".tsx"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSAnyAssertion) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	ext := filepath.Ext(path)
	if ext != ".ts" && ext != ".tsx" {
		return nil
	}
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, node := range jsFindNodesByType(tree.RootNode(), "as_expression") {
		// Second named child is the type
		if node.NamedChildCount() < 2 {
			continue
		}
		typeNode := node.NamedChild(1)
		if typeNode.Type() == "predefined_type" && jsNodeText(typeNode, source) == "any" {
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "`as any` defeats TypeScript's type safety",
				Suggestion: "Use a specific type assertion or fix the type mismatch",
			})
		}
	}
	return findings
}
