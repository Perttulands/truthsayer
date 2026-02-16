package rules

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSCallbackErrIgnored detects callback functions where the error parameter is never referenced.
type JSCallbackErrIgnored struct{}

func (j *JSCallbackErrIgnored) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.js-callback-err-ignored",
		Category:    "silent-fallback",
		Name:        "Callback error parameter ignored",
		Description: "Callback error parameter is never referenced in the function body",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

// errParamNames are conventional error parameter names in callbacks.
var errParamNames = map[string]bool{
	"err":   true,
	"error": true,
	"e":     true,
}

func (j *JSCallbackErrIgnored) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding

	// Find arrow_function and function_expression nodes that appear as arguments to a call
	for _, fn := range jsFindNodesByType(tree.RootNode(), "arrow_function", "function_expression") {
		// Must be inside an arguments node (i.e., passed as a callback)
		if !isInsideArguments(fn) {
			continue
		}

		params := fn.ChildByFieldName("parameters")
		if params == nil {
			continue
		}

		// Need at least 2 parameters (err, data) pattern
		if params.NamedChildCount() < 2 {
			continue
		}

		firstParam := params.NamedChild(0)
		if firstParam == nil || firstParam.Type() != "identifier" {
			continue
		}

		errName := jsNodeText(firstParam, source)
		if !errParamNames[errName] {
			continue
		}

		// Check if the error parameter is referenced in the function body
		body := fn.ChildByFieldName("body")
		if body == nil {
			continue
		}

		if !identifierReferencedInBody(body, errName, source) {
			line := jsLineNumber(fn)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Callback error parameter '" + errName + "' is never referenced",
				Suggestion: "Check the error: if (" + errName + ") { ... } or use _ prefix to indicate intentional ignore",
			})
		}
	}
	return findings
}

// isInsideArguments checks if a node is a direct child of an arguments node.
func isInsideArguments(node *sitter.Node) bool {
	parent := node.Parent()
	if parent == nil {
		return false
	}
	return parent.Type() == "arguments"
}

// identifierReferencedInBody checks if a given identifier name appears anywhere
// in the body, excluding the parameter declaration itself.
func identifierReferencedInBody(body *sitter.Node, name string, source []byte) bool {
	// Quick text check first — if the name doesn't appear in the body text at all, skip traversal
	bodyText := jsNodeText(body, source)
	if !strings.Contains(bodyText, name) {
		return false
	}

	found := false
	jsWalkNode(body, func(n *sitter.Node) {
		if found {
			return
		}
		if n.Type() == "identifier" && jsNodeText(n, source) == name {
			found = true
		}
	})
	return found
}
