package precedent

import (
	"path/filepath"
	"testing"

	"github.com/perttulands/truthsayer/internal/finding"
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

func TestHashPattern_IgnoresVariableRenames(t *testing.T) {
	a := HashPattern("silent-fallback.empty-error-check", "if err != nil { return nil }")
	b := HashPattern("silent-fallback.empty-error-check", "if failure != nil { return nil }")
	if a != b {
		t.Fatalf("expected same pattern hash for variable rename, got %q != %q", a, b)
	}
}

func TestHashPattern_NormalizesWhitespaceAndLiterals(t *testing.T) {
	a := HashPattern("error-context.generic-message", `throw new Error("oops 123")`)
	b := HashPattern("error-context.generic-message", "throw   new   Error('different') ")
	if a != b {
		t.Fatalf("expected same pattern hash for normalized literals and whitespace, got %q != %q", a, b)
	}
}

func TestHashPattern_DiffersForDifferentStructure(t *testing.T) {
	a := HashPattern("silent-fallback.hidden-failure-bash", "cmd || true")
	b := HashPattern("silent-fallback.hidden-failure-bash", "cmd || :")
	if a == b {
		t.Fatal("expected different hash for different logical pattern")
	}
}

func TestHashFindingPattern_UsesRuleAndCode(t *testing.T) {
	f := finding.Finding{
		Rule: "silent-fallback.empty-error-check",
		Code: "if err != nil { return nil }",
	}
	a := HashFindingPattern(f)
	b := HashPattern(f.Rule, f.Code)
	if a != b {
		t.Fatalf("expected finding pattern hash to match HashPattern, got %q != %q", a, b)
	}
}
