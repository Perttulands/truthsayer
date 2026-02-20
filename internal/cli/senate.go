package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/perttulands/truthsayer/internal/senate"
)

func runSenate(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: senate requires a subcommand (parse)")
		return 2
	}
	switch args[0] {
	case "parse":
		return runSenateParse(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown senate subcommand %q\n", args[0])
		return 2
	}
}

func runSenateParse(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: senate parse requires a verdict file path")
		return 2
	}
	verdict, err := senate.ParseVerdictFile(args[len(args)-1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(verdict); err != nil {
		fmt.Fprintf(os.Stderr, "error: write senate parse output: %v\n", err)
		return 2
	}
	return 0
}

