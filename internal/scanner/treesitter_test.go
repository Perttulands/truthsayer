package scanner

import (
	"strings"
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

// --- Utility function tests ---

func parseJS(t *testing.T, src string) *sitter.Node {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, []byte(src))
	if err != nil {
		t.Fatalf("JS parse failed: %v", err)
	}
	return tree.RootNode()
}

func parsePython(t *testing.T, src string) *sitter.Node {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, []byte(src))
	if err != nil {
		t.Fatalf("Python parse failed: %v", err)
	}
	return tree.RootNode()
}

func TestFindNodesByType_JS(t *testing.T) {
	src := `try { doStuff(); } catch (e) { handle(e); }`
	root := parseJS(t, src)

	nodes := FindNodesByType(root, "catch_clause")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 catch_clause, got %d", len(nodes))
	}
	if nodes[0].Type() != "catch_clause" {
		t.Errorf("expected type catch_clause, got %q", nodes[0].Type())
	}
}

func TestFindNodesByType_MultipleTypes(t *testing.T) {
	src := `function foo() { return 1; }
const bar = () => 2;`
	root := parseJS(t, src)

	nodes := FindNodesByType(root, "function_declaration", "arrow_function")
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestFindNodesByType_NoMatch(t *testing.T) {
	src := `const x = 1;`
	root := parseJS(t, src)

	nodes := FindNodesByType(root, "catch_clause")
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(nodes))
	}
}

func TestFindNodesByType_Python(t *testing.T) {
	src := `try:
    do_stuff()
except ValueError:
    pass
except TypeError:
    handle()`
	root := parsePython(t, src)

	nodes := FindNodesByType(root, "except_clause")
	if len(nodes) != 2 {
		t.Fatalf("expected 2 except_clause nodes, got %d", len(nodes))
	}
}

func TestNodeText(t *testing.T) {
	src := `const greeting = "hello";`
	root := parseJS(t, src)

	// The root should be "program", find the string node
	nodes := FindNodesByType(root, "string")
	if len(nodes) == 0 {
		t.Fatal("expected at least 1 string node")
	}
	text := NodeText(nodes[0], []byte(src))
	if text != `"hello"` {
		t.Errorf("expected %q, got %q", `"hello"`, text)
	}
}

func TestLineNumber(t *testing.T) {
	src := `line1
line2
function foo() {}`
	root := parseJS(t, src)

	nodes := FindNodesByType(root, "function_declaration")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 function_declaration, got %d", len(nodes))
	}
	line := LineNumber(nodes[0])
	if line != 3 {
		t.Errorf("expected line 3, got %d", line)
	}
}

func TestLineNumber_FirstLine(t *testing.T) {
	src := `const x = 1;`
	root := parseJS(t, src)

	line := LineNumber(root.NamedChild(0))
	if line != 1 {
		t.Errorf("expected line 1, got %d", line)
	}
}

func TestSourceLine(t *testing.T) {
	src := []byte("first\nsecond\nthird")

	tests := []struct {
		line int
		want string
	}{
		{1, "first"},
		{2, "second"},
		{3, "third"},
		{0, ""},  // out of bounds
		{4, ""},  // out of bounds
		{-1, ""}, // negative
	}
	for _, tt := range tests {
		got := SourceLine(src, tt.line)
		if got != tt.want {
			t.Errorf("SourceLine(src, %d) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestHasChildOfType(t *testing.T) {
	src := `try { doStuff(); } catch (e) { handle(e); }`
	root := parseJS(t, src)

	// try_statement should have catch_clause child
	tryNodes := FindNodesByType(root, "try_statement")
	if len(tryNodes) != 1 {
		t.Fatalf("expected 1 try_statement, got %d", len(tryNodes))
	}
	if !HasChildOfType(tryNodes[0], "catch_clause") {
		t.Error("try_statement should have catch_clause child")
	}
	if HasChildOfType(tryNodes[0], "finally_clause") {
		t.Error("try_statement should not have finally_clause child")
	}
}

func TestHasChildOfType_Python(t *testing.T) {
	src := `try:
    do_stuff()
except ValueError:
    handle()`
	root := parsePython(t, src)

	tryNodes := FindNodesByType(root, "try_statement")
	if len(tryNodes) != 1 {
		t.Fatalf("expected 1 try_statement, got %d", len(tryNodes))
	}
	if !HasChildOfType(tryNodes[0], "except_clause") {
		t.Error("try_statement should have except_clause child")
	}
}

func TestIsInsideFunction_JS(t *testing.T) {
	src := `function outer() {
	const x = 1;
}
const y = 2;`
	root := parseJS(t, src)

	// Find all number nodes — one inside function, one outside
	numbers := FindNodesByType(root, "number")
	if len(numbers) != 2 {
		t.Fatalf("expected 2 number nodes, got %d", len(numbers))
	}

	// First number (1) is inside function
	if !IsInsideFunction(numbers[0]) {
		t.Error("number 1 should be inside function")
	}
	// Second number (2) is outside function
	if IsInsideFunction(numbers[1]) {
		t.Error("number 2 should not be inside function")
	}
}

func TestIsInsideFunction_ArrowFunction(t *testing.T) {
	src := `const fn = () => { const x = 1; }`
	root := parseJS(t, src)

	numbers := FindNodesByType(root, "number")
	if len(numbers) != 1 {
		t.Fatalf("expected 1 number node, got %d", len(numbers))
	}
	if !IsInsideFunction(numbers[0]) {
		t.Error("number should be inside arrow function")
	}
}

func TestIsInsideFunction_Python(t *testing.T) {
	src := `def foo():
    x = 1
y = 2`
	root := parsePython(t, src)

	assignments := FindNodesByType(root, "assignment")
	if len(assignments) < 2 {
		t.Fatalf("expected at least 2 assignments, got %d", len(assignments))
	}

	if !IsInsideFunction(assignments[0]) {
		t.Error("first assignment should be inside function")
	}
	if IsInsideFunction(assignments[1]) {
		t.Error("second assignment should not be inside function")
	}
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// JS/TS test files
		{"src/app.test.js", true},
		{"src/app.spec.ts", true},
		{"src/app.test.tsx", true},
		{"src/app_test.js", true},
		{"src/app_spec.ts", true},
		{"__tests__/app.js", true},

		// Python test files
		{"tests/test_app.py", true},
		{"tests/app_test.py", true},

		// Non-test files
		{"src/app.js", false},
		{"src/app.ts", false},
		{"src/app.py", false},
		{"src/testing.js", false},
		{"src/contest.py", false},
		{"src/latest.ts", false},

		// Directory-based test detection
		{"test/helpers.js", true},
		{"tests/conftest.py", true},
	}

	for _, tt := range tests {
		got := IsTestFile(tt.path)
		if got != tt.want {
			t.Errorf("IsTestFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestNodeText_Python(t *testing.T) {
	src := `name = "hello"`
	root := parsePython(t, src)

	strings_ := FindNodesByType(root, "string")
	if len(strings_) == 0 {
		t.Fatal("expected at least 1 string node")
	}
	text := NodeText(strings_[0], []byte(src))
	if !strings.Contains(text, "hello") {
		t.Errorf("expected text containing 'hello', got %q", text)
	}
}
