package rules

import (
	"go/ast"
	"go/token"

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
type ASTChecker interface {
	Meta() Rule
	CheckAST(fset *token.FileSet, file *ast.File) []finding.Finding
}

// RegexChecker is implemented by rules that match text lines.
type RegexChecker interface {
	Meta() Rule
	CheckLines(path string, lines []string) []finding.Finding
}
