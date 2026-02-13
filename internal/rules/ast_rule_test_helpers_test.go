package rules

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/perttulands/truthsayer/internal/finding"
)

func runASTCheckerOnSource(t *testing.T, checker ASTChecker, filename, src string) []finding.Finding {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	return checker.CheckAST(fset, file, strings.Split(src, "\n"))
}
