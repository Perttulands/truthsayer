package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var defaultExcludeDirs = map[string]bool{
	".git":         true,
	"vendor":       true,
	"node_modules": true,
	"testdata":     true,
	"__pycache__":  true,
	".venv":        true,
	"dist":         true,
	"build":        true,
}

// DefaultExcludeDirs returns a copy of the default excluded directories map.
func DefaultExcludeDirs() map[string]bool {
	m := make(map[string]bool, len(defaultExcludeDirs))
	for k, v := range defaultExcludeDirs {
		m[k] = v
	}
	return m
}

var defaultExcludePatterns = []string{
	"*.min.js",
	"*.bundle.js",
	"*.pyc",
}

var supportedExts = map[string]bool{
	".go":   true,
	".js":   true,
	".jsx":  true,
	".ts":   true,
	".tsx":  true,
	".mjs":  true,
	".cjs":  true,
	".py":   true,
	".pyi":  true,
	".sh":   true,
	".bash": true,
	".rs":   true,
	".toml": true,
	".yaml": true,
	".yml":  true,
	".json": true,
	".env":  true,
}

// Walk recursively lists scannable files under root, skipping excluded directories
// and files matching exclude patterns (glob matched against base name).
func Walk(root string, excludeDirs map[string]bool, excludePatterns []string) ([]string, error) {
	if excludeDirs == nil {
		excludeDirs = defaultExcludeDirs
	}

	allPatterns := make([]string, 0, len(defaultExcludePatterns)+len(excludePatterns))
	allPatterns = append(allPatterns, defaultExcludePatterns...)
	allPatterns = append(allPatterns, excludePatterns...)

	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk directory entry %s: %w", path, err)
		}
		if d.IsDir() {
			name := d.Name()
			if excludeDirs[name] || strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if !supportedExts[ext] {
			return nil
		}
		matched, matchErr := matchesAnyPattern(d.Name(), allPatterns)
		if matchErr != nil {
			return fmt.Errorf("match exclude patterns for %s: %w", path, matchErr)
		}
		if matched {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// matchesAnyPattern returns true if name matches any of the glob patterns.
func matchesAnyPattern(name string, patterns []string) (bool, error) {
	for _, p := range patterns {
		matched, err := filepath.Match(p, name)
		if err != nil {
			return false, fmt.Errorf("invalid exclude pattern %q: %w", p, err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}
