package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// MissingGitignore detects .env files that are not ignored by .gitignore.
type MissingGitignore struct{}

func (m *MissingGitignore) Meta() Rule {
	return Rule{
		ID:          "config-smells.missing-gitignore",
		Category:    "config-smells",
		Name:        "Missing .env in .gitignore",
		Description: ".env file exists but is not listed in .gitignore",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{"*"},
		ScanType:    ScanTypeRegex,
	}
}

func (m *MissingGitignore) CheckLines(path string, lines []string) []finding.Finding {
	base := filepath.Base(path)

	if base == ".gitignore" {
		return m.checkGitignore(path, lines)
	}
	if base != ".env" {
		return nil
	}

	dir := filepath.Dir(path)
	gitignorePath := filepath.Join(dir, ".gitignore")
	gitignoreLines, err := readPlainLines(gitignorePath)
	if err == nil && hasEnvIgnore(gitignoreLines) {
		return nil
	}

	code := ".env"
	if len(lines) > 0 {
		code = lines[0]
	}

	return []finding.Finding{
		{
			Rule:       m.Meta().ID,
			Severity:   m.Meta().Severity,
			File:       path,
			Line:       1,
			Code:       code,
			Message:    ".env file is not ignored by .gitignore",
			Suggestion: "Add '.env' to .gitignore to avoid leaking local secrets",
		},
	}
}

func (m *MissingGitignore) checkGitignore(path string, lines []string) []finding.Finding {
	dir := filepath.Dir(path)
	envPath := filepath.Join(dir, ".env")

	code := ".gitignore"
	if len(lines) > 0 {
		code = lines[0]
	}

	if _, err := os.Stat(envPath); err != nil {
		return m.handleEnvStatError(path, code, err)
	}
	if hasEnvIgnore(lines) {
		return nil
	}

	return []finding.Finding{
		{
			Rule:       m.Meta().ID,
			Severity:   m.Meta().Severity,
			File:       path,
			Line:       1,
			Code:       code,
			Message:    ".env file exists but .gitignore does not ignore it",
			Suggestion: "Add '.env' to .gitignore to avoid leaking local secrets",
		},
	}
}

func readPlainLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	return strings.Split(text, "\n"), nil
}

func (m *MissingGitignore) handleEnvStatError(path, code string, err error) []finding.Finding {
	if os.IsNotExist(err) {
		return nil
	}
	return []finding.Finding{
		{
			Rule:       m.Meta().ID,
			Severity:   m.Meta().Severity,
			File:       path,
			Line:       1,
			Code:       code,
			Message:    fmt.Sprintf("Unable to verify .env presence: %v", err),
			Suggestion: "Ensure the scanner can read .env metadata in this directory",
		},
	}
}

func hasEnvIgnore(lines []string) bool {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "!") {
			continue
		}

		candidate := strings.TrimPrefix(trimmed, "/")
		if candidate == ".env" ||
			strings.HasPrefix(candidate, ".env.") ||
			strings.HasPrefix(candidate, ".env*") ||
			strings.HasSuffix(candidate, "/.env") ||
			strings.Contains(candidate, "/.env.") ||
			strings.Contains(candidate, "/.env*") {
			return true
		}
	}
	return false
}
