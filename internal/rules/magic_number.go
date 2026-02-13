package rules

import (
	"go/ast"
	"go/constant"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// MagicNumber detects direct numeric literals > 1 in function bodies.
type MagicNumber struct{}

func (m *MagicNumber) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.magic-number",
		Category:    "bad-defaults",
		Name:        "Magic number",
		Description: "Numeric literal greater than 1 used directly in function body",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (m *MagicNumber) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	fname := fset.File(file.Pos()).Name()
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	var findings []finding.Finding
	flaggedLines := make(map[int]bool)

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		var stack []ast.Node
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if n == nil {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
				return true
			}

			stack = append(stack, n)
			lit, ok := n.(*ast.BasicLit)
			if !ok {
				return true
			}
			if lit.Kind != token.INT && lit.Kind != token.FLOAT {
				return true
			}
			if !isGreaterThanOne(lit) {
				return true
			}
			if isNegativeLiteral(stack, lit) {
				return true
			}
			if isConstLiteral(stack) {
				return true
			}
			if isIndexLiteral(stack, lit) {
				return true
			}

			pos := fset.Position(lit.Pos())
			if flaggedLines[pos.Line] {
				return true
			}
			flaggedLines[pos.Line] = true

			findings = append(findings, finding.Finding{
				Rule:       m.Meta().ID,
				Severity:   m.Meta().Severity,
				File:       fname,
				Line:       pos.Line,
				Code:       sourceLine(lines, pos.Line),
				Message:    "Magic number used directly in function body",
				Suggestion: "Extract numeric values into a named const for readability and safer defaults",
			})
			return true
		})
	}

	return findings
}

func isGreaterThanOne(lit *ast.BasicLit) bool {
	v := constant.MakeFromLiteral(lit.Value, lit.Kind, 0)
	if v.Kind() == constant.Unknown {
		return false
	}
	return constant.Compare(v, token.GTR, constant.MakeInt64(1))
}

func isNegativeLiteral(stack []ast.Node, lit *ast.BasicLit) bool {
	if len(stack) < 2 {
		return false
	}
	parent, ok := stack[len(stack)-2].(*ast.UnaryExpr)
	if !ok {
		return false
	}
	return parent.Op == token.SUB && parent.X == lit
}

func isConstLiteral(stack []ast.Node) bool {
	for i := len(stack) - 2; i >= 0; i-- {
		gen, ok := stack[i].(*ast.GenDecl)
		if ok && gen.Tok == token.CONST {
			return true
		}
	}
	return false
}

func isIndexLiteral(stack []ast.Node, lit *ast.BasicLit) bool {
	for i := len(stack) - 2; i >= 0; i-- {
		switch idx := stack[i].(type) {
		case *ast.IndexExpr:
			if idx.Index != nil && lit.Pos() >= idx.Index.Pos() && lit.End() <= idx.Index.End() {
				return true
			}
		case *ast.SliceExpr:
			if idx.Low != nil && lit.Pos() >= idx.Low.Pos() && lit.End() <= idx.Low.End() {
				return true
			}
			if idx.High != nil && lit.Pos() >= idx.High.Pos() && lit.End() <= idx.High.End() {
				return true
			}
			if idx.Max != nil && lit.Pos() >= idx.Max.Pos() && lit.End() <= idx.Max.End() {
				return true
			}
		}
	}
	return false
}
