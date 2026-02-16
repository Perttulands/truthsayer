package scanner

import (
	"os"
	"sync"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// --- Modern JS syntax ---

func TestParseModernJS(t *testing.T) {
	src, err := os.ReadFile("../../testdata/js/syntax_modern.js")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	root := tree.RootNode()

	if root.Type() != "program" {
		t.Errorf("root type = %q, want 'program'", root.Type())
	}
	if root.HasError() {
		t.Error("parse tree has errors on valid modern JS")
	}
}

func TestParseJS_OptionalChaining(t *testing.T) {
	src := []byte(`const x = a?.b?.c;`)
	root := parseJS(t, string(src))

	nodes := FindNodesByType(root, "optional_chain_expression")
	if len(nodes) == 0 {
		// Some tree-sitter versions use "member_expression" with optional flag
		nodes = FindNodesByType(root, "member_expression")
	}
	if len(nodes) == 0 {
		t.Error("expected optional chaining nodes")
	}
}

func TestParseJS_NullishCoalescing(t *testing.T) {
	src := []byte(`const x = a ?? "default";`)
	root := parseJS(t, string(src))

	nodes := FindNodesByType(root, "binary_expression")
	if len(nodes) == 0 {
		t.Error("expected binary_expression for nullish coalescing")
	}
}

func TestParseJS_ClassFields(t *testing.T) {
	src := []byte(`class Foo {
  bar = 1;
  static baz = 2;
  #priv = 3;
}`)
	root := parseJS(t, string(src))

	classes := FindNodesByType(root, "class_declaration")
	if len(classes) != 1 {
		t.Fatalf("expected 1 class, got %d", len(classes))
	}
	if classes[0].HasError() {
		t.Error("class with fields should parse without errors")
	}
}

func TestParseJS_AsyncAwait(t *testing.T) {
	src := []byte(`async function fetchData() {
  const result = await fetch("/api");
  return result.json();
}`)
	root := parseJS(t, string(src))

	fns := FindNodesByType(root, "function_declaration")
	if len(fns) != 1 {
		t.Fatalf("expected 1 function, got %d", len(fns))
	}

	awaits := FindNodesByType(root, "await_expression")
	if len(awaits) != 1 {
		t.Errorf("expected 1 await_expression, got %d", len(awaits))
	}
}

func TestParseJS_DynamicImport(t *testing.T) {
	src := []byte(`const mod = import("./module.js");`)

	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("dynamic import should parse without errors")
	}
}

func TestParseJS_GeneratorFunction(t *testing.T) {
	src := []byte(`function* gen() { yield 1; yield 2; }`)
	root := parseJS(t, string(src))

	nodes := FindNodesByType(root, "generator_function_declaration")
	if len(nodes) == 0 {
		nodes = FindNodesByType(root, "generator_function")
	}
	if len(nodes) == 0 {
		t.Error("expected generator function node")
	}
}

// --- Modern TS syntax ---

func TestParseModernTS(t *testing.T) {
	src, err := os.ReadFile("../../testdata/js/syntax_modern.ts")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	root := tree.RootNode()

	if root.Type() != "program" {
		t.Errorf("root type = %q, want 'program'", root.Type())
	}
	if root.HasError() {
		t.Error("parse tree has errors on valid modern TS")
	}
}

func TestParseTS_Generics(t *testing.T) {
	src := []byte(`function identity<T>(arg: T): T { return arg; }`)

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.HasError() {
		t.Error("generic function should parse without errors")
	}

	nodes := FindNodesByType(root, "type_parameters")
	if len(nodes) == 0 {
		t.Error("expected type_parameters node for generics")
	}
}

func TestParseTS_Enum(t *testing.T) {
	src := []byte(`enum Color { Red = "RED", Blue = "BLUE" }`)

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	nodes := FindNodesByType(tree.RootNode(), "enum_declaration")
	if len(nodes) != 1 {
		t.Errorf("expected 1 enum_declaration, got %d", len(nodes))
	}
}

func TestParseTS_TypeAssertion(t *testing.T) {
	src := []byte(`const x = value as string;`)

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	nodes := FindNodesByType(tree.RootNode(), "as_expression")
	if len(nodes) != 1 {
		t.Errorf("expected 1 as_expression, got %d", len(nodes))
	}
}

func TestParseTS_Satisfies(t *testing.T) {
	src := []byte(`const x = { a: 1 } satisfies Record<string, number>;`)

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	root := tree.RootNode()
	nodes := FindNodesByType(root, "satisfies_expression")
	if len(nodes) != 1 {
		// satisfies may not be in all tree-sitter TS grammar versions
		if root.HasError() {
			t.Skip("tree-sitter TS grammar does not support satisfies operator")
		}
	}
}

func TestParseTS_Decorator(t *testing.T) {
	src := []byte(`function dec(t: any, k: string) {}
class Svc {
  @dec
  method() {}
}`)

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("decorator syntax should parse without errors")
	}
}

