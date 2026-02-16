package rules

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSNoErrorHandlerExpress detects Express apps that define routes but lack
// error-handling middleware (err, req, res, next).
type JSNoErrorHandlerExpress struct{}

func (j *JSNoErrorHandlerExpress) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.js-no-error-handler-express",
		Category:    "trace-gaps",
		Name:        "Express app missing error handler",
		Description: "Express app defines routes but has no error-handling middleware (err, req, res, next)",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSNoErrorHandlerExpress) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	root := tree.RootNode()
	info := jsFindExpressApp(root, source)
	if info == nil || !info.hasRoute {
		return nil
	}

	// Scan all app.use() calls for error middleware (4-param function with err first)
	for _, call := range jsFindNodesByType(root, "call_expression") {
		fn := call.ChildByFieldName("function")
		if fn == nil || fn.Type() != "member_expression" {
			continue
		}
		obj := fn.ChildByFieldName("object")
		prop := fn.ChildByFieldName("property")
		if obj == nil || prop == nil {
			continue
		}
		if !info.appNames[jsNodeText(obj, source)] || jsNodeText(prop, source) != "use" {
			continue
		}
		args := call.ChildByFieldName("arguments")
		if args == nil {
			continue
		}
		for i := 0; i < int(args.NamedChildCount()); i++ {
			if isErrorMiddleware(args.NamedChild(i), source) {
				return nil
			}
		}
	}

	line := jsLineNumber(info.firstRouteNode)
	return []finding.Finding{{
		Rule:       j.Meta().ID,
		Severity:   j.Meta().Severity,
		File:       path,
		Line:       line,
		Code:       jsSourceLine(source, line),
		Message:    "Express app defines routes but has no error-handling middleware",
		Suggestion: "Add error middleware: app.use((err, req, res, next) => { ... })",
	}}
}

// isErrorMiddleware checks if a node is a function with 4 parameters where
// the first is named "err" or "error" — the Express error middleware signature.
func isErrorMiddleware(node *sitter.Node, source []byte) bool {
	if node == nil {
		return false
	}

	nodeType := node.Type()
	if nodeType != "arrow_function" && nodeType != "function_expression" && nodeType != "function_declaration" {
		return false
	}

	params := node.ChildByFieldName("parameters")
	if params == nil || params.NamedChildCount() != 4 {
		return false
	}

	firstParam := params.NamedChild(0)
	if firstParam == nil {
		return false
	}

	name := jsNodeText(firstParam, source)
	name = strings.TrimPrefix(name, "...")
	return name == "err" || name == "error"
}
