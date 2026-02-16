package scanner

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// FindNodesByType walks the tree and returns all nodes matching any of the given types.
func FindNodesByType(root *sitter.Node, types ...string) []*sitter.Node {
	typeSet := make(map[string]struct{}, len(types))
	for _, t := range types {
		typeSet[t] = struct{}{}
	}
	var result []*sitter.Node
	walkNode(root, func(n *sitter.Node) {
		if _, ok := typeSet[n.Type()]; ok {
			result = append(result, n)
		}
	})
	return result
}

// NodeText extracts the source text for a given node.
func NodeText(node *sitter.Node, source []byte) string {
	return node.Content(source)
}

// LineNumber returns the 1-indexed line number for a node.
func LineNumber(node *sitter.Node) int {
	return int(node.StartPoint().Row) + 1
}

// SourceLine extracts a single source line (1-indexed) from the source bytes.
func SourceLine(source []byte, line int) string {
	lines := strings.Split(string(source), "\n")
	idx := line - 1
	if idx < 0 || idx >= len(lines) {
		return ""
	}
	return lines[idx]
}

// HasChildOfType checks if a node has any direct child of the given type.
func HasChildOfType(node *sitter.Node, childType string) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		if node.Child(i).Type() == childType {
			return true
		}
	}
	return false
}

// IsInsideFunction checks if a node is inside a function or method definition.
// Works for JS (function_declaration, arrow_function, method_definition)
// and Python (function_definition).
func IsInsideFunction(node *sitter.Node) bool {
	funcTypes := map[string]struct{}{
		"function_declaration": {},
		"function":             {},
		"arrow_function":       {},
		"method_definition":    {},
		"function_definition":  {},
		"generator_function":   {},
	}
	for p := node.Parent(); p != nil; p = p.Parent() {
		if _, ok := funcTypes[p.Type()]; ok {
			return true
		}
	}
	return false
}

// IsTestFile determines if a path is a test file for JS/TS or Python.
func IsTestFile(path string) bool {
	// Check if path contains test directories
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts {
		if p == "__tests__" || p == "test" || p == "tests" {
			return true
		}
	}

	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Python test files: test_*.py, *_test.py
	if ext == ".py" || ext == ".pyi" {
		return strings.HasPrefix(name, "test_") || strings.HasSuffix(name, "_test")
	}

	// JS/TS test files: *.test.js, *.spec.js, *.test.ts, *.spec.ts, etc.
	jsExts := map[string]bool{".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true}
	if jsExts[ext] {
		return strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".spec") ||
			strings.HasSuffix(name, "_test") || strings.HasSuffix(name, "_spec")
	}

	return false
}

// walkNode recursively visits every node in the tree, calling fn for each.
func walkNode(node *sitter.Node, fn func(*sitter.Node)) {
	if node == nil {
		return
	}
	fn(node)
	for i := 0; i < int(node.NamedChildCount()); i++ {
		walkNode(node.NamedChild(i), fn)
	}
}
