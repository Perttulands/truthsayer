package scanner

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

func TestTreeSitterJSParse(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	src := []byte(`function hello() { return "world"; }`)
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("JS parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.Type() != "program" {
		t.Errorf("expected root type 'program', got %q", root.Type())
	}
	if root.ChildCount() == 0 {
		t.Error("expected children in JS AST")
	}
}

func TestTreeSitterTSParse(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	src := []byte(`const greet = (name: string): string => name;`)
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("TS parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.Type() != "program" {
		t.Errorf("expected root type 'program', got %q", root.Type())
	}
	if root.ChildCount() == 0 {
		t.Error("expected children in TS AST")
	}
}

func TestTreeSitterTSXParse(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(tsx.GetLanguage())

	src := []byte(`const App = () => <div>Hello</div>;`)
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("TSX parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.Type() != "program" {
		t.Errorf("expected root type 'program', got %q", root.Type())
	}
	if root.ChildCount() == 0 {
		t.Error("expected children in TSX AST")
	}
}

func TestTreeSitterPythonParse(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	src := []byte(`def hello():\n    return "world"`)
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("Python parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.Type() != "module" {
		t.Errorf("expected root type 'module', got %q", root.Type())
	}
	if root.ChildCount() == 0 {
		t.Error("expected children in Python AST")
	}
}
