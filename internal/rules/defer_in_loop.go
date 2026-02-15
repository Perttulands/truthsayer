package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// DeferInLoop detects defer statements inside for/range loops.
// Deferred calls don't run until the function returns, not at the end of
// each loop iteration, causing resource leaks (e.g., unclosed files/connections).
type DeferInLoop struct{}

func (d *DeferInLoop) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.defer-in-loop",
		Category:    "bad-defaults",
		Name:        "Defer in loop",
		Description: "Defer inside loop body — deferred call runs at function exit, not end of iteration",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

func (d *DeferInLoop) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	fname := fset.File(file.Pos()).Name()
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	var findings []finding.Finding
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		findings = append(findings, d.checkForDefers(fset, fname, lines, fn.Body, 0)...)
	}
	return findings
}

// checkForDefers walks a block, tracking loop nesting depth.
func (d *DeferInLoop) checkForDefers(fset *token.FileSet, fname string, lines []string, block *ast.BlockStmt, loopDepth int) []finding.Finding {
	var findings []finding.Finding

	for _, stmt := range block.List {
		switch s := stmt.(type) {
		case *ast.DeferStmt:
			if loopDepth > 0 {
				pos := fset.Position(s.Pos())
				findings = append(findings, finding.Finding{
					Rule:       d.Meta().ID,
					Severity:   d.Meta().Severity,
					File:       fname,
					Line:       pos.Line,
					Code:       sourceLine(lines, pos.Line),
					Message:    "Defer inside loop — deferred calls accumulate until function returns",
					Suggestion: "Move the deferred call to a helper function, or call the cleanup directly at end of iteration",
				})
			}
		case *ast.ForStmt:
			if s.Body != nil {
				findings = append(findings, d.checkForDefers(fset, fname, lines, s.Body, loopDepth+1)...)
			}
		case *ast.RangeStmt:
			if s.Body != nil {
				findings = append(findings, d.checkForDefers(fset, fname, lines, s.Body, loopDepth+1)...)
			}
		case *ast.IfStmt:
			findings = append(findings, d.checkForDefers(fset, fname, lines, s.Body, loopDepth)...)
			if s.Else != nil {
				switch e := s.Else.(type) {
				case *ast.BlockStmt:
					findings = append(findings, d.checkForDefers(fset, fname, lines, e, loopDepth)...)
				case *ast.IfStmt:
					if e.Body != nil {
						findings = append(findings, d.checkForDefers(fset, fname, lines, e.Body, loopDepth)...)
					}
				}
			}
		case *ast.SwitchStmt:
			if s.Body != nil {
				for _, c := range s.Body.List {
					if clause, ok := c.(*ast.CaseClause); ok {
						findings = append(findings, d.checkForDefers(fset, fname, lines, &ast.BlockStmt{List: clause.Body}, loopDepth)...)
					}
				}
			}
		case *ast.TypeSwitchStmt:
			if s.Body != nil {
				for _, c := range s.Body.List {
					if clause, ok := c.(*ast.CaseClause); ok {
						findings = append(findings, d.checkForDefers(fset, fname, lines, &ast.BlockStmt{List: clause.Body}, loopDepth)...)
					}
				}
			}
		case *ast.SelectStmt:
			if s.Body != nil {
				for _, c := range s.Body.List {
					if clause, ok := c.(*ast.CommClause); ok {
						findings = append(findings, d.checkForDefers(fset, fname, lines, &ast.BlockStmt{List: clause.Body}, loopDepth)...)
					}
				}
			}
		case *ast.BlockStmt:
			findings = append(findings, d.checkForDefers(fset, fname, lines, s, loopDepth)...)
		}
	}
	return findings
}
