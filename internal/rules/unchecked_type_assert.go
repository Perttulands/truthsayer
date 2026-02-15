package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// UncheckedTypeAssert detects type assertions without the comma-ok form.
// Unchecked type assertions panic at runtime if the type doesn't match.
type UncheckedTypeAssert struct{}

func (u *UncheckedTypeAssert) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.unchecked-type-assert",
		Category:    "silent-fallback",
		Name:        "Unchecked type assertion",
		Description: "Type assertion without comma-ok check — panics on type mismatch",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (u *UncheckedTypeAssert) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	fname := fset.File(file.Pos()).Name()
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	var findings []finding.Finding
	ast.Inspect(file, func(n ast.Node) bool {
		// Type assertions used in assignments: v := x.(T) vs v, ok := x.(T)
		// An unchecked assertion is a TypeAssertExpr NOT used as part of a
		// two-value assignment (comma-ok pattern).
		//
		// We detect this by looking for TypeAssertExpr nodes and checking
		// if their parent is an AssignStmt with 2+ LHS values.
		return true
	})

	// Use a different approach: walk assignments and expression statements
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		findings = append(findings, u.checkBlock(fset, fname, lines, fn.Body)...)
	}
	return findings
}

func (u *UncheckedTypeAssert) checkBlock(fset *token.FileSet, fname string, lines []string, block *ast.BlockStmt) []finding.Finding {
	var findings []finding.Finding

	ast.Inspect(block, func(n ast.Node) bool {
		ta, ok := n.(*ast.TypeAssertExpr)
		if !ok {
			return true
		}
		// Type switch uses TypeAssertExpr with Type == nil: x.(type)
		if ta.Type == nil {
			return true
		}
		// Check if this is inside a comma-ok assignment
		if u.isInCommaOkAssign(block, ta) {
			return true
		}
		pos := fset.Position(ta.Pos())
		findings = append(findings, finding.Finding{
			Rule:       u.Meta().ID,
			Severity:   u.Meta().Severity,
			File:       fname,
			Line:       pos.Line,
			Code:       sourceLine(lines, pos.Line),
			Message:    "Type assertion without comma-ok check — will panic on type mismatch",
			Suggestion: "Use comma-ok form: v, ok := x.(T) and check ok",
		})
		return true
	})

	return findings
}

// isInCommaOkAssign checks if a type assert expr is inside a 2-value assignment.
func (u *UncheckedTypeAssert) isInCommaOkAssign(block *ast.BlockStmt, ta *ast.TypeAssertExpr) bool {
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		if found {
			return false
		}
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		if len(assign.Lhs) >= 2 {
			for _, rhs := range assign.Rhs {
				if containsNode(rhs, ta) {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

// containsNode checks if target is contained within root by position.
func containsNode(root ast.Node, target *ast.TypeAssertExpr) bool {
	contains := false
	ast.Inspect(root, func(n ast.Node) bool {
		if contains {
			return false
		}
		if n == target {
			contains = true
			return false
		}
		return true
	})
	return contains
}
