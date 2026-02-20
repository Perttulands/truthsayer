package law

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perttulands/truthsayer/internal/precedent"
)

// DefaultProposalsPath is where Senate proposal markdown is written.
const DefaultProposalsPath = "law-updates.md"

// SuggestAmendment returns a concise amendment direction based on consistent decision.
func SuggestAmendment(decision precedent.Decision) string {
	if decision == precedent.DecisionAllow {
		return "Add a targeted exception for this pattern (pattern consistently judged not-guilty)."
	}
	return "Tighten enforcement for this pattern (pattern consistently judged guilty)."
}

// RenderProposalsMarkdown renders law update proposals in Senate-review format.
func RenderProposalsMarkdown(candidates []Candidate, ruleDescriptions map[string]string, at time.Time) string {
	var b strings.Builder
	if at.IsZero() {
		at = time.Now().UTC()
	}
	b.WriteString("# Law Update Proposals\n\n")
	b.WriteString(fmt.Sprintf("_Generated: %s_\n\n", at.UTC().Format(time.RFC3339)))
	if len(candidates) == 0 {
		b.WriteString("No proposal candidates.\n")
		return b.String()
	}

	for i, c := range candidates {
		desc := strings.TrimSpace(ruleDescriptions[c.RuleID])
		if desc == "" {
			desc = "(rule description unavailable)"
		}
		b.WriteString(fmt.Sprintf("## Proposal %d: `%s`\n\n", i+1, c.RuleID))
		b.WriteString(fmt.Sprintf("- Current rule: %s\n", desc))
		b.WriteString(fmt.Sprintf("- Pattern hash: `%s`\n", c.PatternHash))
		b.WriteString(fmt.Sprintf("- Consistent decision: `%s`\n", c.Decision))
		b.WriteString(fmt.Sprintf("- Evidence count: %d (threshold %d)\n", c.Count, c.Threshold))
		b.WriteString(fmt.Sprintf("- First seen: %s\n", c.FirstSeen.UTC().Format(time.RFC3339)))
		b.WriteString(fmt.Sprintf("- Last seen: %s\n", c.LastSeen.UTC().Format(time.RFC3339)))
		b.WriteString(fmt.Sprintf("- Suggested amendment: %s\n\n", SuggestAmendment(c.Decision)))
	}
	return b.String()
}

// WriteProposals writes proposals to markdown file.
func WriteProposals(path string, candidates []Candidate, ruleDescriptions map[string]string, at time.Time) error {
	content := RenderProposalsMarkdown(candidates, ruleDescriptions, at)
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("law: create dir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("law: write proposals %s: %w", path, err)
	}
	return nil
}

