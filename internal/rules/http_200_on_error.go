package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// HTTP200OnError detects handlers that continue and write a 200-style response after an error branch.
type HTTP200OnError struct{}

func (h *HTTP200OnError) Meta() Rule {
	return Rule{
		ID:          "error-context.http-200-on-error",
		Category:    "error-context",
		Name:        "HTTP 200 after error check",
		Description: "Handler writes success response after err != nil block without returning",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (h *HTTP200OnError) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
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

		isHandler, writerNames, reqNames := extractHTTPHandlerParamNames(fn)
		if !isHandler || (len(writerNames) == 0 && len(reqNames) == 0) {
			continue
		}

		findings = append(findings, h.scanBlock(fset, fname, lines, fn.Body, writerNames)...)
	}

	return findings
}

func (h *HTTP200OnError) scanBlock(
	fset *token.FileSet,
	fname string,
	lines []string,
	block *ast.BlockStmt,
	writerNames map[string]bool,
) []finding.Finding {
	if block == nil {
		return nil
	}

	var findings []finding.Finding
	for i, stmt := range block.List {
		if ifStmt, ok := stmt.(*ast.IfStmt); ok {
			if isErrNilCheck(ifStmt.Cond) && !blockContainsReturn(ifStmt.Body) {
				if successCall := findSuccessWriteCall(block.List[i+1:], writerNames); successCall != nil {
					pos := fset.Position(successCall.Pos())
					findings = append(findings, finding.Finding{
						Rule:       h.Meta().ID,
						Severity:   h.Meta().Severity,
						File:       fname,
						Line:       pos.Line,
						Code:       sourceLine(lines, pos.Line),
						Message:    "Error branch does not return and handler later writes a success response",
						Suggestion: "Return immediately after handling the error before writing a success response",
					})
				}
			}
		}

		for _, nested := range nestedBlocks(stmt) {
			findings = append(findings, h.scanBlock(fset, fname, lines, nested, writerNames)...)
		}
	}
	return findings
}

func blockContainsReturn(block *ast.BlockStmt) bool {
	found := false
	ast.Inspect(block, func(node ast.Node) bool {
		if found {
			return false
		}
		if _, ok := node.(*ast.ReturnStmt); ok {
			found = true
			return false
		}
		return true
	})
	return found
}

func nestedBlocks(stmt ast.Stmt) []*ast.BlockStmt {
	var out []*ast.BlockStmt
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		out = append(out, s)
	case *ast.IfStmt:
		out = append(out, s.Body)
		if s.Else != nil {
			switch e := s.Else.(type) {
			case *ast.BlockStmt:
				out = append(out, e)
			case *ast.IfStmt:
				out = append(out, e.Body)
			}
		}
	case *ast.ForStmt:
		out = append(out, s.Body)
	case *ast.RangeStmt:
		out = append(out, s.Body)
	case *ast.SwitchStmt:
		for _, c := range s.Body.List {
			if clause, ok := c.(*ast.CaseClause); ok {
				out = append(out, &ast.BlockStmt{List: clause.Body})
			}
		}
	case *ast.TypeSwitchStmt:
		for _, c := range s.Body.List {
			if clause, ok := c.(*ast.CaseClause); ok {
				out = append(out, &ast.BlockStmt{List: clause.Body})
			}
		}
	case *ast.SelectStmt:
		for _, c := range s.Body.List {
			if clause, ok := c.(*ast.CommClause); ok {
				out = append(out, &ast.BlockStmt{List: clause.Body})
			}
		}
	}
	return out
}

func findSuccessWriteCall(stmts []ast.Stmt, writerNames map[string]bool) ast.Node {
	for _, stmt := range stmts {
		var found ast.Node
		ast.Inspect(stmt, func(node ast.Node) bool {
			if found != nil {
				return false
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if isSuccessWriteCall(call, writerNames) {
				found = call
				return false
			}
			return true
		})
		if found != nil {
			return found
		}
	}
	return nil
}

func isSuccessWriteCall(call *ast.CallExpr, writerNames map[string]bool) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if wIdent, ok := sel.X.(*ast.Ident); ok && writerNames[wIdent.Name] {
		if sel.Sel.Name == "Write" {
			return true
		}
		if sel.Sel.Name == "WriteHeader" && len(call.Args) > 0 && isStatusOKExpr(call.Args[0]) {
			return true
		}
	}

	if sel.Sel.Name != "Encode" {
		return false
	}
	encCall, ok := sel.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	encSel, ok := encCall.Fun.(*ast.SelectorExpr)
	if !ok || encSel.Sel.Name != "NewEncoder" {
		return false
	}
	jsonPkg, ok := encSel.X.(*ast.Ident)
	if !ok || jsonPkg.Name != "json" {
		return false
	}
	if len(encCall.Args) == 0 {
		return false
	}
	wIdent, ok := encCall.Args[0].(*ast.Ident)
	return ok && writerNames[wIdent.Name]
}

func isStatusOKExpr(expr ast.Expr) bool {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT && lit.Value == "200" {
		return true
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "StatusOK" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "http"
}
