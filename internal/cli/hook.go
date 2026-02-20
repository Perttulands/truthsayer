package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/precedent"
	"github.com/perttulands/truthsayer/internal/report"
)

// runHook scans git staged files for anti-patterns. Returns 1 if errors found, 2 for tool errors.
func runHook(args []string) int {
	configPath, args := parseConfigFlag(args)

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: hook requires a path argument")
		return 2
	}

	repoDir := args[len(args)-1]

	staged, err := stagedFiles(repoDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	if len(staged) == 0 {
		return 0
	}

	eng, err := buildEngine(repoDir, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	start := time.Now()
	var allFindings []finding.Finding
	filesScanned := 0

	for _, f := range staged {
		if isTestdata(f) {
			continue
		}
		path := filepath.Join(repoDir, f)
		result, err := eng.ScanFile(path)
		if err != nil {
			continue // skip files that can't be scanned (e.g. binary, deleted)
		}
		allFindings = append(allFindings, result.Findings...)
		filesScanned++
	}

	finding.Sort(allFindings)
	durationMs := time.Since(start).Milliseconds()

	report.Terminal(os.Stdout, allFindings, filesScanned, durationMs)

	if len(allFindings) == 0 {
		return 0
	}

	// TS-013: pre-commit now runs the scan -> judge pipeline.
	// If judgment is unavailable, fall back to deterministic scan gating.
	judgeInputPath, cleanup, err := writeHookFindingsReport(repoDir, allFindings, filesScanned, durationMs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: hook judge setup failed, falling back to scan gate: %v\n", err)
		if finding.HasErrors(allFindings) {
			return 1
		}
		return 0
	}
	defer cleanup()

	judgeCode := runJudge([]string{
		"--format", "text",
		"--precedents", filepath.Join(repoDir, precedent.DefaultPath),
		judgeInputPath,
	})
	if judgeCode == 2 {
		fmt.Fprintln(os.Stderr, "warning: hook judgment unavailable, falling back to scan gate")
		if finding.HasErrors(allFindings) {
			return 1
		}
		return 0
	}
	return judgeCode
}

// stagedFiles returns the list of staged files (added, copied, modified) in the repo.
func stagedFiles(repoDir string) ([]string, error) {
	argv := []string{"git", "diff", "--cached", "--name-only", "--diff-filter=ACM"}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// isTestdata returns true if the file path is inside a testdata directory.
func isTestdata(path string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(path), "/") {
		if seg == "testdata" {
			return true
		}
	}
	return false
}

func writeHookFindingsReport(repoDir string, findings []finding.Finding, filesScanned int, durationMs int64) (string, func(), error) {
	f, err := os.CreateTemp(repoDir, ".truthsayer-hook-findings-*.json")
	if err != nil {
		return "", nil, fmt.Errorf("create temp findings report: %w", err)
	}
	path := f.Name()
	closeAndCleanup := func() {
		_ = f.Close()
		_ = os.Remove(path)
	}

	if err := report.JSON(f, findings, repoDir, time.Now(), filesScanned, durationMs); err != nil {
		closeAndCleanup()
		return "", nil, fmt.Errorf("write temp findings report: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", nil, fmt.Errorf("close temp findings report: %w", err)
	}

	return path, func() { _ = os.Remove(path) }, nil
}

const preCommitHook = `#!/bin/bash
# Truthsayer pre-commit hook — scans staged files for anti-patterns
truthsayer hook .
`

// runHookInstall installs the Truthsayer pre-commit hook into the git repo.
func runHookInstall(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: hook install requires a path argument")
		return 2
	}

	repoDir := args[len(args)-1]
	hooksDir := filepath.Join(repoDir, ".git", "hooks")

	if _, err := os.Stat(hooksDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s is not a git repository\n", repoDir)
		return 2
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(preCommitHook), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	fmt.Printf("Installed pre-commit hook at %s\n", hookPath)
	return 0
}
