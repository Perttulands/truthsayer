package rules

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"

	"github.com/perttulands/truthsayer/internal/finding"
)

func runPyCheckerOnSource(t *testing.T, checker PyASTChecker, filename, src string) []finding.Finding {
	t.Helper()

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	source := []byte(src)
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	return checker.CheckPyAST(tree, source, filename)
}
