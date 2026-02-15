package rules

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/perttulands/truthsayer/internal/finding"
)

// GenericError detects errors.New("failed") and similar generic messages without variable context.
type GenericError struct{}

func (g *GenericError) Meta() Rule {
	return Rule{
		ID:          "error-context.generic-message",
		Category:    "error-context",
		Name:        "Generic error message",
		Description: "Error message lacks specific context (e.g., variable names, operation details)",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

// genericMessages are error messages too vague to be useful.
var genericMessages = map[string]bool{
	"failed":           true,
	"error":            true,
	"something failed": true,
	"an error occurred": true,
	"internal error":   true,
	"unexpected error": true,
	"unknown error":    true,
	"operation failed": true,
	"request failed":   true,
	"invalid input":    true,
	"bad request":      true,
}

func (g *GenericError) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		// Match errors.New("...") or fmt.Errorf("...")
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		isErrorsNew := pkg.Name == "errors" && sel.Sel.Name == "New"
		isFmtErrorf := pkg.Name == "fmt" && sel.Sel.Name == "Errorf"
		if !isErrorsNew && !isFmtErrorf {
			return true
		}
		if len(call.Args) == 0 {
			return true
		}
		// Check first argument for generic string
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		// Unquote string literal properly
		msg, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}
		if genericMessages[msg] {
			pos := fset.Position(call.Pos())
			findings = append(findings, finding.Finding{
				Rule:       g.Meta().ID,
				Severity:   g.Meta().Severity,
				File:       fname,
				Line:       pos.Line,
				Code:       sourceLine(lines, pos.Line),
				Message:    "Error message is too generic — unhelpful for debugging",
				Suggestion: "Include specific context: what operation, what input, what expected outcome",
			})
		}
		return true
	})
	return findings
}
