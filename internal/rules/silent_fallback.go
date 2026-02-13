package rules

import (
	"go/ast"
	"go/token"

	"github.com/perttulands/truthsayer/internal/finding"
)

// EmptyErrorCheck detects `if err != nil { return nil }` without logging.
type EmptyErrorCheck struct{}

func (e *EmptyErrorCheck) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.empty-error-check",
		Category:    "silent-fallback",
		Name:        "Empty error check",
		Description: "Error checked but returned as nil without logging or wrapping",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (e *EmptyErrorCheck) CheckAST(fset *token.FileSet, file *ast.File) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if !isErrNilCheck(ifStmt.Cond) {
			return true
		}
		// Check if body only contains a return with nil values
		if isNilReturn(ifStmt.Body) {
			pos := fset.Position(ifStmt.Pos())
			findings = append(findings, finding.Finding{
				Rule:       e.Meta().ID,
				Severity:   e.Meta().Severity,
				File:       fname,
				Line:       pos.Line,
				Code:       "if err != nil { return nil }",
				Message:    "Error returned as nil without logging or wrapping",
				Suggestion: "Return the error or log it: return fmt.Errorf(\"operation failed: %w\", err)",
			})
		}
		return true
	})
	return findings
}

// isErrNilCheck checks if the condition is `err != nil`.
func isErrNilCheck(expr ast.Expr) bool {
	binExpr, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	ident, ok := binExpr.X.(*ast.Ident)
	if !ok {
		return false
	}
	if ident.Name != "err" {
		return false
	}
	_, isNil := binExpr.Y.(*ast.Ident)
	return isNil && binExpr.Y.(*ast.Ident).Name == "nil" && binExpr.Op.String() == "!="
}

// isNilReturn checks if a block contains only a return statement with all nil values.
func isNilReturn(block *ast.BlockStmt) bool {
	if len(block.List) != 1 {
		return false
	}
	retStmt, ok := block.List[0].(*ast.ReturnStmt)
	if !ok {
		return false
	}
	if len(retStmt.Results) == 0 {
		return false
	}
	for _, r := range retStmt.Results {
		ident, ok := r.(*ast.Ident)
		if !ok || ident.Name != "nil" {
			return false
		}
	}
	return true
}
