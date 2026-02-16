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
