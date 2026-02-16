package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSEnvTestCheck detects process.env.NODE_ENV === 'test' checks in production code.
// This pattern creates divergent behavior between test and production environments.
type JSEnvTestCheck struct{}

func (j *JSEnvTestCheck) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.js-env-test-check",
		Category:    "mock-leakage",
		Name:        "NODE_ENV test check in production code",
		Description: "process.env.NODE_ENV === 'test' creates test-specific behavior in production code",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSEnvTestCheck) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	root := tree.RootNode()
	var findings []finding.Finding

	// Look for binary expressions like: process.env.NODE_ENV === 'test'
	for _, node := range jsFindNodesByType(root, "binary_expression") {
		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")
		op := node.ChildByFieldName("operator")
		if left == nil || right == nil || op == nil {
			continue
		}

		opText := jsNodeText(op, source)
		if opText != "===" && opText != "==" && opText != "!==" && opText != "!=" {
			continue
		}

		// Check both orientations: NODE_ENV === 'test' and 'test' === NODE_ENV
		match := (isNodeEnvAccess(left, source) && isTestString(right, source)) ||
			(isNodeEnvAccess(right, source) && isTestString(left, source))
		if match {
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Checking NODE_ENV for 'test' creates divergent test/production behavior",
				Suggestion: "Use dependency injection or configuration instead of environment checks",
			})
		}
	}

	return findings
}

// isNodeEnvAccess checks if a node is process.env.NODE_ENV
func isNodeEnvAccess(node *sitter.Node, source []byte) bool {
	text := jsNodeText(node, source)
	return text == "process.env.NODE_ENV"
}

// isTestString checks if a node is the string literal 'test' or "test"
func isTestString(node *sitter.Node, source []byte) bool {
	if node.Type() != "string" {
		return false
	}
	return jsUnquote(jsNodeText(node, source)) == "test"
}
