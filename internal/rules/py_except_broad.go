package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyExceptBroad detects `except Exception:` — overly broad exception catching.
type PyExceptBroad struct{}

func (p *PyExceptBroad) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.py-except-broad",
		Category:    "silent-fallback",
		Name:        "Overly broad exception type",
		Description: "Catching Exception or BaseException is too broad and masks specific errors",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

var broadExceptions = map[string]bool{
	"Exception":     true,
	"BaseException": true,
}

func (p *PyExceptBroad) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "except_clause") {
		excType := exceptType(node, source)
		if broadExceptions[excType] {
			line := pyLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       pySourceLine(source, line),
				Message:    "Catching " + excType + " is too broad — specific exception types are safer",
				Suggestion: "Catch a more specific exception type like ValueError, TypeError, etc.",
			})
		}
	}
	return findings
}

// exceptType extracts the exception type name from an except_clause.
// Returns "" for bare except, the identifier text for simple types,
// and "" for tuple exceptions (which are not overly broad by definition).
func exceptType(node *sitter.Node, source []byte) string {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "identifier":
			return pyNodeText(child, source)
		case "as_pattern":
			// except Exception as e: — the exception type is the first child
			if child.NamedChildCount() > 0 {
				first := child.NamedChild(0)
				if first.Type() == "identifier" {
					return pyNodeText(first, source)
				}
			}
			return ""
		case "block":
			continue
		}
	}
	return ""
}
