package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSRethrowNoWrap detects catch blocks that rethrow the caught error without wrapping.
type JSRethrowNoWrap struct{}

func (j *JSRethrowNoWrap) Meta() Rule {
	return Rule{
		ID:          "error-context.js-rethrow-no-wrap",
		Category:    "error-context",
		Name:        "Rethrow without wrapping",
		Description: "Catch block rethrows the original error without adding context",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSRethrowNoWrap) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, catchNode := range jsFindNodesByType(tree.RootNode(), "catch_clause") {
		param := catchNode.ChildByFieldName("parameter")
		if param == nil {
			continue
		}
		paramName := jsNodeText(param, source)

		body := catchNode.ChildByFieldName("body")
		if body == nil {
			continue
		}

		// Look for throw statements that rethrow the same variable
		for _, throwNode := range jsFindNodesByType(body, "throw_statement") {
			if throwNode.NamedChildCount() == 0 {
				continue
			}
			thrown := throwNode.NamedChild(0)
			if thrown == nil {
				continue
			}
			// Direct rethrow: throw e
			if thrown.Type() == "identifier" && jsNodeText(thrown, source) == paramName {
				line := jsLineNumber(throwNode)
				findings = append(findings, finding.Finding{
					Rule:       j.Meta().ID,
					Severity:   j.Meta().Severity,
					File:       path,
					Line:       line,
					Code:       jsSourceLine(source, line),
					Message:    "Catch block rethrows the original error without adding context",
					Suggestion: "Wrap with new Error('context', { cause: " + paramName + " }) or add logging before rethrowing",
				})
			}
		}
	}
	return findings
}
