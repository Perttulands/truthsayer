package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyLogAndRaise detects `logging.error(e); raise` — duplicate error reporting.
type PyLogAndRaise struct{}

func (p *PyLogAndRaise) Meta() Rule {
	return Rule{
		ID:          "error-context.py-log-and-raise",
		Category:    "error-context",
		Name:        "Log and raise duplicate reporting",
		Description: "Logging and re-raising creates duplicate error reports — the traceback already contains the error",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

// loggingMethods are logging methods that indicate error logging.
var loggingMethods = map[string]bool{
	"error":     true,
	"exception": true,
	"critical":  true,
	"fatal":     true,
	"warning":   true,
}

func (p *PyLogAndRaise) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, exceptNode := range pyFindNodesByType(tree.RootNode(), "except_clause") {
		body := exceptBody(exceptNode)
		if body == nil {
			continue
		}
		hasLogCall := false
		hasRaise := false
		var logLine int

		for i := 0; i < int(body.NamedChildCount()); i++ {
			child := body.NamedChild(i)
			if child.Type() == "expression_statement" && child.NamedChildCount() > 0 {
				expr := child.NamedChild(0)
				if isLoggingErrorCall(expr, source) {
					hasLogCall = true
					logLine = pyLineNumber(child)
				}
			}
			if child.Type() == "raise_statement" {
				hasRaise = true
			}
		}

		if hasLogCall && hasRaise {
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       logLine,
				Code:       pySourceLine(source, logLine),
				Message:    "Logging and re-raising creates duplicate error reports",
				Suggestion: "Either log the error or re-raise it, not both — the traceback already contains the error",
			})
		}
	}
	return findings
}

// isLoggingErrorCall checks if a node is a logging.error/exception/critical/etc call.
func isLoggingErrorCall(node *sitter.Node, source []byte) bool {
	if node.Type() != "call" {
		return false
	}
	fn := node.ChildByFieldName("function")
	if fn == nil || fn.Type() != "attribute" {
		return false
	}
	obj := fn.ChildByFieldName("object")
	attr := fn.ChildByFieldName("attribute")
	if obj == nil || attr == nil {
		return false
	}
	objText := pyNodeText(obj, source)
	attrText := pyNodeText(attr, source)
	// Match logging.error(), logger.error(), log.error(), etc.
	return (objText == "logging" || objText == "logger" || objText == "log") && loggingMethods[attrText]
}
