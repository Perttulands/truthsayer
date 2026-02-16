package rules

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// pyFindNodesByType walks the tree and returns all nodes matching any of the given types.
func pyFindNodesByType(root *sitter.Node, types ...string) []*sitter.Node {
	typeSet := make(map[string]struct{}, len(types))
	for _, t := range types {
		typeSet[t] = struct{}{}
	}
	var result []*sitter.Node
	pyWalkNode(root, func(n *sitter.Node) {
		if _, ok := typeSet[n.Type()]; ok {
			result = append(result, n)
		}
	})
	return result
}

// pyNodeText extracts the source text for a given node.
func pyNodeText(node *sitter.Node, source []byte) string {
	return node.Content(source)
}

// pyLineNumber returns the 1-indexed line number for a node.
func pyLineNumber(node *sitter.Node) int {
	return int(node.StartPoint().Row) + 1
}

// pySourceLine extracts a single source line (1-indexed) from the source bytes.
func pySourceLine(source []byte, line int) string {
	lines := strings.Split(string(source), "\n")
	idx := line - 1
	if idx < 0 || idx >= len(lines) {
		return ""
	}
	return lines[idx]
}

// pyIsTestFile determines if a path is a Python test file.
func pyIsTestFile(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts {
		if p == "test" || p == "tests" || p == "__tests__" {
			return true
		}
	}

	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != ".py" {
		return false
	}
	name := strings.TrimSuffix(base, ext)
	return strings.HasPrefix(name, "test_") || strings.HasSuffix(name, "_test") ||
		name == "conftest"
}

// directChildrenOfType returns the direct named children of node matching the given type.
// Unlike pyFindNodesByType, this does NOT recurse into grandchildren.
func directChildrenOfType(node *sitter.Node, nodeType string) []*sitter.Node {
	var result []*sitter.Node
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == nodeType {
			result = append(result, child)
		}
	}
	return result
}

// pyWalkNode recursively visits every named node in the tree, calling fn for each.
func pyWalkNode(node *sitter.Node, fn func(*sitter.Node)) {
	if node == nil {
		return
	}
	fn(node)
	for i := 0; i < int(node.NamedChildCount()); i++ {
		pyWalkNode(node.NamedChild(i), fn)
	}
}
