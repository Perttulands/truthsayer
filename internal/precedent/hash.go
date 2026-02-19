package precedent

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strconv"
	"strings"
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
