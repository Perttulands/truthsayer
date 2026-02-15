package rules

import (
	"go/ast"
	"go/constant"
	"go/token"
	"strconv"
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
			if isReturnExitCode(stack, lit) {
				return true
			}
			if isOctalLiteral(lit) {
				return true
			}
			if isSmallComparison(stack, lit) {
				return true
			}
			if isLenArithmetic(stack, lit) {
				return true
			}
			if isCommonCallArg(stack, lit) {
				return true
			}
			if isTimeMultiplier(stack) {
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

// isReturnExitCode returns true for small integer return values (exit codes 0-3).
func isReturnExitCode(stack []ast.Node, lit *ast.BasicLit) bool {
	if lit.Kind != token.INT {
		return false
	}
	v, err := strconv.Atoi(lit.Value)
	if err != nil || v > 3 {
		return false
	}
	for i := len(stack) - 2; i >= 0; i-- {
		if _, ok := stack[i].(*ast.ReturnStmt); ok {
			return true
		}
	}
	return false
}

// isOctalLiteral returns true for octal file permission literals (0o755, 0644, etc.).
func isOctalLiteral(lit *ast.BasicLit) bool {
	if lit.Kind != token.INT {
		return false
	}
	return strings.HasPrefix(lit.Value, "0o") || strings.HasPrefix(lit.Value, "0O") ||
		(len(lit.Value) >= 3 && lit.Value[0] == '0' && lit.Value[1] >= '0' && lit.Value[1] <= '7')
}

// isSmallComparison returns true for small integer literals used in
// comparison expressions (<=, >=, <, >, ==, !=). These are typically
// threshold checks, not arbitrary magic numbers.
func isSmallComparison(stack []ast.Node, lit *ast.BasicLit) bool {
	if lit.Kind != token.INT {
		return false
	}
	v, err := strconv.Atoi(lit.Value)
	if err != nil || v > 128 {
		return false
	}
	for i := len(stack) - 2; i >= 0; i-- {
		bin, ok := stack[i].(*ast.BinaryExpr)
		if !ok {
			continue
		}
		switch bin.Op {
		case token.LEQ, token.GEQ, token.LSS, token.GTR, token.EQL, token.NEQ:
			return true
		}
	}
	return false
}

// isLenArithmetic returns true for small offsets in len()-based arithmetic
// like len(x) - 2 or len(x) + 1.
func isLenArithmetic(stack []ast.Node, lit *ast.BasicLit) bool {
	if lit.Kind != token.INT {
		return false
	}
	v, err := strconv.Atoi(lit.Value)
	if err != nil || v > 4 {
		return false
	}
	for i := len(stack) - 2; i >= 0; i-- {
		bin, ok := stack[i].(*ast.BinaryExpr)
		if !ok {
			continue
		}
		if bin.Op != token.ADD && bin.Op != token.SUB {
			continue
		}
		if containsLenCall(bin.X) || containsLenCall(bin.Y) {
			return true
		}
	}
	return false
}

func containsLenCall(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if ok && ident.Name == "len" {
			found = true
		}
		return !found
	})
	return found
}

// isCommonCallArg returns true for well-known function call arguments
// like SplitN(s, sep, 2), ParseFloat(s, 64), ParseInt(s, 10, 64).
func isCommonCallArg(stack []ast.Node, lit *ast.BasicLit) bool {
	for i := len(stack) - 2; i >= 0; i-- {
		call, ok := stack[i].(*ast.CallExpr)
		if !ok {
			continue
		}
		name := callName(call)
		switch name {
		case "SplitN", "SplitAfterN", "FieldsFunc",
			"ParseFloat", "ParseInt", "ParseUint",
			"FormatFloat", "FormatInt", "FormatUint",
			"NewWriter", "NewReader",
			"Repeat", "Replace",
			"make", "SetIndent":
			return true
		}
		return false
	}
	return false
}

func callName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		return fn.Sel.Name
	case *ast.Ident:
		return fn.Name
	}
	return ""
}

// isTimeMultiplier returns true for time multiplier expressions like 100*time.Millisecond.
func isTimeMultiplier(stack []ast.Node) bool {
	for i := len(stack) - 2; i >= 0; i-- {
		bin, ok := stack[i].(*ast.BinaryExpr)
		if !ok {
			continue
		}
		if bin.Op != token.MUL {
			continue
		}
		if isTimePkgSelector(bin.X) || isTimePkgSelector(bin.Y) {
			return true
		}
	}
	return false
}

func isTimePkgSelector(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "time"
}
