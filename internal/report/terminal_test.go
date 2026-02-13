package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/perttulands/truthsayer/internal/finding"
)

func TestTerminal_ShowsAllFindingFields(t *testing.T) {
	findings := []finding.Finding{
		{
			Rule:       "silent-fallback.empty-error-check",
			Severity:   finding.SeverityError,
			File:       "pkg/handler.go",
			Line:       42,
			Code:       `if err != nil { return nil }`,
			Message:    "Error returned as nil without logging or wrapping",
			Suggestion: "Return the error or log it",
		},
	}

	var buf bytes.Buffer
	Terminal(&buf, findings, 5, 100)
	out := buf.String()

	// Must show file path
	if !strings.Contains(out, "pkg/handler.go") {
		t.Error("output missing file path")
	}
	// Must show line number
	if !strings.Contains(out, ":42") {
		t.Error("output missing line number")
	}
	// Must show code snippet
	if !strings.Contains(out, `if err != nil { return nil }`) {
		t.Error("output missing code snippet")
	}
	// Must show suggestion
	if !strings.Contains(out, "Return the error or log it") {
		t.Error("output missing suggestion")
	}
}

func TestTerminal_CodeSnippetTrimmed(t *testing.T) {
	findings := []finding.Finding{
		{
			Rule:     "test-rule",
			Severity: finding.SeverityWarning,
			File:     "a.go",
			Line:     1,
			Code:     "   \t  some indented code   ",
			Message:  "test message",
		},
	}

	var buf bytes.Buffer
	Terminal(&buf, findings, 1, 10)
	out := buf.String()

	// Code should be trimmed of leading/trailing whitespace
	if !strings.Contains(out, "some indented code") {
		t.Error("output should contain trimmed code snippet")
	}
}

func TestTerminal_SeveritySortOrder(t *testing.T) {
	// US-004: findings must appear sorted by severity (error > warning > info)
	findings := []finding.Finding{
		{Rule: "rule-err", Severity: finding.SeverityError, File: "a.go", Line: 10, Message: "error finding"},
		{Rule: "rule-warn", Severity: finding.SeverityWarning, File: "b.go", Line: 20, Message: "warning finding"},
		{Rule: "rule-info", Severity: finding.SeverityInfo, File: "c.go", Line: 30, Message: "info finding"},
	}

	var buf bytes.Buffer
	Terminal(&buf, findings, 3, 50)
	out := buf.String()

	// Verify severity labels appear in correct order
	errIdx := strings.Index(out, "ERROR")
	warnIdx := strings.Index(out, "WARN")
	infoIdx := strings.Index(out, "INFO")

	if errIdx == -1 || warnIdx == -1 || infoIdx == -1 {
		t.Fatalf("missing severity labels in output:\n%s", out)
	}
	if errIdx >= warnIdx {
		t.Errorf("ERROR (pos %d) should appear before WARN (pos %d)", errIdx, warnIdx)
	}
	if warnIdx >= infoIdx {
		t.Errorf("WARN (pos %d) should appear before INFO (pos %d)", warnIdx, infoIdx)
	}

	// Verify summary counts
	if !strings.Contains(out, "1 errors") {
		t.Error("summary should show 1 errors")
	}
	if !strings.Contains(out, "1 warnings") {
		t.Error("summary should show 1 warnings")
	}
	if !strings.Contains(out, "1 info") {
		t.Error("summary should show 1 info")
	}
}

func TestTerminal_CategoryCounts(t *testing.T) {
	findings := []finding.Finding{
		{Rule: "silent-fallback.empty-error-check", Severity: finding.SeverityError, File: "a.go", Line: 1, Message: "m1"},
		{Rule: "silent-fallback.ignored-error", Severity: finding.SeverityError, File: "a.go", Line: 2, Message: "m2"},
		{Rule: "bad-defaults.missing-pipefail", Severity: finding.SeverityError, File: "b.sh", Line: 1, Message: "m3"},
		{Rule: "trace-gaps.no-request-id", Severity: finding.SeverityWarning, File: "c.go", Line: 1, Message: "m4"},
	}

	var buf bytes.Buffer
	Terminal(&buf, findings, 10, 200)
	out := buf.String()

	// Must contain category breakdown
	if !strings.Contains(out, "silent-fallback: 2") {
		t.Errorf("expected category count for silent-fallback: 2, got:\n%s", out)
	}
	if !strings.Contains(out, "bad-defaults: 1") {
		t.Errorf("expected category count for bad-defaults: 1, got:\n%s", out)
	}
	if !strings.Contains(out, "trace-gaps: 1") {
		t.Errorf("expected category count for trace-gaps: 1, got:\n%s", out)
	}
}

func TestTerminal_CategoryCountsEmpty(t *testing.T) {
	var buf bytes.Buffer
	Terminal(&buf, nil, 5, 50)
	out := buf.String()

	// Should still show severity summary even with no findings
	if !strings.Contains(out, "0 errors") {
		t.Errorf("expected 0 errors in output:\n%s", out)
	}
	// No category line when there are no findings
	if strings.Contains(out, "Categories:") {
		t.Error("should not show Categories line when there are no findings")
	}
}

func TestTerminal_EmptyCode(t *testing.T) {
	findings := []finding.Finding{
		{
			Rule:     "test-rule",
			Severity: finding.SeverityInfo,
			File:     "b.go",
			Line:     5,
			Code:     "",
			Message:  "some message",
		},
	}

	var buf bytes.Buffer
	Terminal(&buf, findings, 1, 10)
	out := buf.String()

	// Should not crash with empty code, and should still show file/line/message
	if !strings.Contains(out, "b.go:5") {
		t.Error("output missing file:line even with empty code")
	}
	if !strings.Contains(out, "some message") {
		t.Error("output missing message even with empty code")
	}
}
