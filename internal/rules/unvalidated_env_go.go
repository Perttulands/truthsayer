package rules

import (
	"go/ast"
	"go/token"

	"github.com/perttulands/truthsayer/internal/finding"
)

// UnvalidatedEnvGo detects os.Getenv() used directly without validation.
type UnvalidatedEnvGo struct{}

func (u *UnvalidatedEnvGo) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.unvalidated-env-go",
		Category:    "bad-defaults",
		Name:        "Unvalidated env var in Go",
		Description: "os.Getenv used without checking for empty value",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (u *UnvalidatedEnvGo) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if pkg.Name != "os" || sel.Sel.Name != "Getenv" {
			return true
		}
		// Check if os.Getenv is inside an if/assignment that validates
		// We flag all direct uses — wrapping in validation is the fix
		pos := fset.Position(call.Pos())
		findings = append(findings, finding.Finding{
			Rule:       u.Meta().ID,
			Severity:   u.Meta().Severity,
			File:       fname,
			Line:       pos.Line,
			Code:       sourceLine(lines, pos.Line),
			Message:    "os.Getenv used without validation — empty string on missing var",
			Suggestion: "Use os.LookupEnv and check the ok value, or validate non-empty after Getenv",
		})
		return true
	})
	return findings
}
