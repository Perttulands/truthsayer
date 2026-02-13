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
