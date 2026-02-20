package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/perttulands/truthsayer/internal/precedent"
)

func runWarmup(args []string) int {
	configPath, args := parseConfigFlag(args)
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: warmup requires a repository path")
		return 2
	}
	repoPath := args[0]
	judgeArgs := args[1:]

	info, err := os.Stat(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %s is not a directory\n", repoPath)
		return 2
	}

	eng, err := buildEngine(repoPath, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	start := time.Now()
	result, err := eng.Scan(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	durationMs := time.Since(start).Milliseconds()

	findingsPath, cleanup, err := writeHookFindingsReport(repoPath, result.Findings, result.FilesScanned, durationMs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: warmup findings export failed: %v\n", err)
		return 2
	}
	defer cleanup()

	finalJudgeArgs := []string{
		"--format", "text",
		"--precedents", filepath.Join(repoPath, precedent.DefaultPath),
	}
	finalJudgeArgs = append(finalJudgeArgs, judgeArgs...)
	finalJudgeArgs = append(finalJudgeArgs, findingsPath)

	code := runJudge(finalJudgeArgs)
	if code == 2 {
		return 2
	}

	fmt.Fprintf(os.Stdout, "Warmup complete: files=%d findings=%d precedents=%s\n",
		result.FilesScanned, len(result.Findings), filepath.Join(repoPath, precedent.DefaultPath))
	return 0
}
