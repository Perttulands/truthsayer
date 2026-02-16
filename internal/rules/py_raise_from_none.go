package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyRaiseFromNone detects `raise NewError() from None` — explicitly discards exception chain.
type PyRaiseFromNone struct{}

func (p *PyRaiseFromNone) Meta() Rule {
	return Rule{
		ID:          "error-context.py-raise-from-none",
		Category:    "error-context",
		Name:        "raise from None discards traceback",
		Description: "raise ... from None explicitly discards the original exception chain",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyRaiseFromNone) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "raise_statement") {
		if node.NamedChildCount() < 2 {
			continue
		}
		// Second named child is the "from" target
		fromTarget := node.NamedChild(int(node.NamedChildCount()) - 1)
		if fromTarget.Type() == "none" {
			line := pyLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       pySourceLine(source, line),
				Message:    "raise ... from None explicitly discards the original exception traceback",
				Suggestion: "Use raise ... from e to preserve the exception chain",
			})
		}
	}
	return findings
}
