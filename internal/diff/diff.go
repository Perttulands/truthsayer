package diff

import (
	"fmt"
	"os"
	"strings"
)

// Tracker keeps file content snapshots and computes changed line numbers.
type Tracker struct {
	snapshots map[string][]string
}

// NewTracker creates a Tracker with no snapshots.
func NewTracker() *Tracker {
	return &Tracker{snapshots: make(map[string][]string)}
}

// Update reads the file, computes which lines changed since the last snapshot,
// stores the new snapshot, and returns the set of changed 1-based line numbers.
// For a file seen for the first time (no prior snapshot), all lines are "changed".
// Returns nil map when there are no changes.
func (t *Tracker) Update(path string) (map[int]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}

	newLines := splitLines(string(data))
	oldLines, seen := t.snapshots[path]
	t.snapshots[path] = newLines

	if !seen {
		// New file: all lines are changed
		changed := make(map[int]bool, len(newLines))
		for i := range newLines {
			changed[i+1] = true
		}
		return changed, nil
	}

	return changedLines(oldLines, newLines), nil
}

// splitLines splits content into lines, dropping a trailing empty line
// from a final newline.
func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// changedLines computes which 1-based line numbers in newLines differ from oldLines
// using a longest-common-subsequence approach to align unchanged lines.
func changedLines(old, new []string) map[int]bool {
	// Compute LCS table
	m, n := len(old), len(new)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if old[i-1] == new[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to find which new lines are NOT in the LCS (i.e., changed/added)
	matched := make(map[int]bool, dp[m][n])
	i, j := m, n
	for i > 0 && j > 0 {
		if old[i-1] == new[j-1] {
			matched[j] = true // 1-based line in new
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	// Changed lines = new lines NOT matched by LCS
	changed := make(map[int]bool)
	for line := 1; line <= n; line++ {
		if !matched[line] {
			changed[line] = true
		}
	}
	return changed
}
