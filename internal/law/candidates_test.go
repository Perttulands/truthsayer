package law

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/precedent"
)

func TestDetectCandidates(t *testing.T) {
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	records := make([]precedent.Precedent, 0, 4)
	for i := 0; i < 4; i++ {
		records = append(records, precedent.Precedent{
			RuleID:        "silent-fallback.hidden-failure-bash",
			ViolationHash: "vh",
			PatternHash:   "ph",
			Decision:      precedent.DecisionAllow,
			Rationale:     "known cleanup",
			CreatedAt:     now.Add(time.Duration(i) * time.Hour),
		})
	}

	candidates := DetectCandidates(records, 3)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].Count != 4 {
		t.Fatalf("expected count 4, got %d", candidates[0].Count)
	}
}

func TestCandidateStore_SaveLoad(t *testing.T) {
	store := NewCandidateStore(filepath.Join(t.TempDir(), DefaultCandidatesPath))
	c := Candidate{
		RuleID:      "r",
		PatternHash: "p",
		Decision:    precedent.DecisionDeny,
		Count:       10,
		Threshold:   10,
		FirstSeen:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		LastSeen:    time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
	}
	if err := store.Save([]Candidate{c}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(got))
	}
	if got[0].RuleID != c.RuleID {
		t.Fatalf("unexpected candidate: %+v", got[0])
	}
}
