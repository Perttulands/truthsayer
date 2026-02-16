package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSPromiseReject detects Promise.reject() or reject() with non-Error values.
type JSPromiseReject struct{}

func (j *JSPromiseReject) Meta() Rule {
	return Rule{
		ID:          "error-context.js-promise-reject-non-error",
		Category:    "error-context",
		Name:        "Promise reject with non-Error",
		Description: "Promise.reject() or reject() called with a non-Error value loses stack trace",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSPromiseReject) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, node := range jsFindNodesByType(tree.RootNode(), "call_expression") {
		if !isRejectCall(node, source) {
			continue
		}

		args := node.ChildByFieldName("arguments")
		if args == nil || args.NamedChildCount() == 0 {
			continue
		}

		firstArg := args.NamedChild(0)
		if firstArg == nil {
			continue
		}

		// Flag non-Error arguments: strings, numbers, null, undefined, booleans
		if isNonErrorValue(firstArg) {
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Promise rejected with non-Error value — stack trace will be lost",
				Suggestion: "Use reject(new Error('reason')) to preserve stack traces",
			})
		}
	}
	return findings
}

// isRejectCall checks if a call_expression is Promise.reject(...) or a reject(...) call
// (the latter for new Promise((resolve, reject) => { reject(...) }))
func isRejectCall(node *sitter.Node, source []byte) bool {
	fn := node.ChildByFieldName("function")
	if fn == nil {
		return false
	}

	// Promise.reject(...)
	if fn.Type() == "member_expression" {
		obj := fn.ChildByFieldName("object")
		prop := fn.ChildByFieldName("property")
		if obj != nil && prop != nil &&
			jsNodeText(obj, source) == "Promise" &&
			jsNodeText(prop, source) == "reject" {
			return true
		}
	}

	// reject(...) — bare call inside Promise constructor callback
	if fn.Type() == "identifier" && jsNodeText(fn, source) == "reject" {
		return true
	}

	return false
}

// isNonErrorValue checks if a node represents a non-Error value (string, number, null, etc.)
func isNonErrorValue(node *sitter.Node) bool {
	switch node.Type() {
	case "string", "template_string", "number", "null", "undefined", "true", "false":
		return true
	}
	return false
}
