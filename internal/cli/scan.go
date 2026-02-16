package cli

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/report"
)

type scanOptions struct {
	format        string
	path          string
	lang          string
	parallel      int
	createBeads   bool
	beadThreshold int
}

func runScan(args []string) int {
	configPath, args := parseConfigFlag(args)

	opts, err := parseScanOptions(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	if opts.path == "" {
		fmt.Fprintln(os.Stderr, "error: scan requires a path argument")
		return 2
	}

	info, err := os.Stat(opts.path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %s is not a directory\n", opts.path)
		return 2
	}

	eng, err := buildEngine(opts.path, configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	// --lang flag overrides config language settings
	if opts.lang != "" {
		lc, err := parseLangFlag(opts.lang)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 2
		}
		eng.SetLanguages(lc)
	}
	if opts.parallel > 0 {
		eng.SetParallelism(opts.parallel)
	}

	start := time.Now()
	result, err := eng.Scan(opts.path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	durationMs := time.Since(start).Milliseconds()

	switch opts.format {
	case "json":
		if err := report.JSON(os.Stdout, result.Findings, opts.path, time.Now(), result.FilesScanned, durationMs); err != nil {
			fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
			return 2
		}
	default:
		report.Terminal(os.Stdout, result.Findings, result.FilesScanned, durationMs)
	}

	if opts.createBeads {
		created, err := createErrorBeads(result.Findings, opts.beadThreshold, newProblemBeadCreator())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 2
		}

		if opts.format == "json" {
			printBeadSummary(os.Stderr, created)
		} else {
			printBeadSummary(os.Stdout, created)
		}
	}

	if finding.HasErrors(result.Findings) {
		return 1
	}
	return 0
}

func parseScanOptions(args []string) (scanOptions, error) {
	opts := scanOptions{
		format: "text",
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 >= len(args) {
				return scanOptions{}, fmt.Errorf("--format requires a value")
			}
			opts.format = args[i+1]
			i++
		case "--lang":
			if i+1 >= len(args) {
				return scanOptions{}, fmt.Errorf("--lang requires a value")
			}
			opts.lang = args[i+1]
			i++
		case "--create-beads":
			opts.createBeads = true
		case "--parallel":
			opts.parallel = runtime.NumCPU()
			if i+1 < len(args) {
				next := args[i+1]
				if n, err := strconv.Atoi(next); err == nil {
					if n < 1 {
						return scanOptions{}, fmt.Errorf("--parallel must be >= 1")
					}
					opts.parallel = n
					i++
				} else if strings.HasPrefix(next, "-") {
					// bare --parallel flag with another option following
				}
			}
		case "--bead-threshold":
			if i+1 >= len(args) {
				return scanOptions{}, fmt.Errorf("--bead-threshold requires a value")
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil {
				return scanOptions{}, fmt.Errorf("invalid --bead-threshold %q", args[i+1])
			}
			if n < 0 {
				return scanOptions{}, fmt.Errorf("--bead-threshold must be >= 0")
			}
			opts.beadThreshold = n
			i++
		default:
			opts.path = args[i]
		}
	}

	if opts.format != "text" && opts.format != "json" {
		return scanOptions{}, fmt.Errorf("unknown format %q (use text or json)", opts.format)
	}

	return opts, nil
}
