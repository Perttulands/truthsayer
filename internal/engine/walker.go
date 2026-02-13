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

// Walk recursively lists scannable files under root, skipping excluded directories.
func Walk(root string, excludeDirs map[string]bool) ([]string, error) {
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
		if supportedExts[ext] {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
