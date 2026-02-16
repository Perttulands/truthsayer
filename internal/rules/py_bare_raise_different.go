package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyBareRaiseDifferent detects `except ErrorA: raise ErrorB()` without `from` — loses original traceback.
type PyBareRaiseDifferent struct{}

func (p *PyBareRaiseDifferent) Meta() Rule {
	return Rule{
		ID:          "error-context.py-bare-raise-different",
		Category:    "error-context",
		Name:        "Raise different exception without from",
		Description: "Raising a different exception without from discards the original traceback",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyBareRaiseDifferent) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, exceptNode := range pyFindNodesByType(tree.RootNode(), "except_clause") {
		body := exceptBody(exceptNode)
		if body == nil {
			continue
		}
		// Find raise statements as direct children of the except body
		// (don't recurse into nested try/except blocks)
		for _, raiseNode := range directChildrenOfType(body, "raise_statement") {
			if isRaiseDifferentWithoutFrom(raiseNode) {
				line := pyLineNumber(raiseNode)
				findings = append(findings, finding.Finding{
					Rule:       p.Meta().ID,
					Severity:   p.Meta().Severity,
					File:       path,
					Line:       line,
					Code:       pySourceLine(source, line),
					Message:    "Raising a different exception without 'from' discards the original traceback",
					Suggestion: "Use 'raise NewError(...) from e' to chain exceptions",
				})
			}
		}
	}
	return findings
}

// isRaiseDifferentWithoutFrom checks if a raise_statement raises a new exception
// (via call/constructor) without using "from" to chain it.
func isRaiseDifferentWithoutFrom(node *sitter.Node) bool {
	if node.NamedChildCount() == 0 {
		// bare `raise` — re-raises current exception, this is fine
		return false
	}
	if node.NamedChildCount() >= 2 {
		// Has "from" clause — chained, this is fine
		return false
	}
	// Exactly 1 named child: raising a new exception without from
	child := node.NamedChild(0)
	return child.Type() == "call"
}
