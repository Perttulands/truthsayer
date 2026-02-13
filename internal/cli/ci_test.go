package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIInit_CreatesWorkflowFile(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	code := runCIInit([]string{dir})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	workflowPath := filepath.Join(dir, ".github", "workflows", "truthsayer.yml")
	info, err := os.Stat(workflowPath)
	if err != nil {
		t.Fatalf("workflow file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("workflow file is empty")
	}
}

func TestCIInit_WorkflowContainsScanStep(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	runCIInit([]string{dir})

	workflowPath := filepath.Join(dir, ".github", "workflows", "truthsayer.yml")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	// Must contain scan command that fails on errors
	if !strings.Contains(s, "truthsayer scan") {
		t.Error("workflow does not contain 'truthsayer scan' step")
	}
	// Must contain JSON report generation
	if !strings.Contains(s, "--format json") {
		t.Error("workflow does not contain JSON format output step")
	}
	// Must reference Go setup
	if !strings.Contains(s, "setup-go") {
		t.Error("workflow does not contain Go setup action")
	}
	// Must trigger on push and PR
	if !strings.Contains(s, "push") {
		t.Error("workflow does not trigger on push")
	}
	if !strings.Contains(s, "pull_request") {
		t.Error("workflow does not trigger on pull_request")
	}
}

func TestCIInit_WorkflowUploadsArtifact(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	runCIInit([]string{dir})

	workflowPath := filepath.Join(dir, ".github", "workflows", "truthsayer.yml")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "upload-artifact") {
		t.Error("workflow does not upload report artifact")
	}
}

func TestCIInit_ExitCode2_NotGitRepo(t *testing.T) {
	dir := t.TempDir()

	code := runCIInit([]string{dir})
	if code != 2 {
		t.Errorf("expected exit code 2 (not a git repo), got %d", code)
	}
}

func TestCIInit_ExitCode2_NoArgs(t *testing.T) {
	code := runCIInit(nil)
	if code != 2 {
		t.Errorf("expected exit code 2 (no args), got %d", code)
	}
}

func TestCIInit_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	code := runCIInit([]string{dir})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Verify .github/workflows directory was created
	info, err := os.Stat(filepath.Join(dir, ".github", "workflows"))
	if err != nil {
		t.Fatalf("workflows dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected .github/workflows to be a directory")
	}
}
