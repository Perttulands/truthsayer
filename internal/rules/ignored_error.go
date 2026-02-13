package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// IgnoredError detects error values explicitly assigned to blank identifier _.
type IgnoredError struct{}

func (ig *IgnoredError) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.ignored-error",
		Category:    "silent-fallback",
		Name:        "Ignored error",
		Description: "Error value explicitly discarded via blank identifier",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (ig *IgnoredError) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		// Look for assignments where the last LHS is _ and RHS is a call
		// Pattern: _, _ = foo() or _ = bar() where result is likely an error
		if len(assign.Lhs) < 1 || len(assign.Rhs) < 1 {
			return true
		}
		// Check if last LHS element is blank identifier
		lastLhs := assign.Lhs[len(assign.Lhs)-1]
		ident, ok := lastLhs.(*ast.Ident)
		if !ok || ident.Name != "_" {
			return true
		}
		// RHS should be a function call (errors come from calls)
		lastRhs := assign.Rhs[len(assign.Rhs)-1]
		if _, isCall := lastRhs.(*ast.CallExpr); !isCall {
			return true
		}
		// Multi-return where last is _ → likely ignoring error
		// Single _ = call() is also suspicious but less certain
		// Only flag multi-value assignments (_, err pattern)
		if len(assign.Lhs) >= 2 {
			pos := fset.Position(assign.Pos())
			findings = append(findings, finding.Finding{
				Rule:       ig.Meta().ID,
				Severity:   ig.Meta().Severity,
				File:       fname,
				Line:       pos.Line,
				Code:       sourceLine(lines, pos.Line),
				Message:    "Error return value explicitly discarded with _",
				Suggestion: "Handle the error: check and return it, or log it if truly ignorable",
			})
		}
		return true
	})
	return findings
}
