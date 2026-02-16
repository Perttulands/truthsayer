package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyHardcodedCredentials detects hardcoded passwords/API keys/secrets in Python source.
type PyHardcodedCredentials struct{}

func (h *PyHardcodedCredentials) Meta() Rule {
	return Rule{
		ID:          "config-smells.hardcoded-credentials-py",
		Category:    "config-smells",
		Name:        "Hardcoded credentials in Python",
		Description: "password/api_key/secret string literal in source — use environment variables or secrets manager",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeRegex,
	}
}

// Matches assignments like: password = "...", api_key = "...", secret = "...", token = "..."
// Requires a non-empty string value (at least one char between quotes).
var hardcodedCredentialPattern = regexp.MustCompile(`\b(?:password|passwd|api_key|apikey|secret|secret_key|token|auth_token)\s*=\s*["'][^"']+["']`)

func (h *PyHardcodedCredentials) CheckLines(path string, lines []string) []finding.Finding {
	if pyIsTestFile(path) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if hardcodedCredentialPattern.MatchString(strings.ToLower(line)) {
			findings = append(findings, finding.Finding{
				Rule:       h.Meta().ID,
				Severity:   h.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Hardcoded credential in source code",
				Suggestion: "Use os.environ['KEY'] or a secrets manager instead of hardcoded values",
			})
		}
	}
	return findings
}

// PyRequirementsUnpinned detects unpinned dependencies in requirements.txt.
type PyRequirementsUnpinned struct{}

func (r *PyRequirementsUnpinned) Meta() Rule {
	return Rule{
		ID:          "config-smells.requirements-unpinned",
		Category:    "config-smells",
		Name:        "Unpinned requirements",
		Description: "requirements.txt entries without == version pins — non-reproducible builds",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".txt"},
		ScanType:    ScanTypeRegex,
	}
}

// Matches lines that look like package names without == pinning.
// Valid package lines: "requests", "flask>=2.0", "django~=4.0", but NOT "requests==2.28.0".
var requirementLinePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*`)

func (r *PyRequirementsUnpinned) CheckLines(path string, lines []string) []finding.Finding {
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "requirements") || !strings.HasSuffix(base, ".txt") {
		return nil
	}

	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines, comments, options (e.g., -r, --index-url)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") {
			continue
		}
		if !requirementLinePattern.MatchString(trimmed) {
			continue
		}
		// Check if line has == pinning
		if strings.Contains(trimmed, "==") {
			continue
		}
		findings = append(findings, finding.Finding{
			Rule:       r.Meta().ID,
			Severity:   r.Meta().Severity,
			File:       path,
			Line:       i + 1,
			Code:       line,
			Message:    "Unpinned dependency in requirements file",
			Suggestion: "Pin the version with ==: e.g., requests==2.28.0",
		})
	}
	return findings
}
