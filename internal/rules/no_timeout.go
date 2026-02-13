package rules

import (
	"go/ast"
	"go/token"

	"github.com/perttulands/truthsayer/internal/finding"
)

// NoTimeout detects http.Client{} or http.Get/Post without explicit timeout.
type NoTimeout struct{}

func (nt *NoTimeout) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.no-timeout",
		Category:    "bad-defaults",
		Name:        "No timeout on HTTP client",
		Description: "HTTP client or request without explicit timeout — can hang indefinitely",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (nt *NoTimeout) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	ast.Inspect(file, func(n ast.Node) bool {
		// Detect http.Client{} composite literal without Timeout field
		compLit, ok := n.(*ast.CompositeLit)
		if ok {
			sel, ok := compLit.Type.(*ast.SelectorExpr)
			if ok {
				pkg, ok := sel.X.(*ast.Ident)
				if ok && pkg.Name == "http" && sel.Sel.Name == "Client" {
					hasTimeout := false
					for _, elt := range compLit.Elts {
						kv, ok := elt.(*ast.KeyValueExpr)
						if !ok {
							continue
						}
						key, ok := kv.Key.(*ast.Ident)
						if ok && key.Name == "Timeout" {
							hasTimeout = true
							break
						}
					}
					if !hasTimeout {
						pos := fset.Position(compLit.Pos())
						findings = append(findings, finding.Finding{
							Rule:       nt.Meta().ID,
							Severity:   nt.Meta().Severity,
							File:       fname,
							Line:       pos.Line,
							Code:       sourceLine(lines, pos.Line),
							Message:    "http.Client without Timeout — requests can hang indefinitely",
							Suggestion: "Set explicit timeout: &http.Client{Timeout: 30 * time.Second}",
						})
					}
				}
			}
		}

		// Detect http.Get(), http.Post() etc. (uses default client with no timeout)
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
		if pkg.Name == "http" {
			switch sel.Sel.Name {
			case "Get", "Post", "PostForm", "Head":
				pos := fset.Position(call.Pos())
				findings = append(findings, finding.Finding{
					Rule:       nt.Meta().ID,
					Severity:   nt.Meta().Severity,
					File:       fname,
					Line:       pos.Line,
					Code:       sourceLine(lines, pos.Line),
					Message:    "http." + sel.Sel.Name + "() uses DefaultClient with no timeout",
					Suggestion: "Create a client with explicit timeout instead of using package-level functions",
				})
			}
		}
		return true
	})
	return findings
}
