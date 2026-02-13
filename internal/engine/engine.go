package engine

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

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
				var fileFindings []finding.Finding
				ext := filepath.Ext(path)

				if ext == ".go" {
					results, err := e.goScanner.Scan(path)
					if err == nil {
						fileFindings = append(fileFindings, results...)
					}
				}

				results, err := e.regexScanner.Scan(path)
				if err == nil {
					fileFindings = append(fileFindings, results...)
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
	var allFindings []finding.Finding
	ext := filepath.Ext(path)

	if ext == ".go" {
		results, err := e.goScanner.Scan(path)
		if err != nil {
			return nil, fmt.Errorf("scan go file %s: %w", path, err)
		}
		allFindings = append(allFindings, results...)
	}

	results, err := e.regexScanner.Scan(path)
	if err != nil {
		return nil, fmt.Errorf("scan text rules in %s: %w", path, err)
	}
	allFindings = append(allFindings, results...)

	allFindings = finding.Dedup(allFindings)
	e.reg.ApplyOverrides(allFindings)
	finding.Sort(allFindings)

	return &Result{
		Findings:     allFindings,
		FilesScanned: 1,
	}, nil
}
