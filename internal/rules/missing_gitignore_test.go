package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMissingGitignore_EnvWithoutGitignore(t *testing.T) {
	checker := &MissingGitignore{}
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings := checker.CheckLines(envPath, []string{"KEY=value"})
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestMissingGitignore_EnvIgnored(t *testing.T) {
	checker := &MissingGitignore{}
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	gitignorePath := filepath.Join(dir, ".gitignore")

	if err := os.WriteFile(envPath, []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gitignorePath, []byte(".env\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings := checker.CheckLines(envPath, []string{"KEY=value"})
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestMissingGitignore_CheckGitignoreContent(t *testing.T) {
	checker := &MissingGitignore{}
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	gitignorePath := filepath.Join(dir, ".gitignore")

	if err := os.WriteFile(envPath, []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gitignorePath, []byte("# local files\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings := checker.CheckLines(gitignorePath, []string{"# local files"})
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestMissingGitignore_CheckGitignoreContent_EnvIgnored(t *testing.T) {
	checker := &MissingGitignore{}
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	gitignorePath := filepath.Join(dir, ".gitignore")

	if err := os.WriteFile(envPath, []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gitignorePath, []byte(".env\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings := checker.CheckLines(gitignorePath, []string{".env"})
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
