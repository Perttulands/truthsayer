package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyMutableDefault detects mutable default arguments like def f(items=[]).
type PyMutableDefault struct{}

func (p *PyMutableDefault) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.py-mutable-default-arg",
		Category:    "bad-defaults",
		Name:        "Mutable default argument",
		Description: "Mutable default argument (list, dict, set) is shared across calls and can cause subtle bugs",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

// mutableDefaults are tree-sitter node types for mutable literals.
var mutableDefaults = map[string]bool{
	"list":       true,
	"dictionary": true,
	"set":        true,
}

// mutableCallNames maps function names that produce mutable objects.
// tree-sitter parses set(), list(), dict() as call nodes, not literal nodes.
var mutableCallNames = map[string]bool{
	"set":  true,
	"list": true,
	"dict": true,
}

func (p *PyMutableDefault) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "function_definition") {
		params := node.ChildByFieldName("parameters")
		if params == nil {
			continue
		}
		for i := 0; i < int(params.NamedChildCount()); i++ {
			child := params.NamedChild(i)
			if child.Type() != "default_parameter" {
				continue
			}
			value := child.ChildByFieldName("value")
			if value == nil {
				continue
			}
			if isMutableDefault(value, source) {
				line := pyLineNumber(child)
				findings = append(findings, finding.Finding{
					Rule:       p.Meta().ID,
					Severity:   p.Meta().Severity,
					File:       path,
					Line:       line,
					Code:       pySourceLine(source, line),
					Message:    "Mutable default argument " + pyNodeText(value, source) + " is shared across all calls",
					Suggestion: "Use None as default and create the mutable inside the function body",
				})
			}
		}
	}
	return findings
}

// isMutableDefault returns true if the node is a mutable literal ([], {}, {1,2})
// or a call to set(), list(), or dict().
func isMutableDefault(node *sitter.Node, source []byte) bool {
	if mutableDefaults[node.Type()] {
		return true
	}
	if node.Type() == "call" {
		fn := node.ChildByFieldName("function")
		if fn != nil && fn.Type() == "identifier" {
			name := string(source[fn.StartByte():fn.EndByte()])
			return mutableCallNames[name]
		}
	}
	return false
}
