package scanner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
)

var pyParserPool = sync.Pool{New: func() any {
	p := sitter.NewParser()
	p.SetLanguage(python.GetLanguage())
	return p
}}

// PyScanner scans Python files using tree-sitter AST analysis.
type PyScanner struct {
	checkers []rules.PyASTChecker
}

// NewPyScanner creates a scanner with the given Python AST checkers.
func NewPyScanner(checkers []rules.PyASTChecker) *PyScanner {
	return &PyScanner{checkers: checkers}
}

// Scan parses a Python file and runs all PyASTCheckers against it.
// Returns findings and source lines (for reuse by regex scanner).
func (s *PyScanner) Scan(path string) ([]finding.Finding, []string, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}

	rawParser := pyParserPool.Get()
	parser, ok := rawParser.(*sitter.Parser)
	if !ok || parser == nil {
		parser = sitter.NewParser()
		parser.SetLanguage(python.GetLanguage())
	}
	defer pyParserPool.Put(parser)

	// REASON: scanner API is synchronous and currently has no caller context to propagate.
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", path, err)
	}

	var findings []finding.Finding
	for _, checker := range s.checkers {
		findings = append(findings, checker.CheckPyAST(tree, source, path)...)
	}

	lines := strings.Split(string(source), "\n")
	return findings, lines, nil
}
