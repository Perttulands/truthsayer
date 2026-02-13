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
