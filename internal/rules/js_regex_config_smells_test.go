package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// --- JSHardcodedAPIURL ---

func TestJSHardcodedAPIURL_Localhost(t *testing.T) {
	checker := &JSHardcodedAPIURL{}
	lines := []string{
		`const API_URL = 'http://localhost:3000/api';`,
	}
	findings := checker.CheckLines("src/config.ts", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSHardcodedAPIURL_IP(t *testing.T) {
	checker := &JSHardcodedAPIURL{}
	lines := []string{
		`const url = "http://127.0.0.1:8080/users";`,
	}
	findings := checker.CheckLines("src/api.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSHardcodedAPIURL_ApiDomain(t *testing.T) {
	checker := &JSHardcodedAPIURL{}
	lines := []string{
		`fetch("https://api.example.com/v1/users");`,
	}
	findings := checker.CheckLines("src/client.ts", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSHardcodedAPIURL_EnvVar(t *testing.T) {
	checker := &JSHardcodedAPIURL{}
	lines := []string{
		`const url = process.env.API_URL;`,
		`fetch(url + '/users');`,
	}
	findings := checker.CheckLines("src/client.ts", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for env var usage, got %d", len(findings))
	}
}

func TestJSHardcodedAPIURL_TestFileSkipped(t *testing.T) {
	checker := &JSHardcodedAPIURL{}
	lines := []string{`const url = 'http://localhost:3000';`}
	findings := checker.CheckLines("src/api.test.ts", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSHardcodedAPIURL_CommentSkipped(t *testing.T) {
	checker := &JSHardcodedAPIURL{}
	lines := []string{`// TODO: replace http://localhost:3000 with env var`}
	findings := checker.CheckLines("src/config.ts", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for comment, got %d", len(findings))
	}
}

// --- JSDotenvNoExample ---

func TestJSDotenvNoExample_Positive(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("API_KEY=secret\n"), 0644); err != nil {
		t.Fatal(err)
	}

	checker := &JSDotenvNoExample{}
	lines := []string{`API_KEY=secret`}
	findings := checker.CheckLines(envPath, lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSDotenvNoExample_WithExample(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	examplePath := filepath.Join(dir, ".env.example")
	if err := os.WriteFile(envPath, []byte("API_KEY=secret\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(examplePath, []byte("API_KEY=\n"), 0644); err != nil {
		t.Fatal(err)
	}

	checker := &JSDotenvNoExample{}
	lines := []string{`API_KEY=secret`}
	findings := checker.CheckLines(envPath, lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when .env.example exists, got %d", len(findings))
	}
}

func TestJSDotenvNoExample_NotEnvFile(t *testing.T) {
	checker := &JSDotenvNoExample{}
	lines := []string{`some content`}
	findings := checker.CheckLines("src/config.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for non-.env file, got %d", len(findings))
	}
}

func TestJSDotenvNoExample_EnvLocal(t *testing.T) {
	checker := &JSDotenvNoExample{}
	lines := []string{`API_KEY=secret`}
	findings := checker.CheckLines(".env.local", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for .env.local, got %d", len(findings))
	}
}
