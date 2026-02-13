package engine

import (
	"os"
	"path/filepath"
	"strings"
)

var defaultExcludeDirs = map[string]bool{
	".git":         true,
	"vendor":       true,
	"node_modules": true,
	"testdata":     true,
}

// DefaultExcludeDirs returns a copy of the default excluded directories map.
func DefaultExcludeDirs() map[string]bool {
	m := make(map[string]bool, len(defaultExcludeDirs))
	for k, v := range defaultExcludeDirs {
		m[k] = v
	}
	return m
}

var supportedExts = map[string]bool{
	".go":   true,
	".sh":   true,
	".bash": true,
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

	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
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
		if matchesAnyPattern(d.Name(), excludePatterns) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// matchesAnyPattern returns true if name matches any of the glob patterns.
func matchesAnyPattern(name string, patterns []string) bool {
	for _, p := range patterns {
		if matched, _ := filepath.Match(p, name); matched {
			return true
		}
	}
	return false
}
