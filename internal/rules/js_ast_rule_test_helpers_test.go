package rules

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	"github.com/perttulands/truthsayer/internal/finding"
)

func runJSCheckerOnSource(t *testing.T, checker JSASTChecker, filename, src string) []finding.Finding {
	t.Helper()

	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	source := []byte(src)
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	return checker.CheckJSAST(tree, source, filename)
}

func runTSCheckerOnSource(t *testing.T, checker JSASTChecker, filename, src string) []finding.Finding {
	t.Helper()

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	source := []byte(src)
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	return checker.CheckJSAST(tree, source, filename)
}
