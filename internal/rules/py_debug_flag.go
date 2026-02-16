package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyDebugFlag detects `if __debug__:` or `if DEBUG:` guards in production code.
type PyDebugFlag struct{}

func (p *PyDebugFlag) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.py-debug-flag",
		Category:    "mock-leakage",
		Name:        "Debug flag guard in production code",
		Description: "if __debug__: or if DEBUG: guard in production code may leak debug behavior",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyDebugFlag) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if pyIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "if_statement") {
		cond := node.ChildByFieldName("condition")
		if cond == nil {
			continue
		}
		condText := pyNodeText(cond, source)
		if condText != "__debug__" && condText != "DEBUG" {
			continue
		}
		// Check if the body has side effects (not just pass or assignment)
		body := node.ChildByFieldName("consequence")
		if body == nil || !pyBlockHasSideEffects(body) {
			continue
		}
		line := pyLineNumber(node)
		findings = append(findings, finding.Finding{
			Rule:       p.Meta().ID,
			Severity:   p.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       pySourceLine(source, line),
			Message:    "Debug guard 'if " + condText + ":' with side effects in production code",
			Suggestion: "Remove the debug guard or use a proper feature flag / logging configuration",
		})
	}
	return findings
}

// pyBlockHasSideEffects checks if a block node contains statements with side effects
// (function calls, print, logging, etc.) rather than just pass or variable assignments.
func pyBlockHasSideEffects(block *sitter.Node) bool {
	for i := 0; i < int(block.NamedChildCount()); i++ {
		child := block.NamedChild(i)
		switch child.Type() {
		case "expression_statement":
			// Expression statements containing calls (print, log, etc.) are side effects
			return true
		case "return_statement", "raise_statement", "assert_statement":
			return true
		case "pass_statement":
			continue
		default:
			// Any other statement type (if, for, while, etc.) likely has side effects
			return true
		}
	}
	return false
}
