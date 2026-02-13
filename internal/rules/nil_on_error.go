package rules

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/perttulands/truthsayer/internal/finding"
)

// NilOnError detects error branches that return a nil error together with empty data.
type NilOnError struct{}

func (n *NilOnError) Meta() Rule {
	return Rule{
		ID:          "error-context.nil-on-error",
		Category:    "error-context",
		Name:        "Nil error returned in error path",
		Description: "err != nil branch returns nil error with empty/nil value",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (n *NilOnError) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	ast.Inspect(file, func(node ast.Node) bool {
		ifStmt, ok := node.(*ast.IfStmt)
		if !ok || !isErrNilCheck(ifStmt.Cond) {
			return true
		}

		ast.Inspect(ifStmt.Body, func(inner ast.Node) bool {
			ret, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}
			if !isNilOrEmptyStringAndNilErrorReturn(ret) {
				return true
			}
			pos := fset.Position(ret.Pos())
			findings = append(findings, finding.Finding{
				Rule:       n.Meta().ID,
				Severity:   n.Meta().Severity,
				File:       fname,
				Line:       pos.Line,
				Code:       sourceLine(lines, pos.Line),
				Message:    "Error branch returns nil error, swallowing the failure",
				Suggestion: "Return err (or wrap with context) instead of returning nil error",
			})
			return true
		})
		return true
	})

	return findings
}

func isNilOrEmptyStringAndNilErrorReturn(ret *ast.ReturnStmt) bool {
	if ret == nil || len(ret.Results) != 2 {
		return false
	}
	if !isNilIdent(ret.Results[1]) {
		return false
	}
	return isNilIdent(ret.Results[0]) || isEmptyStringLiteral(ret.Results[0])
}

func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func isEmptyStringLiteral(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	v, err := strconv.Unquote(lit.Value)
	return err == nil && v == ""
}
