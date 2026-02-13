package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const ciWorkflow = `name: Truthsayer

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  truthsayer:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Build Truthsayer
        run: go build -o truthsayer ./cmd/truthsayer

      - name: Truthsayer scan (JSON report)
        run: ./truthsayer scan --format json . > truthsayer-report.json
        continue-on-error: true

      - name: Truthsayer scan (quality gate)
        run: ./truthsayer scan .

      - name: Upload report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: truthsayer-report
          path: truthsayer-report.json
`

// runCIInit generates a GitHub Actions workflow file for Truthsayer.
func runCIInit(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: ci init requires a path argument")
		return 2
	}

	repoDir := args[len(args)-1]

	// Verify it's a git repo
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s is not a git repository\n", repoDir)
		return 2
	}

	workflowDir := filepath.Join(repoDir, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	workflowPath := filepath.Join(workflowDir, "truthsayer.yml")
	if err := os.WriteFile(workflowPath, []byte(ciWorkflow), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	fmt.Printf("Created GitHub Actions workflow at %s\n", workflowPath)
	return 0
}
