package report

import (
	"fmt"
	"io"
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
