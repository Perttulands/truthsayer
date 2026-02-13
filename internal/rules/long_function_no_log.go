package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// LongFunctionNoLog detects functions >20 lines with no log/trace/print call.
type LongFunctionNoLog struct{}

func (l *LongFunctionNoLog) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.long-function-no-log",
		Category:    "trace-gaps",
		Name:        "Long function without logging",
		Description: "Function >20 lines with no logging or tracing calls",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

var logCallNames = map[string]bool{
	"Print": true, "Println": true, "Printf": true,
	"Log": true, "Logf": true,
	"Info": true, "Infof": true, "Infow": true,
	"Warn": true, "Warnf": true, "Warnw": true,
	"Error": true, "Errorf": true, "Errorw": true,
	"Debug": true, "Debugf": true, "Debugw": true,
	"Fatal": true, "Fatalf": true, "Fatalw": true,
	"Trace": true, "Tracef": true,
}

func (l *LongFunctionNoLog) CheckAST(fset *token.FileSet, file *ast.File, lines []string) []finding.Finding {
	var findings []finding.Finding
	fname := fset.File(file.Pos()).Name()

	// Skip test files
	if strings.HasSuffix(fname, "_test.go") {
		return nil
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		startLine := fset.Position(fn.Body.Pos()).Line
		endLine := fset.Position(fn.Body.End()).Line
		bodyLines := endLine - startLine
		if bodyLines <= 20 {
			continue
		}
		if hasLogCall(fn.Body) {
			continue
		}
		pos := fset.Position(fn.Pos())
		findings = append(findings, finding.Finding{
			Rule:       l.Meta().ID,
			Severity:   l.Meta().Severity,
			File:       fname,
			Line:       pos.Line,
			Code:       sourceLine(lines, pos.Line),
			Message:    "Function " + fn.Name.Name + " is " + strings.Replace(string(rune(bodyLines+'0')), "\x00", "", -1) + " lines with no logging",
			Suggestion: "Add structured logging for key operations and error paths",
		})
	}
	return findings
}

func hasLogCall(block *ast.BlockStmt) bool {
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		// Check method calls: logger.Info(), log.Printf(), etc.
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && logCallNames[sel.Sel.Name] {
			found = true
			return false
		}
		// Check package-level calls: fmt.Println(), etc.
		return true
	})
	return found
}
