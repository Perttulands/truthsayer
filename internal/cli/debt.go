package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/perttulands/truthsayer/internal/debt"
)

type debtOptions struct {
	path   string
	format string
}

func runDebt(args []string) int {
	opts, err := parseDebtOptions(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	path := opts.path
	if path == "" {
		path = debt.DefaultPath
	} else if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, debt.DefaultPath)
	}

	store := debt.NewStore(path)
	entries, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	if opts.format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(entries); err != nil {
			fmt.Fprintf(os.Stderr, "error: write debt json: %v\n", err)
			return 2
		}
		return 0
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stdout, "No advisory debt entries found.")
		return 0
	}

	for _, e := range entries {
		fmt.Fprintf(os.Stdout, "- %s %s:%d %s\n", e.RuleID, e.File, e.Line, e.Reasoning)
	}
	fmt.Fprintf(os.Stdout, "Total advisory debt entries: %d\n", len(entries))
	return 0
}

func parseDebtOptions(args []string) (debtOptions, error) {
	opts := debtOptions{format: "text"}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 >= len(args) {
				return debtOptions{}, fmt.Errorf("--format requires a value")
			}
			opts.format = args[i+1]
			i++
		default:
			opts.path = args[i]
		}
	}
	if opts.format != "text" && opts.format != "json" {
		return debtOptions{}, fmt.Errorf("unknown format %q (use text or json)", opts.format)
	}
	return opts, nil
}

