package rules

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSGenericErrorMessage detects throw new Error("static string") with no interpolation or variables.
type JSGenericErrorMessage struct{}

func (j *JSGenericErrorMessage) Meta() Rule {
	return Rule{
		ID:          "error-context.js-generic-error-message",
		Category:    "error-context",
		Name:        "Generic error message",
		Description: "Error thrown with a generic static message provides no context",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSGenericErrorMessage) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, node := range jsFindNodesByType(tree.RootNode(), "throw_statement") {
		if node.NamedChildCount() == 0 {
			continue
		}
		expr := node.NamedChild(0)
		if expr == nil || expr.Type() != "new_expression" {
			continue
		}

		// Check it's new Error(...) or new TypeError(...) etc.
		constructor := expr.ChildByFieldName("constructor")
		if constructor == nil || !isErrorConstructor(jsNodeText(constructor, source)) {
			continue
		}

		args := expr.ChildByFieldName("arguments")
		if args == nil || args.NamedChildCount() == 0 {
			continue
		}

		firstArg := args.NamedChild(0)
		if firstArg == nil {
			continue
		}

		// Only flag plain string literals (no template literals, no concatenation)
		if firstArg.Type() == "string" {
			msg := jsNodeText(firstArg, source)
			if isGenericMessage(msg) {
				line := jsLineNumber(node)
				findings = append(findings, finding.Finding{
					Rule:       j.Meta().ID,
					Severity:   j.Meta().Severity,
					File:       path,
					Line:       line,
					Code:       jsSourceLine(source, line),
					Message:    "Error thrown with generic message " + msg + " — add context about what failed",
					Suggestion: "Include variable values or operation context: new Error(`failed to load ${resource}: ${err.message}`)",
				})
			}
		}
	}
	return findings
}

// isErrorConstructor returns true for built-in JS error constructors.
func isErrorConstructor(name string) bool {
	switch name {
	case "Error", "TypeError", "RangeError", "ReferenceError", "SyntaxError", "URIError", "EvalError":
		return true
	}
	return false
}

// isGenericMessage checks if a quoted string is a generic/unhelpful error message.
func isGenericMessage(quoted string) bool {
	// Strip quotes
	if len(quoted) < 2 {
		return false
	}
	msg := quoted[1 : len(quoted)-1]
	msg = strings.TrimSpace(strings.ToLower(msg))

	genericMessages := []string{
		"error", "failed", "failure", "something went wrong",
		"an error occurred", "unknown error", "unexpected error",
		"internal error", "oops", "error occurred",
		"something failed", "request failed", "operation failed",
	}
	for _, g := range genericMessages {
		if msg == g {
			return true
		}
	}
	return false
}
