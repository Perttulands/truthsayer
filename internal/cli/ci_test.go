package cli

import (
	"os"
	"os/exec"
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
	// Must contain ci quality gate command
	if !strings.Contains(s, "truthsayer ci .") {
		t.Error("workflow does not contain 'truthsayer ci .' quality gate step")
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

func TestCI_CreatesBeadsForNewErrors(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	writeFile(t, dir, "good.go", goClean)
	commitAll(t, dir, "add clean file")
	writeFile(t, dir, "bad.go", goWithError)
	commitAll(t, dir, "introduce error")

	fake := &fakeBeadCreator{}
	oldFactory := newProblemBeadCreator
	newProblemBeadCreator = func() problemBeadCreator { return fake }
	defer func() { newProblemBeadCreator = oldFactory }()

	out := captureStdout(t, func() {
		code := runCI([]string{dir})
		if code != 1 {
			t.Fatalf("expected exit code 1 for new errors, got %d", code)
		}
	})

	if len(fake.calls) != 1 {
		t.Fatalf("expected 1 bead for new error group, got %d", len(fake.calls))
	}
	if !strings.Contains(out, "Beads created: 1") {
		t.Fatalf("expected bead summary, got:\n%s", out)
	}
}

func TestCI_DoesNotCreateBeadsForExistingUnchangedErrors(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	writeFile(t, dir, "bad.go", goWithError)
	commitAll(t, dir, "baseline with error")
	writeFile(t, dir, "good.go", goClean)
	commitAll(t, dir, "clean change only")

	fake := &fakeBeadCreator{}
	oldFactory := newProblemBeadCreator
	newProblemBeadCreator = func() problemBeadCreator { return fake }
	defer func() { newProblemBeadCreator = oldFactory }()

	out := captureStdout(t, func() {
		code := runCI([]string{dir})
		if code != 0 {
			t.Fatalf("expected exit code 0 for no new errors, got %d", code)
		}
	})

	if len(fake.calls) != 0 {
		t.Fatalf("expected 0 bead calls, got %d", len(fake.calls))
	}
	if !strings.Contains(out, "Beads created: 0") {
		t.Fatalf("expected zero bead summary, got:\n%s", out)
	}
}

func commitAll(t *testing.T, dir string, message string) {
	t.Helper()

	add := exec.Command("git", "add", ".")
	add.Dir = dir
	if out, err := add.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %s: %v", out, err)
	}

	commit := exec.Command("git", "commit", "-m", message)
	commit.Dir = dir
	if out, err := commit.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %s: %v", out, err)
	}
}
