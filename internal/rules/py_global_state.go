package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyGlobalState detects module-level mutable global assignments like ITEMS = [] or CACHE = {}.
type PyGlobalState struct{}

func (p *PyGlobalState) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.py-global-state",
		Category:    "bad-defaults",
		Name:        "Module-level mutable global",
		Description: "Module-level mutable global (list, dict, set) can cause hard-to-debug shared state issues",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyGlobalState) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if pyIsTestFile(path) {
		return nil
	}
	var findings []finding.Finding
	root := tree.RootNode()
	// Only check top-level (module-level) assignments — direct children of module.
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		if child.Type() != "expression_statement" {
			continue
		}
		// expression_statement wraps an assignment
		for j := 0; j < int(child.NamedChildCount()); j++ {
			assign := child.NamedChild(j)
			if assign.Type() != "assignment" {
				continue
			}
			right := assign.ChildByFieldName("right")
			if right == nil {
				continue
			}
			if !isMutableDefault(right, source) {
				continue
			}
			left := assign.ChildByFieldName("left")
			name := ""
			if left != nil {
				name = pyNodeText(left, source)
			}
			line := pyLineNumber(assign)
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       pySourceLine(source, line),
				Message:    "Module-level mutable global " + name + " = " + pyNodeText(right, source) + " can cause shared state bugs",
				Suggestion: "Document the purpose of this global, or use a function/class to encapsulate state",
			})
		}
	}
	return findings
}
