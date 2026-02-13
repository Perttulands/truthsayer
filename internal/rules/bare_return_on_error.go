package rules

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// BareReturnOnError detects named-return functions that silently return zero values on error.
type BareReturnOnError struct{}

func (b *BareReturnOnError) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.bare-return-on-error",
		Category:    "silent-fallback",
		Name:        "Bare return on error",
		Description: "Named-return function returns zero values in err != nil branch",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (b *BareReturnOnError) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || !hasNamedReturns(fn.Type.Results) {
			continue
		}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok || !isErrNilCheck(ifStmt.Cond) {
				return true
			}

			ast.Inspect(ifStmt.Body, func(inner ast.Node) bool {
				ret, ok := inner.(*ast.ReturnStmt)
				if !ok {
					return true
				}
				if len(ret.Results) == 0 || isAllZeroValueReturn(ret.Results) {
					pos := fset.Position(ret.Pos())
					findings = append(findings, finding.Finding{
						Rule:       b.Meta().ID,
						Severity:   b.Meta().Severity,
						File:       fname,
						Line:       pos.Line,
						Code:       sourceLine(lines, pos.Line),
						Message:    "Error branch returns zero values silently",
						Suggestion: "Return the actual error (or wrap it) instead of a bare/zero-value return",
					})
				}
				return true
			})
			return true
		})
	}

	return findings
}

func hasNamedReturns(results *ast.FieldList) bool {
	if results == nil {
		return false
	}
	for _, field := range results.List {
		if len(field.Names) > 0 {
			return true
		}
	}
	return false
}

func isAllZeroValueReturn(results []ast.Expr) bool {
	if len(results) == 0 {
		return false
	}
	for _, expr := range results {
		if !isZeroValueExpr(expr) {
			return false
		}
	}
	return true
}

func isZeroValueExpr(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name == "nil" || v.Name == "false"
	case *ast.BasicLit:
		switch v.Kind {
		case token.STRING:
			s, err := strconv.Unquote(v.Value)
			return err == nil && s == ""
		case token.INT, token.FLOAT, token.IMAG:
			return isNumericZero(v.Value)
		}
	case *ast.UnaryExpr:
		if v.Op == token.ADD || v.Op == token.SUB {
			return isZeroValueExpr(v.X)
		}
	}
	return false
}

func isNumericZero(value string) bool {
	clean := strings.ReplaceAll(strings.ToLower(value), "_", "")
	if strings.ContainsAny(clean, ".ep") {
		f, err := strconv.ParseFloat(clean, 64)
		return err == nil && f == 0
	}
	if strings.HasPrefix(clean, "0x") {
		clean = clean[2:]
	} else if strings.HasPrefix(clean, "0b") {
		clean = clean[2:]
	} else if strings.HasPrefix(clean, "0o") {
		clean = clean[2:]
	}
	if clean == "" {
		return false
	}
	for _, ch := range clean {
		if ch != '0' {
			return false
		}
	}
	return true
}
