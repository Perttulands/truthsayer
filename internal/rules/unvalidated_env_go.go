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
		// Skip if os.Getenv is used inside an if-init statement,
		// which means the value is being checked inline.
		if ifStmt, ok := n.(*ast.IfStmt); ok {
			if ifStmt.Init != nil && containsGetenvCall(ifStmt.Init) {
				return false // skip — the if-init validates the value
			}
			return true
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isGetenvCall(call) {
			return true
		}

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

func isGetenvCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return pkg.Name == "os" && sel.Sel.Name == "Getenv"
}

func containsGetenvCall(node ast.Node) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if ok && isGetenvCall(call) {
			found = true
			return false
		}
		return true
	})
	return found
}
