package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSEvalUsage detects eval() and new Function() usage in production code.
type JSEvalUsage struct{}

func (j *JSEvalUsage) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.eval-usage",
		Category:    "bad-defaults",
		Name:        "eval() or new Function() usage",
		Description: "eval() or new Function() in production code — security and debuggability risk",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSEvalUsage) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding

	for _, node := range jsFindNodesByType(tree.RootNode(), "call_expression", "new_expression") {
		fn := node.ChildByFieldName("function")
		if fn == nil {
			// new_expression uses "constructor" field
			fn = node.ChildByFieldName("constructor")
		}
		if fn == nil {
			continue
		}

		fnText := jsNodeText(fn, source)

		switch {
		case node.Type() == "call_expression" && fnText == "eval":
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "eval() executes arbitrary code — security and debuggability risk",
				Suggestion: "Use JSON.parse(), a template engine, or a safer alternative",
			})
		case node.Type() == "new_expression" && fnText == "Function":
			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "new Function() executes arbitrary code — security and debuggability risk",
				Suggestion: "Use a safer alternative to dynamic code generation",
			})
		}
	}

	return findings
}
