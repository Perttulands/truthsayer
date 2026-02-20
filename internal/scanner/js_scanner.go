package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
)

// jsLang identifies the tree-sitter language grammar to use.
type jsLang int

const (
	jsLangJS jsLang = iota
	jsLangTS
	jsLangTSX
)

var jsExtMap = map[string]jsLang{
	".js":  jsLangJS,
	".mjs": jsLangJS,
	".cjs": jsLangJS,
	".jsx": jsLangJS,
	".ts":  jsLangTS,
	".tsx": jsLangTSX,
}

// parserPools holds a sync.Pool per language so parsers with the correct
// grammar already set can be reused without re-initialization.
var parserPools = map[jsLang]*sync.Pool{
	jsLangJS: {New: func() any {
		return newJSParser(jsLangJS)
	}},
	jsLangTS: {New: func() any {
		return newJSParser(jsLangTS)
	}},
	jsLangTSX: {New: func() any {
		return newJSParser(jsLangTSX)
	}},
}

func newJSParser(lang jsLang) *sitter.Parser {
	p := sitter.NewParser()
	switch lang {
	case jsLangJS:
		p.SetLanguage(javascript.GetLanguage())
	case jsLangTS:
		p.SetLanguage(typescript.GetLanguage())
	case jsLangTSX:
		p.SetLanguage(tsx.GetLanguage())
	}
	return p
}

// JSScanner scans JavaScript and TypeScript files using tree-sitter AST analysis.
type JSScanner struct {
	checkers []rules.JSASTChecker
}

// NewJSScanner creates a scanner with the given JS AST checkers.
func NewJSScanner(checkers []rules.JSASTChecker) *JSScanner {
	return &JSScanner{checkers: checkers}
}

// Scan parses a JS/TS file and runs all JSASTCheckers against it.
// Returns findings and source lines (for reuse by regex scanner).
func (s *JSScanner) Scan(path string) ([]finding.Finding, []string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	lang, ok := jsExtMap[ext]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported JS/TS extension: %s", ext)
	}

	source, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}

	pool := parserPools[lang]
	rawParser := pool.Get()
	parser, ok := rawParser.(*sitter.Parser)
	if !ok || parser == nil {
		parser = newJSParser(lang)
	}
	defer pool.Put(parser)

	// REASON: scanner API is synchronous and currently has no caller context to propagate.
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", path, err)
	}

	var findings []finding.Finding
	for _, checker := range s.checkers {
		findings = append(findings, checker.CheckJSAST(tree, source, path)...)
	}

	lines := strings.Split(string(source), "\n")
	return findings, lines, nil
}
