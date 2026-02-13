package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/report"
)

func runCheck(args []string) int {
	configPath, args := parseConfigFlag(args)

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: check requires a file argument")
		return 2
	}

	path := args[len(args)-1]

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %s is a directory, use 'scan' instead\n", path)
		return 2
	}

	scanDir := filepath.Dir(path)
	eng, err := buildEngine(scanDir, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	start := time.Now()
	result, err := eng.ScanFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	durationMs := time.Since(start).Milliseconds()

	report.Terminal(os.Stdout, result.Findings, result.FilesScanned, durationMs)

	if finding.HasErrors(result.Findings) {
		return 1
	}
	return 0
}
