package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSONReport is the top-level structure for JSON scan output.
type JSONReport struct {
	Version    string        `json:"version"`
	ScanTime   string        `json:"scan_time"`
	Path       string        `json:"path"`
	DurationMs int64         `json:"duration_ms"`
	Findings   []JSONFinding `json:"findings"`
	Summary    JSONSummary   `json:"summary"`
}

// JSONFinding mirrors finding.Finding for JSON serialization.
type JSONFinding struct {
	Rule       string `json:"rule"`
	Severity   string `json:"severity"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Code       string `json:"code"`
	Context    string `json:"context,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// JSONSummary holds aggregated scan counts.
type JSONSummary struct {
	Total        int   `json:"total"`
	Errors       int   `json:"errors"`
	Warnings     int   `json:"warnings"`
	Info         int   `json:"info"`
	FilesScanned int   `json:"files_scanned"`
	DurationMs   int64 `json:"duration_ms"`
}

// JSON writes a JSON report to w.
func JSON(w io.Writer, findings []finding.Finding, path string, scanTime time.Time, filesScanned int, durationMs int64) error {
	jFindings := make([]JSONFinding, len(findings))
	for i, f := range findings {
		jFindings[i] = JSONFinding{
			Rule:       f.Rule,
			Severity:   string(f.Severity),
			File:       f.File,
			Line:       f.Line,
			Code:       f.Code,
			Context:    f.Context,
			Message:    f.Message,
			Suggestion: f.Suggestion,
		}
	}

	errors, warnings, infos := countBySeverity(findings)

	rpt := JSONReport{
		Version:    "1",
		ScanTime:   scanTime.UTC().Format(time.RFC3339),
		Path:       path,
		DurationMs: durationMs,
		Findings:   jFindings,
		Summary: JSONSummary{
			Total:        len(findings),
			Errors:       errors,
			Warnings:     warnings,
			Info:         infos,
			FilesScanned: filesScanned,
			DurationMs:   durationMs,
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rpt)
}
