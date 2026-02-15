package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// ErrorStringCompare detects error comparisons using string matching
// (err.Error() == "...") instead of errors.Is or errors.As.
type ErrorStringCompare struct{}

func (e *ErrorStringCompare) Meta() Rule {
	return Rule{
		ID:          "error-context.error-string-compare",
		Category:    "error-context",
		Name:        "Error string comparison",
		Description: "Error compared by string value instead of errors.Is/errors.As — fragile and breaks with wrapping",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (e *ErrorStringCompare) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	fname := fset.File(file.Pos()).Name()
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	var findings []finding.Finding

	ast.Inspect(file, func(n ast.Node) bool {
		bin, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}
		if bin.Op != token.EQL && bin.Op != token.NEQ {
			return true
		}
		// Check for err.Error() == "string" pattern
		if isErrorMethodCall(bin.X) || isErrorMethodCall(bin.Y) {
			pos := fset.Position(bin.Pos())
			findings = append(findings, finding.Finding{
				Rule:       e.Meta().ID,
				Severity:   e.Meta().Severity,
				File:       fname,
				Line:       pos.Line,
				Code:       sourceLine(lines, pos.Line),
				Message:    "Error compared by string value — fragile and breaks with wrapped errors",
				Suggestion: "Use errors.Is(err, target) or errors.As(err, &target) for reliable error matching",
			})
		}
		return true
	})
	return findings
}

// isErrorMethodCall checks if an expression is a call to .Error() method.
func isErrorMethodCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "Error" && len(call.Args) == 0
}
