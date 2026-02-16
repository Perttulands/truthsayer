package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSConsoleErrorNoThrow detects console.error() in catch blocks without a re-throw.
type JSConsoleErrorNoThrow struct{}

func (j *JSConsoleErrorNoThrow) Meta() Rule {
	return Rule{
		ID:          "error-context.js-console-error-no-throw",
		Category:    "error-context",
		Name:        "Console.error without rethrow",
		Description: "console.error() in catch block without re-throwing silently swallows the error",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSConsoleErrorNoThrow) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, catchNode := range jsFindNodesByType(tree.RootNode(), "catch_clause") {
		body := catchNode.ChildByFieldName("body")
		if body == nil {
			continue
		}

		hasConsoleError := false
		var consoleErrorNode *sitter.Node
		hasThrow := false

		// Walk the catch body looking for console.error and throw
		jsWalkNode(body, func(n *sitter.Node) {
			if n.Type() == "call_expression" {
				fn := n.ChildByFieldName("function")
				if fn != nil && fn.Type() == "member_expression" {
					obj := fn.ChildByFieldName("object")
					prop := fn.ChildByFieldName("property")
					if obj != nil && prop != nil &&
						jsNodeText(obj, source) == "console" &&
						jsNodeText(prop, source) == "error" {
						hasConsoleError = true
						consoleErrorNode = n
					}
				}
			}
			if n.Type() == "throw_statement" {
				hasThrow = true
			}
		})

		if hasConsoleError && !hasThrow {
			line := jsLineNumber(consoleErrorNode)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "console.error() in catch block without re-throwing the error",
				Suggestion: "Either rethrow the error after logging, or use a proper error reporting mechanism",
			})
		}
	}
	return findings
}