func TestParseTS_NonNullAssertion(t *testing.T) {
	src := []byte(`const el = document.querySelector(".x")!;`)

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	nodes := FindNodesByType(tree.RootNode(), "non_null_expression")
	if len(nodes) != 1 {
		t.Errorf("expected 1 non_null_expression, got %d", len(nodes))
	}
}

func TestParseTSX_JSXElements(t *testing.T) {
	src := []byte(`const App = () => <div className="app"><span>Hello</span></div>;`)

	parser := sitter.NewParser()
	parser.SetLanguage(tsx.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("JSX should parse without errors")
	}

	nodes := FindNodesByType(tree.RootNode(), "jsx_element")
	if len(nodes) == 0 {
		t.Error("expected jsx_element nodes")
	}
}

// --- Modern Python syntax ---

func TestParseModernPython(t *testing.T) {
	src, err := os.ReadFile("../../testdata/python/syntax_modern.py")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	root := tree.RootNode()

	if root.Type() != "module" {
		t.Errorf("root type = %q, want 'module'", root.Type())
	}
	if root.HasError() {
		t.Error("parse tree has errors on valid modern Python")
	}
}

func TestParsePython_MatchCase(t *testing.T) {
	src := []byte(`match command:
    case "quit":
        return False
    case "hello":
        return True
    case _:
        pass
`)

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.HasError() {
		t.Error("match/case should parse without errors")
	}

	nodes := FindNodesByType(root, "match_statement")
	if len(nodes) != 1 {
		t.Errorf("expected 1 match_statement, got %d", len(nodes))
	}

	cases := FindNodesByType(root, "case_clause")
	if len(cases) != 3 {
		t.Errorf("expected 3 case_clause nodes, got %d", len(cases))
	}
}

func TestParsePython_WalrusOperator(t *testing.T) {
	src := []byte(`if (n := len(data)) > 3:
    print(n)
`)

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("walrus operator should parse without errors")
	}

	nodes := FindNodesByType(tree.RootNode(), "named_expression")
	if len(nodes) != 1 {
		t.Errorf("expected 1 named_expression, got %d", len(nodes))
	}
}

func TestParsePython_FStrings(t *testing.T) {
	src := []byte(`name = "world"
msg = f"Hello, {name}!"
nested = f"result: {1 + 2}"
`)

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("f-strings should parse without errors")
	}
}

func TestParsePython_TypeHints(t *testing.T) {
	src := []byte(`def process(items: list[str], count: int = 0) -> dict[str, int]:
    return {item: count for item in items}
`)

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("type hints should parse without errors")
	}
}

func TestParsePython_AsyncAwait(t *testing.T) {
	src := []byte(`import asyncio

async def fetch(url: str) -> dict:
    await asyncio.sleep(0.1)
    return {"url": url}

async def main():
    result = await fetch("http://example.com")
`)

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("async/await should parse without errors")
	}

	nodes := FindNodesByType(tree.RootNode(), "await")
	if len(nodes) == 0 {
		t.Error("expected await nodes")
	}
}

func TestParsePython_Decorators(t *testing.T) {
	src := []byte(`from dataclasses import dataclass

@dataclass
class Point:
    x: float
    y: float
`)

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("decorators should parse without errors")
	}

	nodes := FindNodesByType(tree.RootNode(), "decorator")
	if len(nodes) != 1 {
		t.Errorf("expected 1 decorator, got %d", len(nodes))
	}
}

func TestParsePython_Comprehensions(t *testing.T) {
	src := []byte(`squares = [x**2 for x in range(10)]
evens = {x for x in range(20) if x % 2 == 0}
mapping = {k: v for k, v in zip("abc", [1, 2, 3])}
gen = (x**2 for x in range(10))
`)

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("comprehensions should parse without errors")
	}

	listComps := FindNodesByType(tree.RootNode(), "list_comprehension")
	if len(listComps) != 1 {
		t.Errorf("expected 1 list_comprehension, got %d", len(listComps))
	}
}

// --- Error-tolerant parsing ---

func TestParseJS_BrokenSyntax_NoError(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"missing semicolons", `const x = 1\nconst y = 2`},
		{"unclosed brace", `function foo() { const x = 1;`},
		{"unclosed string", `const msg = "hello`},
		{"missing closing paren", `console.log("hi"`},
		{"extra comma", `const arr = [1, 2, 3,, 5];`},
		{"incomplete expression", `const result = 1 +`},
	}

	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.ParseCtx(t.Context(), nil, []byte(tc.src))
			if err != nil {
				t.Fatalf("parser should not error on broken syntax, got: %v", err)
			}
			root := tree.RootNode()
			if root == nil {
				t.Fatal("root node should not be nil")
			}
			// Tree may have errors but should still produce nodes
			if root.ChildCount() == 0 && root.NamedChildCount() == 0 {
				t.Error("expected some parsed nodes even from broken syntax")
			}
		})
	}
}

