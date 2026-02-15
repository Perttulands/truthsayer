package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// SwallowedError detects error handling blocks that log the error but don't
// return or propagate it, causing execution to continue silently.
type SwallowedError struct{}

func (s *SwallowedError) Meta() Rule {
	return Rule{
		ID:          "error-context.swallowed-error",
		Category:    "error-context",
		Name:        "Swallowed error",
		Description: "Error is logged but not returned — execution continues past the error",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (s *SwallowedError) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	fname := fset.File(file.Pos()).Name()
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	var findings []finding.Finding

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if !isErrNilCheck(ifStmt.Cond) {
			return true
		}
		// Check if the body logs but does not return or break flow
		if hasLogCall(ifStmt.Body) && !hasFlowControl(ifStmt.Body) {
			pos := fset.Position(ifStmt.Pos())
			findings = append(findings, finding.Finding{
				Rule:       s.Meta().ID,
				Severity:   s.Meta().Severity,
				File:       fname,
				Line:       pos.Line,
				Code:       sourceLine(lines, pos.Line),
				Message:    "Error is logged but not returned — execution continues past the error",
				Suggestion: "Return the error after logging, or use log.Fatal if the error is truly unrecoverable",
			})
		}
		return true
	})
	return findings
}

// hasFlowControl checks if a block contains a return, continue, break,
// or goto statement that would prevent silent continuation.
func hasFlowControl(block *ast.BlockStmt) bool {
	for _, stmt := range block.List {
		switch stmt.(type) {
		case *ast.ReturnStmt, *ast.BranchStmt:
			return true
		}
	}
	return false
}
