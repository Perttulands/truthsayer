package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/report"
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

      - name: Truthsayer CI gate
        run: ./truthsayer ci .

      - name: Upload report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: truthsayer-report
          path: truthsayer-report.json
`

type ciOptions struct {
	path          string
	beadThreshold int
}

// runCI scans only newly changed lines in CI and creates beads for new errors.
func runCI(args []string) int {
	configPath, args := parseConfigFlag(args)

	opts, err := parseCIOptions(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
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

	changedLinesByFile, err := ciChangedLinesByFile(opts.path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	if len(changedLinesByFile) == 0 {
		fmt.Fprintln(os.Stdout, "No changed files detected for CI scan.")
		printBeadSummary(os.Stdout, nil)
		return 0
	}

	changedFiles := make([]string, 0, len(changedLinesByFile))
	for file := range changedLinesByFile {
		changedFiles = append(changedFiles, file)
	}
	sort.Strings(changedFiles)

	start := time.Now()
	var allFindings []finding.Finding
	filesScanned := 0

	for _, relPath := range changedFiles {
		fullPath := filepath.Join(opts.path, relPath)
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			continue
		}

		result, err := eng.ScanFile(fullPath)
		if err != nil {
			continue // skip unsupported or transient files
		}

		filtered := finding.FilterByLines(result.Findings, changedLinesByFile[relPath])
		allFindings = append(allFindings, filtered...)
		filesScanned++
	}

	finding.Sort(allFindings)
	durationMs := time.Since(start).Milliseconds()
	report.Terminal(os.Stdout, allFindings, filesScanned, durationMs)

	created, err := createErrorBeads(allFindings, opts.beadThreshold, newProblemBeadCreator())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	printBeadSummary(os.Stdout, created)

	if finding.HasErrors(allFindings) {
		return 1
	}
	return 0
}

func parseCIOptions(args []string) (ciOptions, error) {
	opts := ciOptions{
		path: ".",
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--bead-threshold":
			if i+1 >= len(args) {
				return ciOptions{}, fmt.Errorf("--bead-threshold requires a value")
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil {
				return ciOptions{}, fmt.Errorf("invalid --bead-threshold %q", args[i+1])
			}
			if n < 0 {
				return ciOptions{}, fmt.Errorf("--bead-threshold must be >= 0")
			}
			opts.beadThreshold = n
			i++
		case "--create-beads":
			// No-op for compatibility; CI always creates beads for new errors.
		default:
			opts.path = args[i]
		}
	}

	return opts, nil
}

func ciChangedLinesByFile(repoDir string) (map[string]map[int]bool, error) {
	diffBase, err := ciDiffBase(repoDir)
	if err != nil {
		return nil, err
	}

	files, err := ciChangedFiles(repoDir, diffBase)
	if err != nil {
		return nil, err
	}

	out := make(map[string]map[int]bool, len(files))
	for _, file := range files {
		lines, err := ciChangedLines(repoDir, diffBase, file)
		if err != nil {
			return nil, err
		}
		out[file] = lines
	}

	return out, nil
}

func ciDiffBase(repoDir string) (string, error) {
	if baseRef := strings.TrimSpace(os.Getenv("GITHUB_BASE_REF")); baseRef != "" {
		ref := "origin/" + baseRef
		if gitRefExists(repoDir, ref) {
			return ref + "...HEAD", nil
		}
	}

	if before := strings.TrimSpace(os.Getenv("GITHUB_EVENT_BEFORE")); before != "" && before != strings.Repeat("0", len(before)) {
		if gitRefExists(repoDir, before) {
			return before + "...HEAD", nil
		}
	}

	if gitRefExists(repoDir, "HEAD~1") {
		return "HEAD~1...HEAD", nil
	}

	return "", fmt.Errorf("unable to determine CI diff base (need GITHUB_BASE_REF, GITHUB_EVENT_BEFORE, or at least one parent commit)")
}

func gitRefExists(repoDir string, ref string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", ref)
	cmd.Dir = repoDir
	return cmd.Run() == nil
}

func ciChangedFiles(repoDir string, diffBase string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMRT", diffBase, "--")
	cmd.Dir = repoDir

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only failed: %w", err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func ciChangedLines(repoDir string, diffBase string, file string) (map[int]bool, error) {
	cmd := exec.Command("git", "diff", "--unified=0", "--diff-filter=ACMRT", diffBase, "--", file)
	cmd.Dir = repoDir

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff for %s failed: %w", file, err)
	}

	return parseChangedLines(string(out)), nil
}

func parseChangedLines(diffOutput string) map[int]bool {
	lines := make(map[int]bool)

	for _, line := range strings.Split(diffOutput, "\n") {
		if !strings.HasPrefix(line, "@@") {
			continue
		}
		start, count, ok := parseHunkHeader(line)
		if !ok || count <= 0 {
			continue
		}
		for n := start; n < start+count; n++ {
			lines[n] = true
		}
	}

	return lines
}

func parseHunkHeader(header string) (start int, count int, ok bool) {
	parts := strings.Fields(header)
	for _, part := range parts {
		if !strings.HasPrefix(part, "+") || part == "+++" {
			continue
		}

		rangePart := strings.TrimPrefix(part, "+")
		segs := strings.SplitN(rangePart, ",", 2)
		startVal, err := strconv.Atoi(segs[0])
		if err != nil {
			return 0, 0, false
		}

		countVal := 1
		if len(segs) == 2 {
			n, err := strconv.Atoi(segs[1])
			if err != nil {
				return 0, 0, false
			}
			countVal = n
		}
		return startVal, countVal, true
	}

	return 0, 0, false
}

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
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	workflowPath := filepath.Join(workflowDir, "truthsayer.yml")
	if err := os.WriteFile(workflowPath, []byte(ciWorkflow), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	fmt.Printf("Created GitHub Actions workflow at %s\n", workflowPath)
	return 0
}
