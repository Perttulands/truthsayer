package scanner

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"os"

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
		return nil, fmt.Errorf("parse go file %s: %w", path, err)
	}

	lines, err := readGoLines(path)
	if err != nil {
		return nil, fmt.Errorf("read go source lines %s: %w", path, err)
	}

	var findings []finding.Finding
	for _, checker := range s.checkers {
		findings = append(findings, checker.CheckAST(fset, file, lines)...)
	}
	return findings, nil
}

func readGoLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan lines in %s: %w", path, err)
	}
	return lines, nil
}
