package report

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
)

func TestJSON_EmptyFindings(t *testing.T) {
	var buf bytes.Buffer
	scanTime := time.Date(2026, 2, 13, 15, 0, 0, 0, time.UTC)
	err := JSON(&buf, nil, "/projects/myapp", scanTime, 87, 342)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var rpt JSONReport
	if err := json.Unmarshal(buf.Bytes(), &rpt); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if rpt.Version != "1" {
		t.Errorf("version = %q, want %q", rpt.Version, "1")
	}
	if rpt.Path != "/projects/myapp" {
		t.Errorf("path = %q, want %q", rpt.Path, "/projects/myapp")
	}
	if rpt.DurationMs != 342 {
		t.Errorf("duration_ms = %d, want 342", rpt.DurationMs)
	}
	if len(rpt.Findings) != 0 {
		t.Errorf("findings len = %d, want 0", len(rpt.Findings))
	}
	if rpt.Summary.Total != 0 {
		t.Errorf("summary.total = %d, want 0", rpt.Summary.Total)
	}
	if rpt.Summary.FilesScanned != 87 {
		t.Errorf("summary.files_scanned = %d, want 87", rpt.Summary.FilesScanned)
	}
}

func TestJSON_WithFindings(t *testing.T) {
	findings := []finding.Finding{
		{
			Rule:       "silent-fallback.empty-error-check",
			Severity:   finding.SeverityError,
			File:       "pkg/handler.go",
			Line:       42,
			Code:       "if err != nil { return nil }",
			Context:    ">> 42 | if err != nil { return nil }",
			Message:    "Error returned as nil without logging or wrapping",
			Suggestion: "Return the error or log it",
		},
		{
			Rule:     "bad-defaults.missing-pipefail",
			Severity: finding.SeverityWarning,
			File:     "deploy.sh",
			Line:     1,
			Code:     "#!/bin/bash",
			Message:  "Missing set -euo pipefail",
		},
		{
			Rule:     "trace-gaps.info-rule",
			Severity: finding.SeverityInfo,
			File:     "main.go",
			Line:     10,
			Code:     "func main() {",
			Message:  "Consider adding logging",
		},
	}

	var buf bytes.Buffer
	scanTime := time.Date(2026, 2, 13, 15, 0, 0, 0, time.UTC)
	err := JSON(&buf, findings, "/projects/myapp", scanTime, 87, 342)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var rpt JSONReport
	if err := json.Unmarshal(buf.Bytes(), &rpt); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if rpt.Version != "1" {
		t.Errorf("version = %q, want %q", rpt.Version, "1")
	}
	if rpt.ScanTime != "2026-02-13T15:00:00Z" {
		t.Errorf("scan_time = %q, want RFC3339 UTC", rpt.ScanTime)
	}
	if rpt.Path != "/projects/myapp" {
		t.Errorf("path = %q", rpt.Path)
	}
	if rpt.DurationMs != 342 {
		t.Errorf("duration_ms = %d, want 342", rpt.DurationMs)
	}
	if len(rpt.Findings) != 3 {
		t.Fatalf("findings len = %d, want 3", len(rpt.Findings))
	}

	// Verify first finding fields
	f := rpt.Findings[0]
	if f.Rule != "silent-fallback.empty-error-check" {
		t.Errorf("finding[0].rule = %q", f.Rule)
	}
	if f.Severity != "error" {
		t.Errorf("finding[0].severity = %q", f.Severity)
	}
	if f.Line != 42 {
		t.Errorf("finding[0].line = %d", f.Line)
	}
	if f.Suggestion != "Return the error or log it" {
		t.Errorf("finding[0].suggestion = %q", f.Suggestion)
	}
	if f.Context == "" {
		t.Errorf("finding[0].context should be present")
	}

	// Verify summary counts
	if rpt.Summary.Total != 3 {
		t.Errorf("summary.total = %d, want 3", rpt.Summary.Total)
	}
	if rpt.Summary.Errors != 1 {
		t.Errorf("summary.errors = %d, want 1", rpt.Summary.Errors)
	}
	if rpt.Summary.Warnings != 1 {
		t.Errorf("summary.warnings = %d, want 1", rpt.Summary.Warnings)
	}
	if rpt.Summary.Info != 1 {
		t.Errorf("summary.info = %d, want 1", rpt.Summary.Info)
	}
	if rpt.Summary.FilesScanned != 87 {
		t.Errorf("summary.files_scanned = %d, want 87", rpt.Summary.FilesScanned)
	}
	if rpt.Summary.DurationMs != 342 {
		t.Errorf("summary.duration_ms = %d, want 342", rpt.Summary.DurationMs)
	}
}

func TestJSON_ValidFormat(t *testing.T) {
	var buf bytes.Buffer
	scanTime := time.Date(2026, 2, 13, 15, 0, 0, 0, time.UTC)
	err := JSON(&buf, nil, "/test", scanTime, 10, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must be valid JSON
	if !json.Valid(buf.Bytes()) {
		t.Errorf("output is not valid JSON: %s", buf.String())
	}

	// Must be indented (pretty-printed)
	if buf.Bytes()[1] != '\n' {
		t.Errorf("expected indented JSON, got: %.40s...", buf.String())
	}
}
