package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/perttulands/truthsayer/internal/judge"
	"github.com/perttulands/truthsayer/internal/precedent"
)

func TestWarmup_BuildsPrecedentsAndPrintsStats(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.go", goWithError)

	fake := &fakeFindingJudge{
		verdict: judge.Verdict{
			Verdict:            judge.VerdictNotGuilty,
			Reasoning:          "warmup allowed for known pattern",
			Confidence:         0.85,
			PrecedentDecision:  precedent.DecisionAllow,
			PrecedentRationale: "accepted during warmup",
			Source:             "llm",
		},
	}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	out := captureStdout(t, func() {
		code := runWarmup([]string{dir})
		if code != 0 {
			t.Fatalf("expected warmup exit code 0, got %d", code)
		}
	})
	if !strings.Contains(out, "Warmup complete:") {
		t.Fatalf("expected warmup summary output, got:\n%s", out)
	}
	if _, err := precedent.NewStore(filepath.Join(dir, precedent.DefaultPath)).Load(); err != nil {
		t.Fatalf("load precedents failed: %v", err)
	}
}

func TestWarmup_NoArgs(t *testing.T) {
	code := runWarmup(nil)
	if code != 2 {
		t.Fatalf("expected exit code 2 for missing args, got %d", code)
	}
}
