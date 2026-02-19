package precedent

import (
	"path/filepath"
	"testing"
)

func TestHashViolation_RelativeToScanRoot(t *testing.T) {
	root := filepath.Join("repo")
	file := filepath.Join("repo", "src", "main.go")

	hashA := HashViolation("silent-fallback.empty-error-check", file, 12, "if err != nil {", root)
	hashB := HashViolation("silent-fallback.empty-error-check", filepath.Join("src", "main.go"), 12, "if err != nil {", "")

	if hashA != hashB {
		t.Fatalf("expected same hash for equivalent relative path, got %q != %q", hashA, hashB)
	}
}

func TestHashViolation_ChangesWhenViolationChanges(t *testing.T) {
	base := HashViolation("silent-fallback.empty-error-check", "src/main.go", 12, "if err != nil {", "")
	otherLine := HashViolation("silent-fallback.empty-error-check", "src/main.go", 13, "if err != nil {", "")
	otherRule := HashViolation("bad-defaults.no-timeout", "src/main.go", 12, "if err != nil {", "")

	if base == otherLine {
		t.Fatal("expected different hash when line changes")
	}
	if base == otherRule {
		t.Fatal("expected different hash when rule changes")
	}
}
