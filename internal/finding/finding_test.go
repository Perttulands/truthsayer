package finding

import (
	"testing"
)

func TestDedup(t *testing.T) {
	findings := []Finding{
		{Rule: "rule-a", File: "foo.go", Line: 10, Severity: SeverityError},
		{Rule: "rule-a", File: "foo.go", Line: 10, Severity: SeverityError}, // dup
		{Rule: "rule-b", File: "foo.go", Line: 10, Severity: SeverityWarning},
		{Rule: "rule-a", File: "foo.go", Line: 20, Severity: SeverityError}, // different line
	}
	got := Dedup(findings)
	if len(got) != 3 {
		t.Fatalf("expected 3 unique findings, got %d", len(got))
	}
}

func TestSort(t *testing.T) {
	findings := []Finding{
		{Rule: "r", Severity: SeverityInfo, File: "b.go", Line: 1},
		{Rule: "r", Severity: SeverityError, File: "a.go", Line: 5},
		{Rule: "r", Severity: SeverityWarning, File: "a.go", Line: 3},
		{Rule: "r", Severity: SeverityError, File: "a.go", Line: 1},
	}
	Sort(findings)

	expected := []struct {
		sev  Severity
		file string
		line int
	}{
		{SeverityError, "a.go", 1},
		{SeverityError, "a.go", 5},
		{SeverityWarning, "a.go", 3},
		{SeverityInfo, "b.go", 1},
	}

	for i, e := range expected {
		if findings[i].Severity != e.sev || findings[i].File != e.file || findings[i].Line != e.line {
			t.Errorf("index %d: expected %v %s:%d, got %v %s:%d",
				i, e.sev, e.file, e.line,
				findings[i].Severity, findings[i].File, findings[i].Line)
		}
	}
}

func TestSort_SeverityPrecedence(t *testing.T) {
	// US-004: severity order is error > warning > info, regardless of file/line
	findings := []Finding{
		{Rule: "r1", Severity: SeverityInfo, File: "aaa.go", Line: 1},
		{Rule: "r2", Severity: SeverityWarning, File: "bbb.go", Line: 1},
		{Rule: "r3", Severity: SeverityError, File: "zzz.go", Line: 99},
	}
	Sort(findings)

	// Error must come first even though zzz.go > aaa.go alphabetically
	if findings[0].Severity != SeverityError {
		t.Errorf("first finding should be error, got %s", findings[0].Severity)
	}
	if findings[1].Severity != SeverityWarning {
		t.Errorf("second finding should be warning, got %s", findings[1].Severity)
	}
	if findings[2].Severity != SeverityInfo {
		t.Errorf("third finding should be info, got %s", findings[2].Severity)
	}
}

func TestFilterByLines(t *testing.T) {
	findings := []Finding{
		{Rule: "r1", File: "a.go", Line: 1, Severity: SeverityError},
		{Rule: "r2", File: "a.go", Line: 5, Severity: SeverityWarning},
		{Rule: "r3", File: "a.go", Line: 10, Severity: SeverityInfo},
	}

	changedLines := map[int]bool{1: true, 10: true}
	got := FilterByLines(findings, changedLines)

	if len(got) != 2 {
		t.Fatalf("expected 2 filtered findings, got %d", len(got))
	}
	if got[0].Line != 1 || got[1].Line != 10 {
		t.Errorf("expected lines 1 and 10, got %d and %d", got[0].Line, got[1].Line)
	}
}

func TestFilterByLines_Empty(t *testing.T) {
	findings := []Finding{
		{Rule: "r1", File: "a.go", Line: 1, Severity: SeverityError},
	}

	// No changed lines — filter out everything
	got := FilterByLines(findings, map[int]bool{})
	if len(got) != 0 {
		t.Fatalf("expected 0 filtered findings, got %d", len(got))
	}
}

func TestFilterByLines_NilPassesAll(t *testing.T) {
	findings := []Finding{
		{Rule: "r1", File: "a.go", Line: 1, Severity: SeverityError},
		{Rule: "r2", File: "a.go", Line: 5, Severity: SeverityWarning},
	}

	// nil changedLines means "all lines changed" (new file or full scan)
	got := FilterByLines(findings, nil)
	if len(got) != 2 {
		t.Fatalf("expected 2 findings (all pass through), got %d", len(got))
	}
}

func TestHasErrors_WithErrors(t *testing.T) {
	findings := []Finding{
		{Rule: "r1", Severity: SeverityWarning},
		{Rule: "r2", Severity: SeverityError},
		{Rule: "r3", Severity: SeverityInfo},
	}
	if !HasErrors(findings) {
		t.Error("expected HasErrors=true when error-severity finding present")
	}
}

func TestHasErrors_WithoutErrors(t *testing.T) {
	findings := []Finding{
		{Rule: "r1", Severity: SeverityWarning},
		{Rule: "r2", Severity: SeverityInfo},
	}
	if HasErrors(findings) {
		t.Error("expected HasErrors=false when no error-severity findings")
	}
}

func TestHasErrors_Empty(t *testing.T) {
	if HasErrors(nil) {
		t.Error("expected HasErrors=false for nil findings")
	}
	if HasErrors([]Finding{}) {
		t.Error("expected HasErrors=false for empty findings")
	}
}

func TestSeverityRank(t *testing.T) {
	// Verify rank values enforce error < warning < info ordering
	if SeverityError.rank() >= SeverityWarning.rank() {
		t.Error("error rank should be less than warning rank")
	}
	if SeverityWarning.rank() >= SeverityInfo.rank() {
		t.Error("warning rank should be less than info rank")
	}
}
