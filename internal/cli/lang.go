package cli

import (
	"fmt"
	"strings"

	"github.com/perttulands/truthsayer/internal/config"
)

// langAliases maps CLI language aliases to canonical language names.
var langAliases = map[string]string{
	"go":         "go",
	"js":         "javascript",
	"javascript": "javascript",
	"ts":         "typescript",
	"typescript": "typescript",
	"python":     "python",
	"py":         "python",
	"bash":       "bash",
	"shell":      "bash",
	"sh":         "bash",
}

// langExts maps canonical language names to their file extensions.
var langExts = map[string][]string{
	"go":         {".go"},
	"javascript": {".js", ".jsx", ".mjs", ".cjs"},
	"typescript": {".ts", ".tsx"},
	"python":     {".py", ".pyi"},
	"bash":       {".sh", ".bash"},
}

// parseLangFlag parses a comma-separated --lang value into a LanguageConfig
// where only the specified languages are enabled (all others disabled).
func parseLangFlag(value string) (*config.LanguageConfig, error) {
	parts := strings.Split(value, ",")
	enabled := make(map[string]bool)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		canonical, ok := langAliases[strings.ToLower(p)]
		if !ok {
			return nil, fmt.Errorf("unknown language %q (valid: go, js, ts, python, bash)", p)
		}
		enabled[canonical] = true
	}
	if len(enabled) == 0 {
		return nil, fmt.Errorf("--lang requires at least one language")
	}

	f, t := false, true
	lc := &config.LanguageConfig{
		Go:         &f,
		JavaScript: &f,
		TypeScript: &f,
		Python:     &f,
		Bash:       &f,
	}
	for lang := range enabled {
		switch lang {
		case "go":
			lc.Go = &t
		case "javascript":
			lc.JavaScript = &t
		case "typescript":
			lc.TypeScript = &t
		case "python":
			lc.Python = &t
		case "bash":
			lc.Bash = &t
		}
	}
	return lc, nil
}

// langFilterExts returns the set of file extensions for the given --lang value.
// Used by the rules command to filter rules by FileTypes.
func langFilterExts(value string) (map[string]bool, error) {
	parts := strings.Split(value, ",")
	exts := make(map[string]bool)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		canonical, ok := langAliases[strings.ToLower(p)]
		if !ok {
			return nil, fmt.Errorf("unknown language %q (valid: go, js, ts, python, bash)", p)
		}
		for _, ext := range langExts[canonical] {
			exts[ext] = true
		}
	}
	if len(exts) == 0 {
		return nil, fmt.Errorf("--lang requires at least one language")
	}
	return exts, nil
}
