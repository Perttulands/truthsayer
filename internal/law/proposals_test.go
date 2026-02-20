package law

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/precedent"
)

func TestRenderProposalsMarkdown(t *testing.T) {
	md := RenderProposalsMarkdown([]Candidate{
		{
			RuleID:      "silent-fallback.hidden-failure-bash",
			PatternHash: "abc123",
			Decision:    precedent.DecisionAllow,
			Count:       12,
			Threshold:   10,
			FirstSeen:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			LastSeen:    time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
		},
	}, map[string]string{
		"silent-fallback.hidden-failure-bash": "Command failure silently suppressed",
	}, time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC))

	if !strings.Contains(md, "Law Update Proposals") {
		t.Fatalf("missing title:\n%s", md)
	}
	if !strings.Contains(md, "Suggested amendment") {
		t.Fatalf("missing amendment guidance:\n%s", md)
	}
}

func TestWriteProposals(t *testing.T) {
	path := filepath.Join(t.TempDir(), DefaultProposalsPath)
	err := WriteProposals(path, []Candidate{}, map[string]string{}, time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("write proposals failed: %v", err)
	}
}
