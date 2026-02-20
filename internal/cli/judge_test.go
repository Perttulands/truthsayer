package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/judge"
	"github.com/perttulands/truthsayer/internal/precedent"
	"github.com/perttulands/truthsayer/internal/report"
)

type fakeFindingJudge struct {
	verdict judge.Verdict
	err     error
	calls   int
}

func (f *fakeFindingJudge) JudgeFinding(context.Context, judge.PromptInput) (judge.Verdict, error) {
	f.calls++
	if f.err != nil {
		return judge.Verdict{}, f.err
	}
	return f.verdict, nil
}

func writeFindingsJSON(t *testing.T, dir string, findings []finding.Finding) string {
	t.Helper()
	path := filepath.Join(dir, "findings.json")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create findings json: %v", err)
	}
	defer f.Close()
	if err := report.JSON(f, findings, dir, time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC), 1, 1); err != nil {
		t.Fatalf("write report json: %v", err)
	}
	return path
}

func TestRunJudge_JSONOutputAndPrecedentWrite(t *testing.T) {
	dir := t.TempDir()
	findingsPath := writeFindingsJSON(t, dir, []finding.Finding{
		{
			Rule:    "silent-fallback.hidden-failure-bash",
			Severity: finding.SeverityError,
			File:    "script.sh",
			Line:    2,
			Code:    "cmd || true",
			Context: ">> 2 | cmd || true",
			Message: "hidden failure",
		},
	})

	fake := &fakeFindingJudge{
		verdict: judge.Verdict{
			Verdict:            judge.VerdictNotGuilty,
			Reasoning:          "intentional cleanup trap",
			Confidence:         0.91,
			PrecedentDecision:  precedent.DecisionAllow,
			PrecedentRationale: "trap cleanup context",
			Source:             "llm",
		},
	}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	var out judgeOutput
	stdout := captureStdout(t, func() {
		code := runJudge([]string{findingsPath})
		if code != 0 {
			t.Fatalf("expected exit code 0, got %d", code)
		}
	})
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("decode judge output: %v\n%s", err, stdout)
	}
	if out.Summary.Total != 1 || out.Summary.NotGuilty != 1 {
		t.Fatalf("unexpected summary: %+v", out.Summary)
	}
	if out.Summary.LLMCalls != 1 {
		t.Fatalf("expected one llm call, got %d", out.Summary.LLMCalls)
	}

	precedentsPath := filepath.Join(dir, precedent.DefaultPath)
	records, err := precedent.NewStore(precedentsPath).Load()
	if err != nil {
		t.Fatalf("load precedents: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("expected saved precedent records")
	}
}

func TestRunJudge_ExitCode1OnGuilty(t *testing.T) {
	dir := t.TempDir()
	findingsPath := writeFindingsJSON(t, dir, []finding.Finding{
		{
			Rule:    "silent-fallback.empty-error-check",
			Severity: finding.SeverityError,
			File:    "handler.go",
			Line:    8,
			Code:    "if err != nil { return nil }",
			Message: "swallowed error",
		},
	})

	fake := &fakeFindingJudge{
		verdict: judge.Verdict{
			Verdict:            judge.VerdictGuilty,
			Reasoning:          "error suppression is harmful",
			Confidence:         0.85,
			PrecedentDecision:  precedent.DecisionDeny,
			PrecedentRationale: "must return or wrap error",
			Source:             "llm",
		},
	}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	code := runJudge([]string{findingsPath})
	if code != 1 {
		t.Fatalf("expected exit code 1 for guilty verdict, got %d", code)
	}
}

func TestRunJudge_FallbackToPrecedentWhenJudgeFails(t *testing.T) {
	dir := t.TempDir()
	findings := []finding.Finding{
		{
			Rule:    "silent-fallback.hidden-failure-bash",
			Severity: finding.SeverityError,
			File:    "script.sh",
			Line:    3,
			Code:    "cmd || true",
			Message: "hidden failure",
		},
	}
	findingsPath := writeFindingsJSON(t, dir, findings)

	p := precedent.Precedent{
		RuleID:        findings[0].Rule,
		ViolationHash: "vh",
		PatternHash:   precedent.HashFindingPattern(findings[0]),
		Decision:      precedent.DecisionAllow,
		Rationale:     "known cleanup context",
		Confidence:    0.95,
		SeenCount:     5,
		CreatedAt:     time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	}
	if err := precedent.NewStore(filepath.Join(dir, precedent.DefaultPath)).Save([]precedent.Precedent{p}); err != nil {
		t.Fatalf("save precedent seed: %v", err)
	}

	fake := &fakeFindingJudge{err: errors.New("llm down")}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	stdout := captureStdout(t, func() {
		code := runJudge([]string{findingsPath})
		if code != 0 {
			t.Fatalf("expected exit code 0 via allow precedent fallback, got %d", code)
		}
	})
	if !strings.Contains(stdout, `"source": "precedent"`) {
		t.Fatalf("expected precedent fallback source in output, got:\n%s", stdout)
	}
}

