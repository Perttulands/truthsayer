package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSEmptyCatch detects catch clauses with empty bodies in JS/TS code.
type JSEmptyCatch struct{}

func (j *JSEmptyCatch) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.js-empty-catch",
		Category:    "silent-fallback",
		Name:        "Empty catch block",
		Description: "Empty catch block silently swallows errors",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSEmptyCatch) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, node := range jsFindNodesByType(tree.RootNode(), "catch_clause") {
		body := node.ChildByFieldName("body")
		if body == nil {
			continue
		}
		if body.NamedChildCount() == 0 {
			bodyText := jsNodeText(body, source)
			if hasJSComment(bodyText) {
				continue
			}
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Empty catch block silently swallows errors",
				Suggestion: "Log the error or add a comment explaining why it's intentionally ignored",
			})
		}
	}
	return findings
}

// hasJSComment checks if a string contains a JS comment.
func hasJSComment(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '/' && (s[i+1] == '/' || s[i+1] == '*') {
			return true
		}
	}
	return false
}
