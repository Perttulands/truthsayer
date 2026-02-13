package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// UnwrappedError detects bare `return err` without fmt.Errorf wrapping.
type UnwrappedError struct{}

func (u *UnwrappedError) Meta() Rule {
	return Rule{
		ID:          "error-context.unwrapped-error",
		Category:    "error-context",
		Name:        "Unwrapped error return",
		Description: "Error returned without context wrapping via fmt.Errorf",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (u *UnwrappedError) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	ast.Inspect(file, func(n ast.Node) bool {
		// Look for return statements inside if err != nil blocks
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if !isErrNilCheck(ifStmt.Cond) {
			return true
		}
		// Check body for bare return err
		for _, stmt := range ifStmt.Body.List {
			retStmt, ok := stmt.(*ast.ReturnStmt)
			if !ok {
				continue
			}
			for _, result := range retStmt.Results {
				ident, ok := result.(*ast.Ident)
				if !ok || ident.Name != "err" {
					continue
				}
				// Check if the line contains fmt.Errorf (wrapped)
				pos := fset.Position(retStmt.Pos())
				line := sourceLine(lines, pos.Line)
				if strings.Contains(line, "Errorf") || strings.Contains(line, "Wrap") {
					continue
				}
				findings = append(findings, finding.Finding{
					Rule:       u.Meta().ID,
					Severity:   u.Meta().Severity,
					File:       fname,
					Line:       pos.Line,
					Code:       line,
					Message:    "Error returned without context — hard to trace in call chains",
					Suggestion: "Wrap with context: return fmt.Errorf(\"operation: %w\", err)",
				})
			}
		}
		return true
	})
	return findings
}
