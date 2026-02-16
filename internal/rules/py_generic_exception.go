package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyGenericException detects `raise Exception("failed")` — base Exception without specific type.
type PyGenericException struct{}

func (p *PyGenericException) Meta() Rule {
	return Rule{
		ID:          "error-context.py-generic-exception",
		Category:    "error-context",
		Name:        "Generic Exception raised",
		Description: "Raising base Exception instead of a specific exception type",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

// pyGenericExceptions are base exception types that should be more specific.
var pyGenericExceptions = map[string]bool{
	"Exception":     true,
	"BaseException": true,
}

func (p *PyGenericException) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "raise_statement") {
		if node.NamedChildCount() == 0 {
			continue // bare raise
		}
		child := node.NamedChild(0)
		if child.Type() != "call" {
			continue
		}
		fn := child.ChildByFieldName("function")
		if fn == nil {
			continue
		}
		name := pyNodeText(fn, source)
		if !pyGenericExceptions[name] {
			continue
		}
		line := pyLineNumber(node)
		findings = append(findings, finding.Finding{
			Rule:       p.Meta().ID,
			Severity:   p.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       pySourceLine(source, line),
			Message:    "Raising " + name + " instead of a specific exception type",
			Suggestion: "Use a specific exception type like ValueError, TypeError, or define a custom exception",
		})
	}
	return findings
}
