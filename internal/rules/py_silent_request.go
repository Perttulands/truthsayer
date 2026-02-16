package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PySilentRequest detects requests.get/post() without raise_for_status() or status_code check.
type PySilentRequest struct{}

func (p *PySilentRequest) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.py-silent-request",
		Category:    "trace-gaps",
		Name:        "HTTP request without status check",
		Description: "requests.get/post() without raise_for_status() or status_code check silently ignores HTTP errors",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PySilentRequest) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding

	// Find all call nodes that are requests.get/post/etc.
	for _, node := range pyFindNodesByType(tree.RootNode(), "call") {
		fn := node.ChildByFieldName("function")
		if fn == nil || fn.Type() != "attribute" {
			continue
		}
		obj := fn.ChildByFieldName("object")
		attr := fn.ChildByFieldName("attribute")
		if obj == nil || attr == nil {
			continue
		}
		if pyNodeText(obj, source) != "requests" || !requestsMethods[pyNodeText(attr, source)] {
			continue
		}

		// Find the enclosing scope (function body, module, etc.) and check
		// if the result is used with raise_for_status() or status_code
		if hasStatusCheck(node, source) {
			continue
		}

		line := pyLineNumber(node)
		findings = append(findings, finding.Finding{
			Rule:       p.Meta().ID,
			Severity:   p.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       pySourceLine(source, line),
			Message:    "requests." + pyNodeText(attr, source) + "() without response status check",
			Suggestion: "Add response.raise_for_status() or check response.status_code after the call",
		})
	}
	return findings
}

// hasStatusCheck checks if a requests call has a corresponding raise_for_status()
// or status_code check in the same scope.
func hasStatusCheck(callNode *sitter.Node, source []byte) bool {
	// Check if this call is inside a method chain like requests.get(url).raise_for_status()
	parent := callNode.Parent()
	if parent != nil && parent.Type() == "attribute" {
		attrNode := parent.ChildByFieldName("attribute")
		if attrNode != nil {
			attrName := pyNodeText(attrNode, source)
			if attrName == "raise_for_status" || attrName == "status_code" || attrName == "ok" {
				return true
			}
		}
	}

	// Walk up to find the enclosing expression_statement (the statement-level node)
	stmt := callNode
	for stmt != nil && stmt.Type() != "expression_statement" {
		stmt = stmt.Parent()
	}
	if stmt == nil {
		return false
	}

	// Check if the expression_statement contains an assignment
	// Tree structure: expression_statement > assignment > call
	for i := 0; i < int(stmt.NamedChildCount()); i++ {
		child := stmt.NamedChild(i)
		if child.Type() == "assignment" {
			varNode := child.ChildByFieldName("left")
			if varNode != nil {
				varName := pyNodeText(varNode, source)
				return hasResponseCheckInSiblings(stmt, varName, source)
			}
		}
	}

	return false
}

// hasResponseCheckInSiblings looks at sibling statements after the given node
// to find raise_for_status() or status_code checks on the named variable.
func hasResponseCheckInSiblings(node *sitter.Node, varName string, source []byte) bool {
	parent := node.Parent()
	if parent == nil {
		return false
	}

	// Find this node among siblings and check subsequent ones
	found := false
	for i := 0; i < int(parent.NamedChildCount()); i++ {
		sibling := parent.NamedChild(i)
		if sibling == node {
			found = true
			continue
		}
		if !found {
			continue
		}
		// Check subsequent siblings for varName.raise_for_status() or varName.status_code
		if containsResponseCheck(sibling, varName, source) {
			return true
		}
	}
	return false
}

// containsResponseCheck recursively checks if a node contains a reference to
// varName.raise_for_status(), varName.status_code, or varName.ok.
func containsResponseCheck(node *sitter.Node, varName string, source []byte) bool {
	if node == nil {
		return false
	}
	if node.Type() == "attribute" {
		obj := node.ChildByFieldName("object")
		attr := node.ChildByFieldName("attribute")
		if obj != nil && attr != nil && pyNodeText(obj, source) == varName {
			attrName := pyNodeText(attr, source)
			if attrName == "raise_for_status" || attrName == "status_code" || attrName == "ok" {
				return true
			}
		}
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		if containsResponseCheck(node.NamedChild(i), varName, source) {
			return true
		}
	}
	return false
}
