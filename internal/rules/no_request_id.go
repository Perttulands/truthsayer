package rules

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// NoRequestID detects HTTP handlers that don't extract a request ID from headers/context.
type NoRequestID struct{}

func (n *NoRequestID) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.no-request-id",
		Category:    "trace-gaps",
		Name:        "No request ID extraction",
		Description: "HTTP handler does not extract request ID from headers or context",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (n *NoRequestID) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		isHandler, _, reqNames := extractHTTPHandlerParamNames(fn)
		if !isHandler {
			continue
		}
		if handlerHasRequestIDExtraction(fn.Body, reqNames) {
			continue
		}

		pos := fset.Position(fn.Pos())
		findings = append(findings, finding.Finding{
			Rule:       n.Meta().ID,
			Severity:   n.Meta().Severity,
			File:       fname,
			Line:       pos.Line,
			Code:       sourceLine(lines, pos.Line),
			Message:    "HTTP handler does not extract request ID from header or context",
			Suggestion: "Extract request ID (e.g. r.Header.Get(\"X-Request-ID\") or r.Context().Value(...)) for traceability",
		})
	}

	return findings
}

func extractHTTPHandlerParamNames(fn *ast.FuncDecl) (bool, map[string]bool, map[string]bool) {
	if fn == nil || fn.Type == nil || fn.Type.Params == nil {
		return false, nil, nil
	}

	type paramInfo struct {
		name     string
		typeExpr ast.Expr
	}

	var params []paramInfo
	for _, field := range fn.Type.Params.List {
		if len(field.Names) == 0 {
			params = append(params, paramInfo{name: "", typeExpr: field.Type})
			continue
		}
		for _, name := range field.Names {
			params = append(params, paramInfo{name: name.Name, typeExpr: field.Type})
		}
	}

	if len(params) < 2 {
		return false, nil, nil
	}
	if !isResponseWriterType(params[0].typeExpr) || !isRequestPtrType(params[1].typeExpr) {
		return false, nil, nil
	}

	writerNames := map[string]bool{}
	reqNames := map[string]bool{}
	if params[0].name != "" {
		writerNames[params[0].name] = true
	}
	if params[1].name != "" {
		reqNames[params[1].name] = true
	}
	return true, writerNames, reqNames
}

func isResponseWriterType(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "ResponseWriter" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "http"
}

func isRequestPtrType(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Request" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "http"
}

func handlerHasRequestIDExtraction(body *ast.BlockStmt, reqNames map[string]bool) bool {
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		if found {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isRequestHeaderGetCall(call, reqNames) || isRequestContextValueCall(call, reqNames) || isRequestIDHelperCall(call, reqNames) {
			found = true
			return false
		}
		return true
	})
	return found
}

func isRequestHeaderGetCall(call *ast.CallExpr, reqNames map[string]bool) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Get" {
		return false
	}
	headerSel, ok := sel.X.(*ast.SelectorExpr)
	if !ok || headerSel.Sel.Name != "Header" {
		return false
	}
	reqIdent, ok := headerSel.X.(*ast.Ident)
	if !ok || !reqNames[reqIdent.Name] {
		return false
	}
	if len(call.Args) == 0 {
		return false
	}
	return isRequestIDLiteral(call.Args[0])
}

func isRequestContextValueCall(call *ast.CallExpr, reqNames map[string]bool) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Value" {
		return false
	}
	return isRequestContextExpr(sel.X, reqNames)
}

func isRequestIDHelperCall(call *ast.CallExpr, reqNames map[string]bool) bool {
	name := callFunctionName(call.Fun)
	normalized := strings.ToLower(name)
	if !strings.Contains(normalized, "reqid") && !strings.Contains(normalized, "requestid") {
		return false
	}
	for _, arg := range call.Args {
		if isRequestContextExpr(arg, reqNames) {
			return true
		}
	}
	return false
}

func callFunctionName(expr ast.Expr) string {
	switch fn := expr.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		if fn.Sel != nil {
			return fn.Sel.Name
		}
	}
	return ""
}

func isRequestContextExpr(expr ast.Expr, reqNames map[string]bool) bool {
	ctxCall, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := ctxCall.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Context" {
		return false
	}
	reqIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return reqNames[reqIdent.Name]
}

func isRequestIDLiteral(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	val, err := strconv.Unquote(lit.Value)
	if err != nil {
		return false
	}
	normalized := strings.ToLower(val)
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	return strings.Contains(normalized, "requestid")
}
