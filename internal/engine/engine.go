package engine

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/perttulands/truthsayer/internal/config"
	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
	"github.com/perttulands/truthsayer/internal/scanner"
)

// Result holds the output of a scan.
type Result struct {
	Findings     []finding.Finding
	FilesScanned int
}

// Engine orchestrates concurrent file scanning.
type Engine struct {
	reg             *rules.Registry
	goScanner       *scanner.GoScanner
	regexScanner    *scanner.RegexScanner
	excludeDirs     map[string]bool
	excludePatterns []string
	langs           *config.LanguageConfig

	jsOnce sync.Once
	js     *scanner.JSScanner
	pyOnce sync.Once
	py     *scanner.PyScanner
}

// New creates a scan engine from a rule registry.
func New(reg *rules.Registry) *Engine {
	return &Engine{
		reg:          reg,
		goScanner:    scanner.NewGoScanner(reg.ASTCheckers()),
		regexScanner: scanner.NewRegexScanner(reg.RegexCheckers()),
	}
}

// SetExcludeDirs overrides the default excluded directories.
func (e *Engine) SetExcludeDirs(dirs map[string]bool) {
	e.excludeDirs = dirs
}

// SetExcludePatterns sets glob patterns for files to exclude from scanning.
func (e *Engine) SetExcludePatterns(patterns []string) {
	e.excludePatterns = patterns
}

// SetLanguages configures which languages are enabled for scanning.
func (e *Engine) SetLanguages(lc *config.LanguageConfig) {
	e.langs = lc
}

func (e *Engine) langEnabled(lang string) bool {
	if e.langs == nil {
		return true
	}
	return e.langs.IsEnabled(lang)
}

// jsExt maps JS/TS file extensions for routing.
var jsExts = map[string]bool{
	".js":  true,
	".jsx": true,
	".ts":  true,
	".tsx": true,
	".mjs": true,
	".cjs": true,
}

func isJSExt(ext string) bool {
	return jsExts[ext]
}

func isPyExt(ext string) bool {
	return ext == ".py" || ext == ".pyi"
}

func (e *Engine) getJSScanner() *scanner.JSScanner {
	e.jsOnce.Do(func() {
		e.js = scanner.NewJSScanner(e.reg.JSASTCheckers())
	})
	return e.js
}

func (e *Engine) getPyScanner() *scanner.PyScanner {
	e.pyOnce.Do(func() {
		e.py = scanner.NewPyScanner(e.reg.PyASTCheckers())
	})
	return e.py
}

// extLang maps file extensions to language names for config filtering.
func extLang(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py", ".pyi":
		return "python"
	case ".sh", ".bash":
		return "bash"
	default:
		return ""
	}
}

// scanFileFindings scans a single file and returns findings.
func (e *Engine) scanFileFindings(path string) ([]finding.Finding, error) {
	ext := filepath.Ext(path)

	lang := extLang(ext)
	if lang != "" && !e.langEnabled(lang) {
		return nil, nil
	}

	switch {
	case ext == ".go":
		results, lines, err := e.goScanner.Scan(path)
		if err != nil {
			return nil, err
		}
		return append(results, e.regexScanner.ScanLines(path, lines)...), nil

	case isJSExt(ext):
		results, lines, err := e.getJSScanner().Scan(path)
		if err != nil {
			return nil, err
		}
		return append(results, e.regexScanner.ScanLines(path, lines)...), nil

	case isPyExt(ext):
		results, lines, err := e.getPyScanner().Scan(path)
		if err != nil {
			return nil, err
		}
		return append(results, e.regexScanner.ScanLines(path, lines)...), nil

	default:
		results, err := e.regexScanner.Scan(path)
		if err != nil {
			return nil, err
		}
		return results, nil
	}
}

// Scan walks the path and scans all matching files concurrently.
func (e *Engine) Scan(root string) (*Result, error) {
	files, err := Walk(root, e.excludeDirs, e.excludePatterns)
	if err != nil {
		return nil, fmt.Errorf("walk files under %s: %w", root, err)
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > len(files) {
		numWorkers = len(files)
	}
	if numWorkers == 0 {
		return &Result{}, nil
	}

	jobs := make(chan string, len(files))
	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	var mu sync.Mutex
	var allFindings []finding.Finding
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				fileFindings, err := e.scanFileFindings(path)
				if err != nil {
					continue
				}
				if len(fileFindings) > 0 {
					mu.Lock()
					allFindings = append(allFindings, fileFindings...)
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	allFindings = finding.Dedup(allFindings)
	e.reg.ApplyOverrides(allFindings)
	finding.Sort(allFindings)

	return &Result{
		Findings:     allFindings,
		FilesScanned: len(files),
	}, nil
}

// ScanFile scans a single file and returns findings.
func (e *Engine) ScanFile(path string) (*Result, error) {
	allFindings, err := e.scanFileFindings(path)
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}

	allFindings = finding.Dedup(allFindings)
	e.reg.ApplyOverrides(allFindings)
	finding.Sort(allFindings)

	return &Result{
		Findings:     allFindings,
		FilesScanned: 1,
	}, nil
}
