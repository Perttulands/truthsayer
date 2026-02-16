package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyBareExcept detects `except:` without an exception type.
type PyBareExcept struct{}

func (p *PyBareExcept) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.py-bare-except",
		Category:    "silent-fallback",
		Name:        "Bare except clause",
		Description: "except: without an exception type silently catches all exceptions including KeyboardInterrupt and SystemExit",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyBareExcept) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "except_clause") {
		if isBareExcept(node) {
			line := pyLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       pySourceLine(source, line),
				Message:    "Bare except catches all exceptions including KeyboardInterrupt and SystemExit",
				Suggestion: "Specify an exception type, e.g. except Exception:",
			})
		}
	}
	return findings
}

// isBareExcept checks if an except_clause has no exception type specified.
// A bare except has no named children between "except" keyword and ":" / block.
func isBareExcept(node *sitter.Node) bool {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "block":
			continue
		default:
			// Any named child that's not "block" is an exception type or as_pattern
			return false
		}
	}
	return true
}
