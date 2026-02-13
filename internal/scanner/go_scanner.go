package scanner

import (
	"go/parser"
	"go/token"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
)

// GoScanner scans Go files using AST analysis.
type GoScanner struct {
	checkers []rules.ASTChecker
}

// NewGoScanner creates a scanner with the given AST checkers.
func NewGoScanner(checkers []rules.ASTChecker) *GoScanner {
	return &GoScanner{checkers: checkers}
}

// Scan parses a Go file and runs all AST checkers against it.
func (s *GoScanner) Scan(path string) ([]finding.Finding, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var findings []finding.Finding
	for _, checker := range s.checkers {
		findings = append(findings, checker.CheckAST(fset, file)...)
	}
	return findings, nil
}
