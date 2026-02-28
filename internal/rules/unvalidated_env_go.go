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

	// Pass 1: collect variable names compared to "" in if-conditions.
	// These are considered validated even if the check is on a later line.
	validatedVars := collectValidatedVars(file)

	ast.Inspect(file, func(n ast.Node) bool {
		// Skip if os.Getenv is used inside an if-init statement,
		// which means the value is being checked inline.
		if ifStmt, ok := n.(*ast.IfStmt); ok {
			if ifStmt.Init != nil && containsGetenvCall(ifStmt.Init) {
				return false // skip — the if-init validates the value
			}
			return true
		}

		// Skip if os.Getenv result is assigned to a variable that gets
		// validated (compared to "") elsewhere in the same function.
		if assign, ok := n.(*ast.AssignStmt); ok {
			if containsGetenvCall(assign) {
				for _, lhs := range assign.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						if validatedVars[ident.Name] {
							return false
						}
					}
				}
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

// collectValidatedVars finds all variable names compared to "" in if-conditions.
func collectValidatedVars(file *ast.File) map[string]bool {
	vars := make(map[string]bool)
	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		collectEmptyStringChecks(ifStmt.Cond, vars)
		return true
	})
	return vars
}

// collectEmptyStringChecks extracts variable names from comparisons to "".
// Handles both `varName != ""` and `varName == ""`.
func collectEmptyStringChecks(expr ast.Expr, vars map[string]bool) {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return
	}
	// Handle && and || by recursing into both sides
	if bin.Op == token.LAND || bin.Op == token.LOR {
		collectEmptyStringChecks(bin.X, vars)
		collectEmptyStringChecks(bin.Y, vars)
		return
	}
	if bin.Op != token.NEQ && bin.Op != token.EQL {
		return
	}
	// Check both orderings: varName != "" and "" != varName
	if ident, lit := asIdentAndStringLit(bin.X, bin.Y); ident != nil && lit == `""` {
		vars[ident.Name] = true
	}
	if ident, lit := asIdentAndStringLit(bin.Y, bin.X); ident != nil && lit == `""` {
		vars[ident.Name] = true
	}
}

func asIdentAndStringLit(a, b ast.Expr) (*ast.Ident, string) {
	ident, ok := a.(*ast.Ident)
	if !ok {
		return nil, ""
	}
	lit, ok := b.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return nil, ""
	}
	return ident, lit.Value
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
