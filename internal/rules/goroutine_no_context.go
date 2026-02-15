package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// GoroutineNoContext detects goroutines launched without context or done channel.
// These goroutines cannot be cancelled or timed out, leading to goroutine leaks.
type GoroutineNoContext struct{}

func (g *GoroutineNoContext) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.goroutine-no-context",
		Category:    "bad-defaults",
		Name:        "Goroutine without context",
		Description: "Goroutine launched without context.Context or done channel — may leak",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (g *GoroutineNoContext) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	fname := fset.File(file.Pos()).Name()
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	var findings []finding.Finding
	ast.Inspect(file, func(n ast.Node) bool {
		goStmt, ok := n.(*ast.GoStmt)
		if !ok {
			return true
		}

		// Get the function body of the goroutine
		var body *ast.BlockStmt
		switch fn := goStmt.Call.Fun.(type) {
		case *ast.FuncLit:
			body = fn.Body
		default:
			// go someFunc(args...) — check if any arg is a context
			for _, arg := range goStmt.Call.Args {
				if referencesContext(arg) {
					return true
				}
			}
			// Named function call without context arg
			pos := fset.Position(goStmt.Pos())
			findings = append(findings, finding.Finding{
				Rule:       g.Meta().ID,
				Severity:   g.Meta().Severity,
				File:       fname,
				Line:       pos.Line,
				Code:       sourceLine(lines, pos.Line),
				Message:    "Goroutine launched without context — no way to cancel or timeout",
				Suggestion: "Pass a context.Context or done channel to enable cancellation",
			})
			return true
		}

		if body == nil {
			return true
		}

		// For func literals: check if captured/param vars reference context or done channel
		funcLit := goStmt.Call.Fun.(*ast.FuncLit)

		// Check params for context.Context
		if hasContextParam(funcLit.Type) {
			return true
		}

		// Check if body references ctx, context, or done channel patterns
		if bodyReferencesContextOrDone(body) {
			return true
		}

		pos := fset.Position(goStmt.Pos())
		findings = append(findings, finding.Finding{
			Rule:       g.Meta().ID,
			Severity:   g.Meta().Severity,
			File:       fname,
			Line:       pos.Line,
			Code:       sourceLine(lines, pos.Line),
			Message:    "Goroutine launched without context or done channel — may leak",
			Suggestion: "Pass a context.Context or done channel to enable cancellation",
		})
		return true
	})

	return findings
}

// referencesContext checks if an expression likely references a context variable.
func referencesContext(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		name := strings.ToLower(ident.Name)
		if name == "ctx" || name == "context" || strings.HasSuffix(name, "ctx") {
			found = true
			return false
		}
		return true
	})
	return found
}

// hasContextParam checks if a function type has a context.Context parameter.
func hasContextParam(ft *ast.FuncType) bool {
	if ft == nil || ft.Params == nil {
		return false
	}
	for _, field := range ft.Params.List {
		if isContextType(field.Type) {
			return true
		}
	}
	return false
}

// isContextType checks if a type expression is context.Context.
func isContextType(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return pkg.Name == "context" && sel.Sel.Name == "Context"
}

// bodyReferencesContextOrDone checks if a block references context or done patterns.
func bodyReferencesContextOrDone(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		switch node := n.(type) {
		case *ast.Ident:
			name := strings.ToLower(node.Name)
			if name == "ctx" || name == "context" || strings.HasSuffix(name, "ctx") || name == "done" {
				found = true
				return false
			}
		case *ast.SelectorExpr:
			// Check for ctx.Done(), ctx.Err()
			if ident, ok := node.X.(*ast.Ident); ok {
				name := strings.ToLower(ident.Name)
				if (name == "ctx" || strings.HasSuffix(name, "ctx")) &&
					(node.Sel.Name == "Done" || node.Sel.Name == "Err" || node.Sel.Name == "Deadline") {
					found = true
					return false
				}
			}
		case *ast.ChanType:
			// done channel pattern
			found = true
			return false
		}
		return true
	})
	return found
}
