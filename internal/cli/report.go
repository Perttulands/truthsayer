package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/perttulands/truthsayer/internal/engine"
	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/report"
	"github.com/perttulands/truthsayer/internal/rules"
)

const defaultReportOutput = "truthsayer-report.json"

func runReport(args []string) int {
	output := defaultReportOutput
	var path string

	// Parse --output flag
	for i := 0; i < len(args); i++ {
		if args[i] == "--output" && i+1 < len(args) {
			output = args[i+1]
			i++
		} else {
			path = args[i]
		}
	}

	if path == "" {
		fmt.Fprintln(os.Stderr, "error: report requires a path argument")
		return 2
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %s is not a directory\n", path)
		return 2
	}

	reg := rules.DefaultRegistry()
	eng := engine.New(reg)

	start := time.Now()
	result, err := eng.Scan(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	durationMs := time.Since(start).Milliseconds()

	f, err := os.Create(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	defer f.Close()

	scanTime := time.Now()
	if err := report.JSON(f, result.Findings, path, scanTime, result.FilesScanned, durationMs); err != nil {
		fmt.Fprintf(os.Stderr, "error writing report: %v\n", err)
		return 2
	}

	fmt.Fprintf(os.Stderr, "Report written to %s (%d findings, %d files scanned)\n", output, len(result.Findings), result.FilesScanned)

	if finding.HasErrors(result.Findings) {
		return 1
	}
	return 0
}
