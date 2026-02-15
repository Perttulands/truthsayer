package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
)

// RegexScanner scans files using line-based regex checkers.
type RegexScanner struct {
	checkers []rules.RegexChecker
}

// NewRegexScanner creates a scanner with the given regex checkers.
func NewRegexScanner(checkers []rules.RegexChecker) *RegexScanner {
	return &RegexScanner{checkers: checkers}
}

// Scan reads a file's lines and runs all matching regex checkers.
func (s *RegexScanner) Scan(path string) ([]finding.Finding, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, fmt.Errorf("read lines from %s: %w", path, err)
	}
	return s.ScanLines(path, lines), nil
}

// ScanLines runs all matching regex checkers against pre-read lines.
// Use this to avoid re-reading files that were already read for AST parsing.
func (s *RegexScanner) ScanLines(path string, lines []string) []finding.Finding {
	ext := filepath.Ext(path)
	var findings []finding.Finding
	for _, checker := range s.checkers {
		meta := checker.Meta()
		if !matchesExt(meta.FileTypes, ext) {
			continue
		}
		findings = append(findings, checker.CheckLines(path, lines)...)
	}
	return findings
}

func readLines(path string) ([]string, error) {
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

func matchesExt(allowed []string, ext string) bool {
	for _, a := range allowed {
		if a == "*" || a == ext {
			return true
		}
	}
	return false
}
