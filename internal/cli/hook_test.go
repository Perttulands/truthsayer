package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initGitRepo creates a git repo in dir with initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init failed: %s: %v", out, err)
		}
	}
}

// stageFile adds a file to the git index.
func stageFile(t *testing.T, dir, name string) {
	t.Helper()
	cmd := exec.Command("git", "add", name)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %s: %v", out, err)
	}
}

func TestHook_ExitCode1_StagedFileWithErrors(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "bad.go", goWithError)
	stageFile(t, dir, "bad.go")

	code := runHook([]string{dir})
	if code != 1 {
		t.Errorf("expected exit code 1 (staged file has errors), got %d", code)
	}
}

func TestHook_ExitCode0_StagedCleanFile(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "good.go", goClean)
	stageFile(t, dir, "good.go")

	code := runHook([]string{dir})
	if code != 0 {
		t.Errorf("expected exit code 0 (staged file is clean), got %d", code)
	}
}

func TestHook_ExitCode0_NoStagedFiles(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	code := runHook([]string{dir})
	if code != 0 {
		t.Errorf("expected exit code 0 (no staged files), got %d", code)
	}
}

func TestHook_IgnoresUnstagedFiles(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	// Write a bad file but don't stage it
	writeFile(t, dir, "bad.go", goWithError)
	// Stage only the clean file
	writeFile(t, dir, "good.go", goClean)
	stageFile(t, dir, "good.go")

	code := runHook([]string{dir})
	if code != 0 {
		t.Errorf("expected exit code 0 (only clean file staged), got %d", code)
	}
}

func TestHook_ExitCode2_NotGitRepo(t *testing.T) {
	dir := t.TempDir()

	code := runHook([]string{dir})
	if code != 2 {
		t.Errorf("expected exit code 2 (not a git repo), got %d", code)
	}
}

func TestHook_ExitCode2_NoArgs(t *testing.T) {
	code := runHook(nil)
	if code != 2 {
		t.Errorf("expected exit code 2 (no args), got %d", code)
	}
}

func TestHookInstall_CreatesPreCommitHook(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	code := runHookInstall([]string{dir})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("hook file not created: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("hook file is not executable")
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) == 0 {
		t.Error("hook file is empty")
	}
}

func TestHookInstall_ExitCode2_NotGitRepo(t *testing.T) {
	dir := t.TempDir()

	code := runHookInstall([]string{dir})
	if code != 2 {
		t.Errorf("expected exit code 2 (not a git repo), got %d", code)
	}
}

func TestHookInstall_RespectsConfig(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "bad.go", goWithError)
	stageFile(t, dir, "bad.go")
	writeFile(t, dir, ".truthsayer.toml", `
[rules]
disable = ["silent-fallback.empty-error-check"]
`)
	stageFile(t, dir, ".truthsayer.toml")

	code := runHook([]string{dir})
	if code != 0 {
		t.Errorf("expected exit code 0 (rule disabled via config), got %d", code)
	}
}
