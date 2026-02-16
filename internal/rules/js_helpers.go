package rules

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// jsFindNodesByType walks the tree and returns all nodes matching any of the given types.
func jsFindNodesByType(root *sitter.Node, types ...string) []*sitter.Node {
	typeSet := make(map[string]struct{}, len(types))
	for _, t := range types {
		typeSet[t] = struct{}{}
	}
	var result []*sitter.Node
	jsWalkNode(root, func(n *sitter.Node) {
		if _, ok := typeSet[n.Type()]; ok {
			result = append(result, n)
		}
	})
	return result
}

// jsNodeText extracts the source text for a given node.
func jsNodeText(node *sitter.Node, source []byte) string {
	return node.Content(source)
}

// jsLineNumber returns the 1-indexed line number for a node.
func jsLineNumber(node *sitter.Node) int {
	return int(node.StartPoint().Row) + 1
}

// jsSourceLine extracts a single source line (1-indexed) from the source bytes.
func jsSourceLine(source []byte, line int) string {
	lines := strings.Split(string(source), "\n")
	idx := line - 1
	if idx < 0 || idx >= len(lines) {
		return ""
	}
	return lines[idx]
}

// jsIsTestFile determines if a path is a JS/TS test file.
func jsIsTestFile(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts {
		if p == "__tests__" || p == "test" || p == "tests" {
			return true
		}
	}

	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	jsExts := map[string]bool{".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true}
	if jsExts[ext] {
		return strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".spec") ||
			strings.HasSuffix(name, "_test") || strings.HasSuffix(name, "_spec")
	}
	return false
}

// jsUnquote strips surrounding quotes from a JS/TS string literal.
func jsUnquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'' || s[0] == '`') {
		return s[1 : len(s)-1]
	}
	return s
}

// expressAppInfo holds detected Express/Koa app variable names and route info.
type expressAppInfo struct {
	appNames       map[string]bool
	hasRoute       bool
	firstRouteNode *sitter.Node
}

// jsFindExpressApp scans for express()/Koa() variable declarations and route definitions.
func jsFindExpressApp(root *sitter.Node, source []byte) *expressAppInfo {
	calls := jsFindNodesByType(root, "call_expression")

	var appVarNames []string
	for _, call := range calls {
		fn := call.ChildByFieldName("function")
		if fn == nil {
			continue
		}
		fnText := jsNodeText(fn, source)
		if fnText == "express" || fnText == "Koa" {
			parent := call.Parent()
			if parent != nil && parent.Type() == "variable_declarator" {
				nameNode := parent.ChildByFieldName("name")
				if nameNode != nil {
					appVarNames = append(appVarNames, jsNodeText(nameNode, source))
				}
			}
		}
	}

	if len(appVarNames) == 0 {
		return nil
	}

	info := &expressAppInfo{
		appNames: make(map[string]bool, len(appVarNames)),
	}
	for _, n := range appVarNames {
		info.appNames[n] = true
	}

	routeMethods := map[string]bool{
		"get": true, "post": true, "put": true, "delete": true,
		"patch": true, "all": true,
	}

	for _, call := range calls {
		fn := call.ChildByFieldName("function")
		if fn == nil || fn.Type() != "member_expression" {
			continue
		}
		obj := fn.ChildByFieldName("object")
		prop := fn.ChildByFieldName("property")
		if obj == nil || prop == nil {
			continue
		}
		if !info.appNames[jsNodeText(obj, source)] {
			continue
		}
		if routeMethods[jsNodeText(prop, source)] {
			if !info.hasRoute {
				info.hasRoute = true
				info.firstRouteNode = call
			}
		}
	}

	return info
}

// jsWalkNode recursively visits every named node in the tree, calling fn for each.
func jsWalkNode(node *sitter.Node, fn func(*sitter.Node)) {
	if node == nil {
		return
	}
	fn(node)
	for i := 0; i < int(node.NamedChildCount()); i++ {
		jsWalkNode(node.NamedChild(i), fn)
	}
}
