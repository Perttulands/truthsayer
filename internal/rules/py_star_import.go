package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyStarImport detects `from module import *` wildcard imports.
type PyStarImport struct{}

func (p *PyStarImport) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.py-star-import",
		Category:    "bad-defaults",
		Name:        "Wildcard import",
		Description: "from module import * pollutes the namespace and hides dependencies",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyStarImport) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if pyIsTestFile(path) {
		return nil
	}
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "import_from_statement") {
		if hasWildcardImport(node) {
			line := pyLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       pySourceLine(source, line),
				Message:    "Wildcard import pollutes namespace and hides dependencies",
				Suggestion: "Import specific names: from module import name1, name2",
			})
		}
	}
	return findings
}

// hasWildcardImport checks if an import_from_statement contains a wildcard_import child.
func hasWildcardImport(node *sitter.Node) bool {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		if node.NamedChild(i).Type() == "wildcard_import" {
			return true
		}
	}
	return false
}
