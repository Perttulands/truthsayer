package rules

import (
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSNonNullAssertion detects `variable!` non-null assertions in TypeScript files.
type JSNonNullAssertion struct{}

func (j *JSNonNullAssertion) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.non-null-assertion",
		Category:    "bad-defaults",
		Name:        "Non-null assertion",
		Description: "`variable!` overrides null safety — runtime null possible despite compile-time override",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".ts", ".tsx"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSNonNullAssertion) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	ext := filepath.Ext(path)
	if ext != ".ts" && ext != ".tsx" {
		return nil
	}
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, node := range jsFindNodesByType(tree.RootNode(), "non_null_expression") {
		line := jsLineNumber(node)
		findings = append(findings, finding.Finding{
			Rule:       j.Meta().ID,
			Severity:   j.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       jsSourceLine(source, line),
			Message:    "Non-null assertion `!` overrides null safety — runtime null is still possible",
			Suggestion: "Use a null check or optional chaining instead",
		})
	}
	return findings
}
