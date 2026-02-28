package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// ContextTodo detects context.TODO() and context.Background() used outside of
// main, init, and test functions. These are placeholder contexts that should
// be replaced with real, cancellable contexts in production code.
type ContextTodo struct{}

func (c *ContextTodo) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.context-todo",
		Category:    "bad-defaults",
		Name:        "Placeholder context",
		Description: "context.TODO() or context.Background() used outside main/init — replace with a real context",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (c *ContextTodo) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	fname := fset.File(file.Pos()).Name()
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	var findings []finding.Finding

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		// Allow in main() and init() functions
		name := fn.Name.Name
		if name == "main" || name == "init" {
			continue
		}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
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
			if pkg.Name != "context" {
				return true
			}
			if sel.Sel.Name != "TODO" && sel.Sel.Name != "Background" {
				return true
			}

			pos := fset.Position(call.Pos())
			// Suppress if a REASON comment or nolint directive is on the same
			// line or the line immediately above.
			if hasReasonComment(lines, pos.Line) {
				return true
			}
			what := "context.TODO()"
			if sel.Sel.Name == "Background" {
				what = "context.Background()"
			}
			findings = append(findings, finding.Finding{
				Rule:       c.Meta().ID,
				Severity:   c.Meta().Severity,
				File:       fname,
				Line:       pos.Line,
				Code:       sourceLine(lines, pos.Line),
				Message:    what + " used outside main/init — this placeholder context is not cancellable",
				Suggestion: "Accept context.Context as a parameter and propagate it from the caller",
			})
			return true
		})
	}
	return findings
}

// hasReasonComment checks the line (1-based) and the line above for
// a "// REASON:" or "//nolint:" comment, indicating documented justification.
func hasReasonComment(lines []string, line int) bool {
	for _, i := range []int{line - 1, line - 2} {
		if i < 0 || i >= len(lines) {
			continue
		}
		lower := strings.ToLower(lines[i])
		if strings.Contains(lower, "// reason:") || strings.Contains(lower, "//nolint:") {
			return true
		}
	}
	return false
}
