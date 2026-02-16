package rules

import (
	"go/ast"
	"go/token"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// ScanType indicates whether a rule uses AST or regex-based scanning.
type ScanType int

const (
	ScanTypeAST ScanType = iota
	ScanTypeRegex
)

// Rule describes a detection rule's metadata.
type Rule struct {
	ID          string
	Category    string
	Name        string
	Description string
	Severity    finding.Severity
	FileTypes   []string // e.g. [".go"], [".sh", ".bash"], ["*"]
	ScanType    ScanType
}

// ASTChecker is implemented by rules that analyze Go AST nodes.
// Source lines are provided so rules can include actual code snippets.
type ASTChecker interface {
	Meta() Rule
	CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding
}

// RegexChecker is implemented by rules that match text lines.
type RegexChecker interface {
	Meta() Rule
	CheckLines(path string, lines []string) []finding.Finding
}

// JSASTChecker is implemented by rules that analyze JS/TS AST nodes.
type JSASTChecker interface {
	Meta() Rule
	CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding
}

// PyASTChecker is implemented by rules that analyze Python AST nodes.
type PyASTChecker interface {
	Meta() Rule
	CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding
}
