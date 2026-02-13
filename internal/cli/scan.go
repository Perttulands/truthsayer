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

func runScan(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: scan requires a path argument")
		return 2
	}

	path := args[len(args)-1]

	// Verify path exists
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

	report.Terminal(os.Stdout, result.Findings, result.FilesScanned, durationMs)

	for _, f := range result.Findings {
		if f.Severity == finding.SeverityError {
			return 1
		}
	}
	return 0
}