func TestParseJudgeOptions(t *testing.T) {
	opts, err := parseJudgeOptions([]string{"--precedents", "p.json", "--debt", "d.json", "--law-candidates", "c.json", "--law-threshold", "7", "--min-confidence", "0.9", "--auto-apply-threshold", "0.95", "f.json"})
	if err != nil {
		t.Fatalf("parseJudgeOptions failed: %v", err)
	}
	if opts.precedentsPath != "p.json" {
		t.Fatalf("unexpected precedents path: %q", opts.precedentsPath)
	}
	if opts.debtPath != "d.json" {
		t.Fatalf("unexpected debt path: %q", opts.debtPath)
	}
	if opts.lawCandidatesPath != "c.json" {
		t.Fatalf("unexpected law candidates path: %q", opts.lawCandidatesPath)
	}
	if opts.lawThreshold != 7 {
		t.Fatalf("unexpected law threshold: %d", opts.lawThreshold)
	}
	if opts.minConfidence != 0.9 {
		t.Fatalf("unexpected min confidence: %f", opts.minConfidence)
	}
	if opts.inputPath != "f.json" {
		t.Fatalf("unexpected input path: %q", opts.inputPath)
	}
	if opts.autoApplyThreshold != 0.95 {
		t.Fatalf("unexpected auto-apply threshold: %f", opts.autoApplyThreshold)
	}
}

func TestRunJudge_AutoApplyHighConfidenceSkipsLLM(t *testing.T) {
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

	p := precedent.Precedent{
		RuleID:        findings[0].Rule,
		ViolationHash: "vh",
		PatternHash:   precedent.HashFindingPattern(findings[0]),
		Decision:      precedent.DecisionAllow,
		Rationale:     "known cleanup context",
		Confidence:    0.95,
		SeenCount:     9,
		CreatedAt:     time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	}
	if err := precedent.NewStore(filepath.Join(dir, precedent.DefaultPath)).Save([]precedent.Precedent{p}); err != nil {
		t.Fatalf("save precedent seed: %v", err)
	}

	fake := &fakeFindingJudge{
		verdict: judge.Verdict{
			Verdict:            judge.VerdictGuilty,
			Reasoning:          "should not be used",
			Confidence:         0.1,
			PrecedentDecision:  precedent.DecisionDeny,
			PrecedentRationale: "should not be used",
		},
	}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	stdout := captureStdout(t, func() {
		code := runJudge([]string{findingsPath})
		if code != 0 {
			t.Fatalf("expected exit code 0 from auto-applied allow precedent, got %d", code)
		}
	})
	if fake.calls != 0 {
		t.Fatalf("expected zero llm calls due auto-apply, got %d", fake.calls)
	}
	if !strings.Contains(stdout, `"auto_applied": 1`) {
		t.Fatalf("expected auto_applied summary count, got:\n%s", stdout)
	}
}

func TestRunJudge_AdvisoryWritesDebtEntry(t *testing.T) {
	dir := t.TempDir()
	findingsPath := writeFindingsJSON(t, dir, []finding.Finding{
		{
			Rule:     "trace-gaps.long-function-no-log",
			Severity: finding.SeverityWarning,
			File:     "service.go",
			Line:     22,
			Code:     "func process(){ ... }",
			Message:  "function has no log statements",
		},
	})

	fake := &fakeFindingJudge{
		verdict: judge.Verdict{
			Verdict:            judge.VerdictAdvisory,
			Reasoning:          "low risk now, but should be improved",
			Confidence:         0.8,
			PrecedentDecision:  precedent.DecisionDeny,
			PrecedentRationale: "tracking for future rule update",
			Source:             "llm",
		},
	}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	stdout := captureStdout(t, func() {
		code := runJudge([]string{findingsPath})
		if code != 0 {
			t.Fatalf("expected exit code 0 for advisory-only run, got %d", code)
		}
	})

	if !strings.Contains(stdout, `"advisories_tracked": 1`) {
		t.Fatalf("expected advisories_tracked summary, got:\n%s", stdout)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".truthsayer-debt.json"))
	if err != nil {
		t.Fatalf("read debt file: %v", err)
	}
	if !strings.Contains(string(raw), `"rule_id": "trace-gaps.long-function-no-log"`) {
		t.Fatalf("expected advisory debt entry, got:\n%s", string(raw))
	}
}

func TestRunJudge_LogsLawCandidateWhenThresholdReached(t *testing.T) {
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

	seed := make([]precedent.Precedent, 0, 2)
	for i := 0; i < 2; i++ {
		seed = append(seed, precedent.Precedent{
			RuleID:        findings[0].Rule,
			ViolationHash: "vh",
			PatternHash:   precedent.HashFindingPattern(findings[0]),
			Decision:      precedent.DecisionAllow,
			Rationale:     "known cleanup",
			Confidence:    0.8,
			SeenCount:     2 + i,
			CreatedAt:     time.Date(2026, 2, 20, i, 0, 0, 0, time.UTC),
		})
	}
	if err := precedent.NewStore(filepath.Join(dir, precedent.DefaultPath)).Save(seed); err != nil {
		t.Fatalf("save seed precedents: %v", err)
	}

	fake := &fakeFindingJudge{
		verdict: judge.Verdict{
			Verdict:            judge.VerdictNotGuilty,
			Reasoning:          "approved cleanup context",
			Confidence:         0.9,
			PrecedentDecision:  precedent.DecisionAllow,
			PrecedentRationale: "stable allow decision",
			Source:             "llm",
		},
	}
	oldFactory := newFindingJudge
	newFindingJudge = func() (findingJudge, error) { return fake, nil }
	defer func() { newFindingJudge = oldFactory }()

	stdout := captureStdout(t, func() {
		code := runJudge([]string{"--law-threshold", "3", findingsPath})
		if code != 0 {
			t.Fatalf("expected exit code 0, got %d", code)
		}
	})
	if !strings.Contains(stdout, `"law_candidates": 1`) {
		t.Fatalf("expected law candidate summary, got:\n%s", stdout)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".truthsayer-law-candidates.json"))
	if err != nil {
		t.Fatalf("read law candidates file: %v", err)
	}
	if !strings.Contains(string(raw), `"rule_id": "silent-fallback.hidden-failure-bash"`) {
		t.Fatalf("expected law candidate entry, got:\n%s", string(raw))
	}
}
