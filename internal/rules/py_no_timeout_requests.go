package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyNoTimeoutRequests detects requests.get/post/put/delete/patch() without timeout=.
type PyNoTimeoutRequests struct{}

func (p *PyNoTimeoutRequests) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.py-no-timeout-requests",
		Category:    "bad-defaults",
		Name:        "requests without timeout",
		Description: "requests.get/post() without timeout= parameter can block indefinitely",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

// requestsMethods are HTTP methods on the requests module.
var requestsMethods = map[string]bool{
	"get":     true,
	"post":    true,
	"put":     true,
	"delete":  true,
	"patch":   true,
	"head":    true,
	"options": true,
	"request": true,
}

func (p *PyNoTimeoutRequests) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
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
		if hasKeywordArg(node, source, "timeout") {
			continue
		}
		line := pyLineNumber(node)
		findings = append(findings, finding.Finding{
			Rule:       p.Meta().ID,
			Severity:   p.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       pySourceLine(source, line),
			Message:    "requests." + pyNodeText(attr, source) + "() without timeout parameter — can block indefinitely",
			Suggestion: "Add timeout= parameter, e.g. requests.get(url, timeout=30)",
		})
	}
	return findings
}

// hasKeywordArg checks if a call node has a keyword argument with the given name.
func hasKeywordArg(callNode *sitter.Node, source []byte, name string) bool {
	args := callNode.ChildByFieldName("arguments")
	if args == nil {
		return false
	}
	for i := 0; i < int(args.NamedChildCount()); i++ {
		child := args.NamedChild(i)
		if child.Type() == "keyword_argument" {
			argName := child.ChildByFieldName("name")
			if argName != nil && pyNodeText(argName, source) == name {
				return true
			}
		}
		// **kwargs spread might contain timeout
		if child.Type() == "dictionary_splat" {
			return true
		}
	}
	return false
}
