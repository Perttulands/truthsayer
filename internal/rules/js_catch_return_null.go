package rules

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSCatchReturnNull detects .catch(() => null) and .catch(() => undefined) patterns.
type JSCatchReturnNull struct{}

func (j *JSCatchReturnNull) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.js-catch-return-null",
		Category:    "silent-fallback",
		Name:        "Catch returns null/undefined",
		Description: ".catch() handler that returns null or undefined silently swallows errors",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSCatchReturnNull) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	// Find all call_expression nodes — .catch(...) is a call on a member_expression
	for _, node := range jsFindNodesByType(tree.RootNode(), "call_expression") {
		fn := node.ChildByFieldName("function")
		if fn == nil || fn.Type() != "member_expression" {
			continue
		}
		prop := fn.ChildByFieldName("property")
		if prop == nil || jsNodeText(prop, source) != "catch" {
			continue
		}
		args := node.ChildByFieldName("arguments")
		if args == nil || args.NamedChildCount() == 0 {
			continue
		}
		callback := args.NamedChild(0)
		if callback == nil {
			continue
		}
		if returnsNullOrUndefined(callback, source) {
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    ".catch() handler returns null/undefined, silently swallowing the error",
				Suggestion: "Log or rethrow the error instead of returning null",
			})
		}
	}
	return findings
}

// returnsNullOrUndefined checks if a callback returns null or undefined.
// Handles both expression bodies: () => null
// and block bodies: () => { return null; }
func returnsNullOrUndefined(node *sitter.Node, source []byte) bool {
	if node.Type() != "arrow_function" && node.Type() != "function" {
		return false
	}

	body := node.ChildByFieldName("body")
	if body == nil {
		return false
	}

	// Expression body: () => null / () => undefined
	if body.Type() == "null" || body.Type() == "undefined" {
		return true
	}
	// Identifier "undefined" as expression body
	if body.Type() == "identifier" && jsNodeText(body, source) == "undefined" {
		return true
	}

	// Block body: check for return null/undefined
	if body.Type() == "statement_block" {
		for i := 0; i < int(body.NamedChildCount()); i++ {
			stmt := body.NamedChild(i)
			if stmt.Type() == "return_statement" {
				retText := strings.TrimSpace(jsNodeText(stmt, source))
				if retText == "return null;" || retText == "return null" ||
					retText == "return undefined;" || retText == "return undefined" {
					return true
				}
			}
		}
	}
	return false
}
