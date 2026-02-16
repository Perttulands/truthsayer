package rules

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSHardcodedAPIURL detects hardcoded localhost/API URLs in source.
type JSHardcodedAPIURL struct{}

func (h *JSHardcodedAPIURL) Meta() Rule {
	return Rule{
		ID:          "config-smells.hardcoded-api-url",
		Category:    "config-smells",
		Name:        "Hardcoded API URL",
		Description: "Hardcoded http://localhost or https://api. URL in source — use environment variables",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeRegex,
	}
}

var hardcodedURLPattern = regexp.MustCompile(`['"\x60]https?://(localhost|127\.0\.0\.1|api\.)[\w./:?&=-]*['"\x60]`)

func (h *JSHardcodedAPIURL) CheckLines(path string, lines []string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}
		if hardcodedURLPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       h.Meta().ID,
				Severity:   h.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Hardcoded API URL in source code",
				Suggestion: "Use an environment variable (process.env.API_URL) or config file instead",
			})
		}
	}
	return findings
}

// JSDotenvNoExample detects .env files without a corresponding .env.example.
type JSDotenvNoExample struct{}

func (d *JSDotenvNoExample) Meta() Rule {
	return Rule{
		ID:          "config-smells.dotenv-no-example",
		Category:    "config-smells",
		Name:        ".env without .env.example",
		Description: ".env file without .env.example — teammates won't know which vars are required",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{"*"},
		ScanType:    ScanTypeRegex,
	}
}

func (d *JSDotenvNoExample) CheckLines(path string, lines []string) []finding.Finding {
	base := filepath.Base(path)
	if base != ".env" {
		return nil
	}

	dir := filepath.Dir(path)
	examplePath := filepath.Join(dir, ".env.example")
	if _, err := os.Stat(examplePath); err == nil {
		return nil
	}

	code := ".env"
	if len(lines) > 0 {
		code = lines[0]
	}

	return []finding.Finding{{
		Rule:       d.Meta().ID,
		Severity:   d.Meta().Severity,
		File:       path,
		Line:       1,
		Code:       code,
		Message:    ".env file without .env.example",
		Suggestion: "Create a .env.example with placeholder values so teammates know which variables are required",
	}}
}
