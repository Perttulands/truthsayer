package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSNoAfterallCleanup detects beforeAll/beforeEach without matching afterAll/afterEach cleanup.
type JSNoAfterallCleanup struct{}

func (j *JSNoAfterallCleanup) Meta() Rule {
	return Rule{
		ID:          "test-isolation.no-afterall-cleanup",
		Category:    "test-isolation",
		Name:        "Missing test cleanup",
		Description: "beforeAll/beforeEach without matching afterAll/afterEach — resources may leak between tests",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSNoAfterallCleanup) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	// Only check test files
	if !jsIsTestFile(path) {
		return nil
	}

	// Scan for beforeAll/beforeEach/afterAll/afterEach at each describe block level
	// and at the top level of the file.
	return j.checkNode(tree.RootNode(), source, path)
}

func (j *JSNoAfterallCleanup) checkNode(node *sitter.Node, source []byte, path string) []finding.Finding {
	var findings []finding.Finding

	// Collect setup/teardown calls at this block level
	type setupCall struct {
		name string
		node *sitter.Node
	}
	var setups []setupCall
	hasAfterAll := false
	hasAfterEach := false

	// Check direct children (not nested describe blocks)
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)

		// Look inside expression_statement for call_expression
		callNode := child
		if child.Type() == "expression_statement" && child.NamedChildCount() > 0 {
			callNode = child.NamedChild(0)
		}

		if callNode.Type() != "call_expression" {
			continue
		}
		fn := callNode.ChildByFieldName("function")
		if fn == nil {
			continue
		}
		fnName := jsNodeText(fn, source)

		switch fnName {
		case "beforeAll", "beforeEach":
			setups = append(setups, setupCall{name: fnName, node: callNode})
		case "afterAll":
			hasAfterAll = true
		case "afterEach":
			hasAfterEach = true
		case "describe", "context":
			// Recurse into describe/context blocks
			args := callNode.ChildByFieldName("arguments")
			if args != nil {
				for k := 0; k < int(args.NamedChildCount()); k++ {
					arg := args.NamedChild(k)
					if arg.Type() == "arrow_function" || arg.Type() == "function_expression" {
						body := arg.ChildByFieldName("body")
						if body != nil {
							findings = append(findings, j.checkNode(body, source, path)...)
						}
					}
				}
			}
		}
	}

	// Report: beforeAll without afterAll, beforeEach without afterEach
	for _, s := range setups {
		var missing string
		switch s.name {
		case "beforeAll":
			if hasAfterAll {
				continue
			}
			missing = "afterAll"
		case "beforeEach":
			if hasAfterEach {
				continue
			}
			missing = "afterEach"
		}
		line := jsLineNumber(s.node)
		findings = append(findings, finding.Finding{
			Rule:       j.Meta().ID,
			Severity:   j.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       jsSourceLine(source, line),
			Message:    s.name + " without matching " + missing + " — resources may leak between tests",
			Suggestion: "Add " + missing + " to clean up resources created in " + s.name,
		})
	}

	return findings
}
