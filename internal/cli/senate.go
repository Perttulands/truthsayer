package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	case "apply":
		return runSenateApply(args[1:])
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

func runSenateApply(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: senate apply requires a verdict file path")
		return 2
	}
	verdictPath := args[0]
	repoDir := "."
	if len(args) > 1 {
		repoDir = args[1]
	}

	verdict, err := senate.ParseVerdictFile(verdictPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	if verdict.Status != senate.StatusApproved {
		fmt.Fprintf(os.Stdout, "No amendments applied (verdict status: %s)\n", verdict.Status)
		return 0
	}

	storePath := filepath.Join(repoDir, senate.DefaultAmendmentsPath)
	added, err := senate.NewAmendmentStore(storePath).ApplyVerdict(verdict, time.Now().UTC())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: apply amendments: %v\n", err)
		return 2
	}
	auditPath := filepath.Join(repoDir, senate.DefaultAmendmentsAuditPath)
	if err := senate.AppendAudit(auditPath, added); err != nil {
		fmt.Fprintf(os.Stderr, "error: append amendment audit: %v\n", err)
		return 2
	}
	fmt.Fprintf(os.Stdout, "Applied %d amendment(s) from verdict %s\n", len(added), verdict.ID)
	return 0
}
