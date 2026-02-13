package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// DebugGuard detects debug/testing guards in production files.
type DebugGuard struct{}

func (d *DebugGuard) Meta() Rule {
	return Rule{
		ID:          "mock-leakage.debug-guard",
		Category:    "mock-leakage",
		Name:        "Debug/testing guard in production code",
		Description: "Debug or test-only conditional guard in non-test Go file",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (d *DebugGuard) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
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
		if !hasDebugGuardIdent(ifStmt.Cond) {
			return true
		}

		pos := fset.Position(ifStmt.Pos())
		findings = append(findings, finding.Finding{
			Rule:       d.Meta().ID,
			Severity:   d.Meta().Severity,
			File:       fname,
			Line:       pos.Line,
			Code:       sourceLine(lines, pos.Line),
			Message:    "Debug/test guard detected in production code",
			Suggestion: "Remove test-only guard logic from production paths or gate it via explicit configuration",
		})
		return true
	})

	return findings
}

func hasDebugGuardIdent(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		switch strings.ToLower(ident.Name) {
		case "debug", "testing", "istest", "istesting":
			found = true
			return false
		default:
			return true
		}
	})
	return found
}
