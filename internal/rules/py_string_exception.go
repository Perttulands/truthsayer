package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyStringException detects `raise "error"` — Python 2 idiom that is a TypeError in Python 3.
type PyStringException struct{}

func (p *PyStringException) Meta() Rule {
	return Rule{
		ID:          "error-context.py-string-exception",
		Category:    "error-context",
		Name:        "String exception raised",
		Description: "raise \"string\" is a Python 2 idiom that raises TypeError in Python 3",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyStringException) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "raise_statement") {
		if node.NamedChildCount() == 0 {
			continue // bare raise
		}
		child := node.NamedChild(0)
		if child.Type() == "string" || child.Type() == "concatenated_string" {
			line := pyLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       pySourceLine(source, line),
				Message:    "Raising a string is a Python 2 idiom — TypeError in Python 3",
				Suggestion: "Use raise Exception(\"message\") or a specific exception type",
			})
		}
	}
	return findings
}
