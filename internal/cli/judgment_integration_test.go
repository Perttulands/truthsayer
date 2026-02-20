package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/judge"
	"github.com/perttulands/truthsayer/internal/precedent"
)

func TestIntegration_ScanToJudge_PersistsPrecedents(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.go", goWithError)

	scanOut := captureStdout(t, func() {
		code := runScan([]string{"--format", "json", dir})
		if code != 1 {
			t.Fatalf("expected scan exit code 1, got %d", code)
		}
	})
	findingsPath := filepath.Join(dir, "findings.json")
	if err := os.WriteFile(findingsPath, []byte(scanOut), 0o644); err != nil {
		t.Fatalf("write findings json: %v", err)
	}

	fake := &fakeFindingJudge{
		verdict: judge.Verdict{
			Verdict:            judge.VerdictNotGuilty,
			Reasoning:          "integration approval",
			Confidence:         0.9,
			PrecedentDecision:  precedent.DecisionAllow,
			PrecedentRationale: "safe in this context",
			Source:             "llm",
			InputTokens:        120,
			OutputTokens:       40,
		},
	}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	code := runJudge([]string{findingsPath})
	if code != 0 {
		t.Fatalf("expected judge exit code 0, got %d", code)
	}

	records, err := precedent.NewStore(filepath.Join(dir, precedent.DefaultPath)).Load()
	if err != nil {
		t.Fatalf("load precedents: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("expected precedents saved by integration pipeline")
	}
}

func TestIntegration_JudgeFallsBackToPrecedentWhenLLMUnavailable(t *testing.T) {
	dir := t.TempDir()
	findings := []finding.Finding{
		{
			Rule:     "silent-fallback.hidden-failure-bash",
			Severity: finding.SeverityError,
			File:     "script.sh",
			Line:     3,
			Code:     "cmd || true",
			Message:  "hidden failure",
		},
	}
	findingsPath := writeFindingsJSON(t, dir, findings)

	seed := precedent.Precedent{
		RuleID:        findings[0].Rule,
		ViolationHash: "vh",
		PatternHash:   precedent.HashFindingPattern(findings[0]),
		Decision:      precedent.DecisionAllow,
		Rationale:     "already approved pattern",
		Confidence:    0.8,
		SeenCount:     4,
		CreatedAt:     time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	}
	if err := precedent.NewStore(filepath.Join(dir, precedent.DefaultPath)).Save([]precedent.Precedent{seed}); err != nil {
		t.Fatalf("seed precedent: %v", err)
	}

	fake := &fakeFindingJudge{err: errors.New("llm unavailable")}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	out := captureStdout(t, func() {
		code := runJudge([]string{"--auto-apply-threshold", "1", findingsPath})
		if code != 0 {
			t.Fatalf("expected judge exit code 0 via precedent fallback, got %d", code)
		}
	})
	if !strings.Contains(out, `"source": "precedent"`) {
		t.Fatalf("expected precedent source in output, got:\n%s", out)
	}
}
