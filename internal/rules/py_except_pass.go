package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyExceptPass detects `except SomeError: pass` — catching an exception and doing nothing.
type PyExceptPass struct{}

func (p *PyExceptPass) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.py-except-pass",
		Category:    "silent-fallback",
		Name:        "Exception caught and ignored",
		Description: "except with only pass silently swallows the exception",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyExceptPass) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "except_clause") {
		body := exceptBody(node)
		if body == nil {
			continue
		}
		// Check if body contains only a pass_statement
		if body.NamedChildCount() == 1 && body.NamedChild(0).Type() == "pass_statement" {
			line := pyLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       pySourceLine(source, line),
				Message:    "Exception caught but silently ignored with pass",
				Suggestion: "Log the error or re-raise it",
			})
		}
	}
	return findings
}

// exceptBody returns the block node inside an except_clause.
func exceptBody(node *sitter.Node) *sitter.Node {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "block" {
			return child
		}
	}
	return nil
}
