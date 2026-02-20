package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/judge"
	"github.com/perttulands/truthsayer/internal/precedent"
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

func TestHook_JudgmentNotGuiltyPassesAndWritesPrecedent(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "bad.go", goWithError)
	stageFile(t, dir, "bad.go")

	fake := &fakeFindingJudge{
		verdict: judge.Verdict{
			Verdict:            judge.VerdictNotGuilty,
			Reasoning:          "known intentional fallback",
			Confidence:         0.92,
			PrecedentDecision:  precedent.DecisionAllow,
			PrecedentRationale: "approved cleanup exception",
			Source:             "llm",
		},
	}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	code := runHook([]string{dir})
	if code != 0 {
		t.Fatalf("expected exit code 0 from not_guilty judgment, got %d", code)
	}

	records, err := precedent.NewStore(filepath.Join(dir, precedent.DefaultPath)).Load()
	if err != nil {
		t.Fatalf("load precedents: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("expected precedents to be written by hook judgment")
	}
	if records[len(records)-1].Decision != precedent.DecisionAllow {
		t.Fatalf("expected last precedent decision allow, got %q", records[len(records)-1].Decision)
	}
}

func TestHook_JudgmentGuiltyBlocksCommit(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "bad.go", goWithError)
	stageFile(t, dir, "bad.go")

	fake := &fakeFindingJudge{
		verdict: judge.Verdict{
			Verdict:            judge.VerdictGuilty,
			Reasoning:          "silent error suppression remains harmful",
			Confidence:         0.88,
			PrecedentDecision:  precedent.DecisionDeny,
			PrecedentRationale: "must return or wrap error",
			Source:             "llm",
			InputTokens:        120,
			OutputTokens:       42,
		},
	}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	code := runHook([]string{dir})
	if code != 1 {
		t.Fatalf("expected exit code 1 from guilty judgment, got %d", code)
	}

	records, err := precedent.NewStore(filepath.Join(dir, precedent.DefaultPath)).Load()
	if err != nil {
		t.Fatalf("load precedents: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("expected precedents to be written by hook judgment")
	}
	if records[len(records)-1].Decision != precedent.DecisionDeny {
		t.Fatalf("expected last precedent decision deny, got %q", records[len(records)-1].Decision)
	}
	if records[len(records)-1].CreatedAt.Before(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected recent precedent timestamp, got %s", records[len(records)-1].CreatedAt)
	}
}
