package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// LongFunctionNoLog detects functions >30 lines with no log/trace/print call.
// Functions with names matching common pure-logic patterns (checkers, parsers,
// validators, formatters) are excluded since they typically don't need logging.
type LongFunctionNoLog struct{}

func (l *LongFunctionNoLog) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.long-function-no-log",
		Category:    "trace-gaps",
		Name:        "Long function without logging",
		Description: "Function >30 lines with no logging or tracing calls",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".go"},
		ScanType:    ScanTypeAST,
	}
}

var logCallNames = map[string]bool{
	"Print": true, "Println": true, "Printf": true,
	"Fprintf": true, "Fprintln": true,
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
		if bodyLines <= 30 {
			continue
		}
		if hasLogCall(fn.Body) {
			continue
		}
		if isPureLogicFunction(fn.Name.Name) {
			continue
		}
		pos := fset.Position(fn.Pos())
		findings = append(findings, finding.Finding{
			Rule:       l.Meta().ID,
			Severity:   l.Meta().Severity,
			File:       fname,
			Line:       pos.Line,
			Code:       sourceLine(lines, pos.Line),
			Message:    fmt.Sprintf("Function %s is %d lines with no logging", fn.Name.Name, bodyLines),
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
		// Check method calls: logger.Info(), log.Printf(), fmt.Fprintf(), etc.
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && logCallNames[sel.Sel.Name] {
			found = true
			return false
		}
		return true
	})
	return found
}

// isPureLogicFunction returns true for function names that conventionally
// perform pure data transformation and don't need logging. Unexported
// helper functions are also excluded since they typically don't represent
// top-level operations worth tracing.
func isPureLogicFunction(name string) bool {
	// Unexported helper functions rarely need logging
	if len(name) > 0 && name[0] >= 'a' && name[0] <= 'z' {
		return true
	}

	lower := strings.ToLower(name)
	for _, prefix := range []string{
		"is", "has", "can", "should",
		"check", "parse", "format", "count",
		"match", "find", "extract", "convert",
		"compare", "sort", "filter", "validate",
		"marshal", "unmarshal", "encode", "decode",
		"render", "compile", "build",
		"default", "new", "make", "create",
	} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}
