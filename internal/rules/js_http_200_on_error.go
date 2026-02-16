package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSHTTP200OnError detects res.status(200) or res.json() inside catch blocks
// in Express/Koa/Fastify handlers — sending success responses after errors.
type JSHTTP200OnError struct{}

func (j *JSHTTP200OnError) Meta() Rule {
	return Rule{
		ID:          "error-context.js-http-200-on-error",
		Category:    "error-context",
		Name:        "HTTP 200 on error",
		Description: "Express/Koa/Fastify handler sends success status inside catch block",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSHTTP200OnError) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, catchNode := range jsFindNodesByType(tree.RootNode(), "catch_clause") {
		body := catchNode.ChildByFieldName("body")
		if body == nil {
			continue
		}

		jsWalkNode(body, func(n *sitter.Node) {
			if n.Type() != "call_expression" {
				return
			}
			fn := n.ChildByFieldName("function")
			if fn == nil {
				return
			}

			if isSuccessStatusCall(fn, source) || isResJsonCall(fn, n, source) {
				line := jsLineNumber(n)
				findings = append(findings, finding.Finding{
					Rule:       j.Meta().ID,
					Severity:   j.Meta().Severity,
					File:       path,
					Line:       line,
					Code:       jsSourceLine(source, line),
					Message:    "Sending success response inside catch block — error is masked from the client",
					Suggestion: "Send an error status (e.g., res.status(500).json({ error: ... })) in catch blocks",
				})
			}
		})
	}
	return findings
}

// isSuccessStatusCall checks for res.status(200), res.status(201), etc. (2xx)
func isSuccessStatusCall(fn *sitter.Node, source []byte) bool {
	if fn.Type() != "member_expression" {
		return false
	}
	prop := fn.ChildByFieldName("property")
	if prop == nil || jsNodeText(prop, source) != "status" {
		return false
	}

	// Check the argument is a 2xx status code
	callNode := fn.Parent()
	if callNode == nil {
		return false
	}
	args := callNode.ChildByFieldName("arguments")
	if args == nil || args.NamedChildCount() == 0 {
		return false
	}
	arg := args.NamedChild(0)
	if arg == nil {
		return false
	}
	code := jsNodeText(arg, source)
	return code == "200" || code == "201" || code == "202" || code == "204"
}

// isResJsonCall checks for res.json(...) — sending JSON response in catch (without prior error status).
// Only flags res.json(), not res.status(4xx).json() chains.
func isResJsonCall(fn *sitter.Node, callNode *sitter.Node, source []byte) bool {
	if fn.Type() != "member_expression" {
		return false
	}
	prop := fn.ChildByFieldName("property")
	if prop == nil || jsNodeText(prop, source) != "json" {
		return false
	}
	obj := fn.ChildByFieldName("object")
	if obj == nil {
		return false
	}

	// If the object is res.status(4xx/5xx), this is proper error handling — skip.
	// Pattern: res.status(500).json(...)  →  obj is a call_expression for res.status(500)
	if obj.Type() == "call_expression" {
		innerFn := obj.ChildByFieldName("function")
		if innerFn != nil && innerFn.Type() == "member_expression" {
			innerProp := innerFn.ChildByFieldName("property")
			if innerProp != nil && jsNodeText(innerProp, source) == "status" {
				args := obj.ChildByFieldName("arguments")
				if args != nil && args.NamedChildCount() > 0 {
					code := jsNodeText(args.NamedChild(0), source)
					if isErrorStatusCode(code) {
						return false
					}
				}
			}
		}
	}

	// Only flag if object is a simple identifier (res, ctx, response)
	if obj.Type() == "identifier" {
		name := jsNodeText(obj, source)
		return name == "res" || name == "ctx" || name == "response"
	}
	return false
}

func isErrorStatusCode(code string) bool {
	if len(code) != 3 {
		return false
	}
	return code[0] == '4' || code[0] == '5'
}
