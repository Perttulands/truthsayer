package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// Known promise-returning function names.
var promiseFuncs = map[string]bool{
	"fetch": true,
}

// JSFloatingPromise detects unhandled promise-returning calls (not awaited, assigned, or returned).
type JSFloatingPromise struct{}

func (j *JSFloatingPromise) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.js-floating-promise",
		Category:    "silent-fallback",
		Name:        "Floating promise",
		Description: "Promise-returning call is not awaited, assigned, or returned",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSFloatingPromise) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	// Look for expression_statement containing a call to a known promise function
	for _, stmt := range jsFindNodesByType(tree.RootNode(), "expression_statement") {
		expr := stmt.NamedChild(0)
		if expr == nil {
			continue
		}
		if isFloatingPromiseCall(expr, source) {
			line := jsLineNumber(stmt)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Promise-returning call is not awaited, assigned, or returned",
				Suggestion: "Use await, assign to a variable, or chain .then()/.catch()",
			})
		}
	}
	return findings
}

// isFloatingPromiseCall checks if an expression node is an unhandled call to a
// known promise function. Returns false if the call is chained with .then/.catch.
func isFloatingPromiseCall(node *sitter.Node, source []byte) bool {
	if node.Type() == "call_expression" {
		return isKnownPromiseCall(node, source)
	}
	return false
}

// isKnownPromiseCall checks if a call_expression calls a known promise-returning function.
// It also ensures the call is not chained with .then/.catch (which would be a member_expression parent).
func isKnownPromiseCall(node *sitter.Node, source []byte) bool {
	fn := node.ChildByFieldName("function")
	if fn == nil {
		return false
	}

	// Direct call: fetch(...)
	if fn.Type() == "identifier" {
		name := jsNodeText(fn, source)
		return promiseFuncs[name]
	}

	// Chained call: fetch(...).then(...) — the parent is a member_expression
	// In this case the call_expression for fetch() is inside a member_expression,
	// which is inside another call_expression. We should NOT flag this.
	// But actually in tree-sitter: fetch("/api").then(fn)
	// is: call_expression(member_expression(call_expression(fetch), then), arguments)
	// So if fn is a member_expression where the object is a call to a promise func,
	// this outer call is the .then/.catch — we skip it.
	if fn.Type() == "member_expression" {
		prop := fn.ChildByFieldName("property")
		if prop != nil {
			propName := jsNodeText(prop, source)
			if propName == "then" || propName == "catch" || propName == "finally" {
				return false
			}
		}
		// Check if this is something.fetch(...)
		if prop != nil && promiseFuncs[jsNodeText(prop, source)] {
			return true
		}
	}

	return false
}
