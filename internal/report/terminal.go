package report

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// maxCodeLen is the maximum length of a code snippet in terminal output.
const maxCodeLen = 120

// Terminal writes a human-readable scan report to w.
func Terminal(w io.Writer, findings []finding.Finding, filesScanned int, durationMs int64) {
	errors, warnings, infos := countBySeverity(findings)

	for _, f := range findings {
		label := severityLabel(f.Severity)
		fmt.Fprintf(w, "%s  %s\n", label, f.Rule)
		fmt.Fprintf(w, "  %s:%d\n", f.File, f.Line)
		if code := strings.TrimSpace(f.Code); code != "" {
			if len(code) > maxCodeLen {
				code = code[:maxCodeLen] + "…"
			}
			fmt.Fprintf(w, "  > %s\n", code)
		}
		fmt.Fprintf(w, "  %s\n", f.Message)
		if f.Suggestion != "" {
			fmt.Fprintf(w, "  → %s\n", f.Suggestion)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "%s\n", strings.Repeat("─", 50))
	fmt.Fprintf(w, "Summary: %d errors, %d warnings, %d info (%d files scanned in %dms)\n",
		errors, warnings, infos, filesScanned, durationMs)

	if cats := countByCategory(findings); len(cats) > 0 {
		fmt.Fprintf(w, "Categories: %s\n", formatCategoryCounts(cats))
	}
}

func severityLabel(s finding.Severity) string {
	switch s {
	case finding.SeverityError:
		return "ERROR "
	case finding.SeverityWarning:
		return "WARN  "
	case finding.SeverityInfo:
		return "INFO  "
	default:
		return "      "
	}
}

type categoryCount struct {
	name  string
	count int
}

func countByCategory(findings []finding.Finding) []categoryCount {
	counts := make(map[string]int)
	for _, f := range findings {
		cat := f.Rule
		if idx := strings.IndexByte(cat, '.'); idx > 0 {
			cat = cat[:idx]
		}
		counts[cat]++
	}
	out := make([]categoryCount, 0, len(counts))
	for name, count := range counts {
		out = append(out, categoryCount{name, count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count // highest count first
		}
		return out[i].name < out[j].name
	})
	return out
}

func formatCategoryCounts(cats []categoryCount) string {
	parts := make([]string, len(cats))
	for i, c := range cats {
		parts[i] = fmt.Sprintf("%s: %d", c.name, c.count)
	}
	return strings.Join(parts, ", ")
}

func countBySeverity(findings []finding.Finding) (errors, warnings, infos int) {
	for _, f := range findings {
		switch f.Severity {
		case finding.SeverityError:
			errors++
		case finding.SeverityWarning:
			warnings++
		case finding.SeverityInfo:
			infos++
		}
	}
	return
}
