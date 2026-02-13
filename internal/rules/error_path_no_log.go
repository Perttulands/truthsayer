package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// ErrorPathNoLog detects error handling branches (if err != nil) without any logging.
type ErrorPathNoLog struct{}

func (e *ErrorPathNoLog) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.error-path-no-log",
		Category:    "trace-gaps",
		Name:        "Error path without logging",
		Description: "Error handling block without logging before return",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (e *ErrorPathNoLog) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if !isErrNilCheck(ifStmt.Cond) {
			return true
		}
		// Check if body has any log call or explicit stderr write.
		if hasLogCall(ifStmt.Body) || hasStderrWriteCall(ifStmt.Body) {
			return true
		}
		// Check body length — single-line returns are often fine
		if len(ifStmt.Body.List) <= 1 {
			return true
		}
		pos := fset.Position(ifStmt.Pos())
		findings = append(findings, finding.Finding{
			Rule:       e.Meta().ID,
			Severity:   e.Meta().Severity,
			File:       fname,
			Line:       pos.Line,
			Code:       sourceLine(lines, pos.Line),
			Message:    "Error handling block without logging — errors may be lost silently",
			Suggestion: "Add logging before return: log.Error(\"operation failed\", \"err\", err)",
		})
		return true
	})
	return findings
}

func hasStderrWriteCall(block *ast.BlockStmt) bool {
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "fmt" {
			return true
		}
		if sel.Sel.Name != "Fprint" && sel.Sel.Name != "Fprintf" && sel.Sel.Name != "Fprintln" {
			return true
		}
		if len(call.Args) == 0 {
			return true
		}
		if isOSStderr(call.Args[0]) {
			found = true
			return false
		}
		return true
	})
	return found
}

func isOSStderr(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return pkg.Name == "os" && sel.Sel.Name == "Stderr"
}
