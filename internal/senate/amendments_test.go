package senate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAmendmentStore_ApplyVerdict(t *testing.T) {
	store := NewAmendmentStore(filepath.Join(t.TempDir(), DefaultAmendmentsPath))
	verdict := Verdict{
		ID:     "quick-1",
		Status: StatusApproved,
		Amendments: []Amendment{
			{RuleID: "silent-fallback.empty-error-check", Action: ActionSetSeverity, Severity: "warning"},
		},
	}
	added, err := store.ApplyVerdict(verdict, time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("apply verdict failed: %v", err)
	}
	if len(added) != 1 {
		t.Fatalf("expected 1 added amendment, got %d", len(added))
	}
}

func TestAppendAudit(t *testing.T) {
	path := filepath.Join(t.TempDir(), DefaultAmendmentsAuditPath)
	err := AppendAudit(path, []AppliedAmendment{
		{
			VerdictID: "quick-1",
			AppliedAt: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
			Amendment: Amendment{RuleID: "r", Action: ActionDisableRule},
		},
	})
	if err != nil {
		t.Fatalf("append audit failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected audit content")
	}
}
