package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/perttulands/truthsayer/internal/engine"
	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/report"
)

func runScan(args []string) int {
	configPath, args := parseConfigFlag(args)

	format := "text"
	var path string

	for i := 0; i < len(args); i++ {
		if args[i] == "--format" && i+1 < len(args) {
			format = args[i+1]
			i++
		} else {
			path = args[i]
		}
	}

	if path == "" {
		fmt.Fprintln(os.Stderr, "error: scan requires a path argument")
		return 2
	}

	if format != "text" && format != "json" {
		fmt.Fprintf(os.Stderr, "error: unknown format %q (use text or json)\n", format)
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

	reg, err := buildRegistry(path, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	eng := engine.New(reg)

	start := time.Now()
	result, err := eng.Scan(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	durationMs := time.Since(start).Milliseconds()

	switch format {
	case "json":
		if err := report.JSON(os.Stdout, result.Findings, path, time.Now(), result.FilesScanned, durationMs); err != nil {
			fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
			return 2
		}
	default:
		report.Terminal(os.Stdout, result.Findings, result.FilesScanned, durationMs)
	}

	if finding.HasErrors(result.Findings) {
		return 1
	}
	return 0
}
