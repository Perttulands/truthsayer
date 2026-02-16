package rules

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyUnittestImport detects unittest.mock or mock imports in non-test files.
type PyUnittestImport struct{}

func (p *PyUnittestImport) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.py-unittest-import",
		Category:    "mock-leakage",
		Name:        "unittest.mock import in source",
		Description: "unittest.mock or mock import in non-test file leaks test infrastructure into production code",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

// pyTestModules are module names that indicate test-only imports.
var pyTestModules = []string{
	"unittest.mock",
	"unittest",
	"mock",
	"pytest",
}

func (p *PyUnittestImport) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if pyIsTestFile(path) {
		return nil
	}

	root := tree.RootNode()
	var findings []finding.Finding

	// Check import_from_statement: from unittest.mock import patch, MagicMock
	for _, node := range pyFindNodesByType(root, "import_from_statement") {
		modNode := node.ChildByFieldName("module_name")
		if modNode == nil {
			continue
		}
		modName := pyNodeText(modNode, source)
		if isPyTestModule(modName) {
			line := pyLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       pySourceLine(source, line),
				Message:    "Test module '" + modName + "' imported in non-test file",
				Suggestion: "Move this import and its usage to a test file",
			})
		}
	}

	// Check import_statement: import mock, import unittest.mock
	for _, node := range pyFindNodesByType(root, "import_statement") {
		// import_statement children are dotted_name or aliased_import
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			var modName string
			switch child.Type() {
			case "dotted_name":
				modName = pyNodeText(child, source)
			case "aliased_import":
				nameNode := child.ChildByFieldName("name")
				if nameNode != nil {
					modName = pyNodeText(nameNode, source)
				}
			}
			if modName != "" && isPyTestModule(modName) {
				line := pyLineNumber(node)
				findings = append(findings, finding.Finding{
					Rule:       p.Meta().ID,
					Severity:   p.Meta().Severity,
					File:       path,
					Line:       line,
					Code:       pySourceLine(source, line),
					Message:    "Test module '" + modName + "' imported in non-test file",
					Suggestion: "Move this import and its usage to a test file",
				})
			}
		}
	}

	return findings
}

func isPyTestModule(modName string) bool {
	for _, mod := range pyTestModules {
		if modName == mod || strings.HasPrefix(modName, mod+".") {
			return true
		}
	}
	return false
}
