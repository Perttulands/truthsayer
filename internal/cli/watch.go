package cli

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/perttulands/truthsayer/internal/diff"
	"github.com/perttulands/truthsayer/internal/engine"
	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/report"
	"github.com/perttulands/truthsayer/internal/watcher"
)

func runWatch(args []string) int {
	configPath, args := parseConfigFlag(args)

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: watch requires a path argument")
		return 2
	}

	path := args[len(args)-1]

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %s is not a directory (use check for single files)\n", path)
		return 2
	}

	reg, err := buildRegistry(path, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	eng := engine.New(reg)

	w, err := watcher.New(path, 100*time.Millisecond)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	defer w.Close()

	fmt.Fprintf(os.Stderr, "Watching %s for changes... (Ctrl+C to stop)\n", path)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	tracker := diff.NewTracker()
	sawError := false

	for {
		select {
		case filePath, ok := <-w.Events():
			if !ok {
				return exitCode(sawError)
			}

			changedLines, err := tracker.Update(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "diff error: %v\n", err)
				continue
			}

			start := time.Now()
			result, err := eng.ScanFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
				continue
			}
			durationMs := time.Since(start).Milliseconds()

			filtered := finding.FilterByLines(result.Findings, changedLines)
			if len(filtered) > 0 {
				report.Terminal(os.Stdout, filtered, result.FilesScanned, durationMs)
				for _, f := range filtered {
					if f.Severity == finding.SeverityError {
						sawError = true
						break
					}
				}
			}

		case <-sig:
			fmt.Fprintln(os.Stderr, "\nStopping watch...")
			return exitCode(sawError)
		}
	}
}

func exitCode(sawError bool) int {
	if sawError {
		return 1
	}
	return 0
}
