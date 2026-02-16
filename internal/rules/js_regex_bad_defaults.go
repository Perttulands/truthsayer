package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSTsIgnore detects @ts-ignore without explanation.
type JSTsIgnore struct{}

func (t *JSTsIgnore) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.ts-ignore",
		Category:    "bad-defaults",
		Name:        "@ts-ignore without explanation",
		Description: "@ts-ignore suppresses type errors without explanation — use @ts-expect-error with reason",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".ts", ".tsx"},
		ScanType:    ScanTypeRegex,
	}
}

// Matches @ts-ignore optionally followed by whitespace but no explanatory text.
var tsIgnorePattern = regexp.MustCompile(`//\s*@ts-ignore\s*$`)

func (t *JSTsIgnore) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		if tsIgnorePattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       t.Meta().ID,
				Severity:   t.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "@ts-ignore without explanation",
				Suggestion: "Use @ts-expect-error with a reason, or fix the type error",
			})
		}
	}
	return findings
}

// JSEslintDisableNoReason detects eslint-disable without comment.
type JSEslintDisableNoReason struct{}

func (e *JSEslintDisableNoReason) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.eslint-disable-no-reason",
		Category:    "bad-defaults",
		Name:        "eslint-disable without reason",
		Description: "eslint-disable suppresses lint rules without justification",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeRegex,
	}
}

// Matches eslint-disable (inline or block) with optional rule name but no trailing " -- reason".
// eslint-disable-next-line rule-name -- reason  → OK (has " -- ")
// eslint-disable-next-line rule-name            → flagged
// eslint-disable-next-line                      → flagged
var eslintDisablePattern = regexp.MustCompile(`//\s*eslint-disable(?:-next-line|-line)?\b`)

func (e *JSEslintDisableNoReason) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		if eslintDisablePattern.MatchString(line) {
			// Check if there's a -- comment after the rule names
			if strings.Contains(line, " -- ") {
				continue
			}
			findings = append(findings, finding.Finding{
				Rule:       e.Meta().ID,
				Severity:   e.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "eslint-disable without reason",
				Suggestion: "Add a reason: // eslint-disable-next-line rule-name -- explanation",
			})
		}
	}
	return findings
}

// JSNoStrictMode detects CommonJS .cjs files without 'use strict'.
type JSNoStrictMode struct{}

func (n *JSNoStrictMode) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.no-strict-mode",
		Category:    "bad-defaults",
		Name:        "CommonJS without strict mode",
		Description: "CommonJS file without 'use strict' — sloppy mode enables silent failures",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".cjs"},
		ScanType:    ScanTypeRegex,
	}
}

var useStrictPattern = regexp.MustCompile(`['"]use strict['"]`)

func (n *JSNoStrictMode) CheckLines(path string, lines []string) []finding.Finding {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}
		if useStrictPattern.MatchString(line) {
			return nil
		}
		// First non-comment, non-empty line reached without 'use strict'
		return []finding.Finding{{
			Rule:       n.Meta().ID,
			Severity:   n.Meta().Severity,
			File:       path,
			Line:       1,
			Code:       lines[0],
			Message:    "CommonJS file without 'use strict'",
			Suggestion: "Add 'use strict'; at the top of the file",
		}}
	}
	return nil
}