func TestParsePython_BrokenSyntax_NoError(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"missing colon", `def foo()\n    pass`},
		{"unclosed string", `msg = "hello`},
		{"bad indentation", "def foo():\npass"},
		{"incomplete expression", `result = 1 +`},
		{"missing closing paren", `print("hi"`},
	}

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.ParseCtx(t.Context(), nil, []byte(tc.src))
			if err != nil {
				t.Fatalf("parser should not error on broken syntax, got: %v", err)
			}
			root := tree.RootNode()
			if root == nil {
				t.Fatal("root node should not be nil")
			}
			if root.ChildCount() == 0 && root.NamedChildCount() == 0 {
				t.Error("expected some parsed nodes even from broken syntax")
			}
		})
	}
}

func TestParseTS_BrokenSyntax_NoError(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"unclosed generic", `function foo<T(arg: T) {}`},
		{"missing type", `const x: = 1;`},
		{"incomplete interface", `interface Foo {`},
	}

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parser.ParseCtx(t.Context(), nil, []byte(tc.src))
			if err != nil {
				t.Fatalf("parser should not error on broken syntax, got: %v", err)
			}
			root := tree.RootNode()
			if root == nil {
				t.Fatal("root node should not be nil")
			}
		})
	}
}

// --- Concurrent parser access (race safety) ---

func TestConcurrentParsing_JS(t *testing.T) {
	snippets := []string{
		`const x = 1;`,
		`function foo() { return "bar"; }`,
		`class Foo { method() {} }`,
		`const fn = async () => await fetch("/api");`,
		`try { x(); } catch (e) { console.error(e); }`,
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			parser := sitter.NewParser()
			parser.SetLanguage(javascript.GetLanguage())
			src := []byte(snippets[idx%len(snippets)])
			tree, err := parser.ParseCtx(t.Context(), nil, src)
			if err != nil {
				t.Errorf("goroutine %d: parse failed: %v", idx, err)
				return
			}
			root := tree.RootNode()
			if root.Type() != "program" {
				t.Errorf("goroutine %d: root = %q, want 'program'", idx, root.Type())
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentParsing_Python(t *testing.T) {
	snippets := []string{
		`x = 1`,
		"def foo():\n    return 'bar'",
		"class Foo:\n    pass",
		"async def fetch():\n    await asyncio.sleep(1)",
		"try:\n    x()\nexcept Exception as e:\n    print(e)",
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			parser := sitter.NewParser()
			parser.SetLanguage(python.GetLanguage())
			src := []byte(snippets[idx%len(snippets)])
			tree, err := parser.ParseCtx(t.Context(), nil, src)
			if err != nil {
				t.Errorf("goroutine %d: parse failed: %v", idx, err)
				return
			}
			root := tree.RootNode()
			if root.Type() != "module" {
				t.Errorf("goroutine %d: root = %q, want 'module'", idx, root.Type())
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentParsing_MixedLanguages(t *testing.T) {
	type langSnippet struct {
		lang *sitter.Language
		src  string
		root string
	}
	snippets := []langSnippet{
		{javascript.GetLanguage(), `const x = 1;`, "program"},
		{typescript.GetLanguage(), `const x: number = 1;`, "program"},
		{tsx.GetLanguage(), `const App = () => <div/>;`, "program"},
		{python.GetLanguage(), `x = 1`, "module"},
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s := snippets[idx%len(snippets)]
			parser := sitter.NewParser()
			parser.SetLanguage(s.lang)
			tree, err := parser.ParseCtx(t.Context(), nil, []byte(s.src))
			if err != nil {
				t.Errorf("goroutine %d: parse failed: %v", idx, err)
				return
			}
			root := tree.RootNode()
			if root.Type() != s.root {
				t.Errorf("goroutine %d: root = %q, want %q", idx, root.Type(), s.root)
			}
		}(i)
	}
	wg.Wait()
}

// --- Empty and large file handling ---

func TestParseJS_EmptyFile(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, []byte(""))
	if err != nil {
		t.Fatalf("empty file parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.Type() != "program" {
		t.Errorf("root = %q, want 'program'", root.Type())
	}
}

func TestParsePython_EmptyFile(t *testing.T) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, []byte(""))
	if err != nil {
		t.Fatalf("empty file parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.Type() != "module" {
		t.Errorf("root = %q, want 'module'", root.Type())
	}
}

func TestParseJS_CommentsOnly(t *testing.T) {
	src := []byte(`// just a comment
/* block comment */
// another comment`)

	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(t.Context(), nil, src)
	if err != nil {
		t.Fatalf("comments-only parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.HasError() {
		t.Error("comments-only file should not have parse errors")
	}
}
