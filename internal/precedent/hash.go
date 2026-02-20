package precedent

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// HashViolation returns a stable hash for a violation instance.
// Hash input: rule_id + normalized file path + line + code snippet.
func HashViolation(ruleID, file string, line int, code string, scanRoot string) string {
	normalizedFile := normalizeFilePath(file, scanRoot)
	canonical := strings.Join([]string{
		strings.TrimSpace(ruleID),
		normalizedFile,
		strconv.Itoa(line),
		strings.TrimSpace(code),
	}, "\n")

	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

var (
	identifierPattern = regexp.MustCompile(`\b[A-Za-z_][A-Za-z0-9_]*\b`)
	numberPattern     = regexp.MustCompile(`\b\d+(?:\.\d+)?\b`)
	stringPattern     = regexp.MustCompile(`"(?:\\.|[^"])*"|'(?:\\.|[^'])*'`)
	whitespacePattern = regexp.MustCompile(`\s+`)
)

var languageKeywords = map[string]struct{}{
	"if": {}, "else": {}, "for": {}, "range": {}, "switch": {}, "case": {}, "default": {},
	"try": {}, "catch": {}, "finally": {}, "except": {}, "def": {}, "func": {}, "return": {},
	"nil": {}, "null": {}, "true": {}, "false": {}, "throw": {}, "raise": {}, "await": {},
	"async": {}, "const": {}, "let": {}, "var": {}, "class": {}, "new": {}, "import": {},
	"from": {}, "package": {}, "break": {}, "continue": {}, "pass": {}, "while": {},
}

// HashPattern returns a normalized pattern hash for precedent matching.
// Variable names, literals, and spacing differences are normalized away.
func HashPattern(ruleID, code string) string {
	normalized := normalizeCodePattern(code)
	canonical := strings.Join([]string{
		strings.TrimSpace(ruleID),
		normalized,
	}, "\n")
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

// HashFindingPattern returns a normalized pattern hash for a finding.
func HashFindingPattern(f finding.Finding) string {
	return HashPattern(f.Rule, f.Code)
}

func normalizeCodePattern(code string) string {
	value := strings.TrimSpace(code)
	if value == "" {
		return ""
	}
	value = stringPattern.ReplaceAllString(value, "str")
	value = numberPattern.ReplaceAllString(value, "num")
	value = identifierPattern.ReplaceAllStringFunc(value, func(tok string) string {
		lower := strings.ToLower(tok)
		if _, ok := languageKeywords[lower]; ok {
			return lower
		}
		return "id"
	})
	value = whitespacePattern.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func normalizeFilePath(file, scanRoot string) string {
	normalized := filepath.Clean(file)
	if scanRoot != "" {
		if rel, err := filepath.Rel(scanRoot, normalized); err == nil {
			if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				normalized = rel
			}
		}
	}
	normalized = filepath.ToSlash(normalized)
	return strings.TrimPrefix(normalized, "./")
}
