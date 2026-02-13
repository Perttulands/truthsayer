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
