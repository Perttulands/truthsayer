package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/perttulands/truthsayer/internal/debt"
	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/judge"
	"github.com/perttulands/truthsayer/internal/law"
	"github.com/perttulands/truthsayer/internal/llm"
	"github.com/perttulands/truthsayer/internal/precedent"
	"github.com/perttulands/truthsayer/internal/report"
	"github.com/perttulands/truthsayer/internal/rules"
)

type judgeOptions struct {
	inputPath          string
	format             string
	precedentsPath     string
	debtPath           string
	lawCandidatesPath  string
	lawThreshold       int
	minConfidence      float64
	autoApplyThreshold float64
}

type findingJudge interface {
	JudgeFinding(ctx context.Context, input judge.PromptInput) (judge.Verdict, error)
}

var newFindingJudge = func() (findingJudge, error) {
	apiKey := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	client, err := llm.NewClaudeClient(llm.ClientOptions{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	return judge.NewLLMJudge(client)
}

type judgeOutput struct {
	Version   string              `json:"version"`
	JudgedAt  string              `json:"judged_at"`
	Source    string              `json:"source"`
	Verdicts  []judgeFindingVerdict `json:"verdicts"`
	Summary   judgeSummary        `json:"summary"`
}

type judgeFindingVerdict struct {
	RuleID             string              `json:"rule_id"`
	File               string              `json:"file"`
	Line               int                 `json:"line"`
	Code               string              `json:"code"`
	Context            string              `json:"context,omitempty"`
	Verdict            judge.VerdictType   `json:"verdict"`
	Reasoning          string              `json:"reasoning"`
	Confidence         float64             `json:"confidence"`
	PrecedentDecision  precedent.Decision  `json:"precedent_decision"`
	PrecedentRationale string              `json:"precedent_rationale"`
	Source             string              `json:"source"`
	InputTokens        int                 `json:"input_tokens,omitempty"`
	OutputTokens       int                 `json:"output_tokens,omitempty"`
}

type judgeSummary struct {
	Total            int `json:"total"`
	Guilty           int `json:"guilty"`
	NotGuilty        int `json:"not_guilty"`
	Advisory         int `json:"advisory"`
	AdvisoriesTracked int `json:"advisories_tracked"`
	LawCandidates    int `json:"law_candidates"`
	LLMCalls         int `json:"llm_calls"`
	AutoApplied      int `json:"auto_applied"`
	PrecedentMatches int `json:"precedent_matches"`
}

func runJudge(args []string) int {
	opts, err := parseJudgeOptions(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	findings, err := loadFindingsFromJSON(opts.inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	if opts.precedentsPath == "" {
		opts.precedentsPath = filepath.Join(filepath.Dir(opts.inputPath), precedent.DefaultPath)
	}
	if opts.debtPath == "" {
		opts.debtPath = filepath.Join(filepath.Dir(opts.inputPath), debt.DefaultPath)
	}
	if opts.lawCandidatesPath == "" {
		opts.lawCandidatesPath = filepath.Join(filepath.Dir(opts.inputPath), law.DefaultCandidatesPath)
	}

	store := precedent.NewStore(opts.precedentsPath)
	records, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	debtStore := debt.NewStore(opts.debtPath)

	judger, err := newFindingJudge()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	ruleDescriptions := buildRuleDescriptionMap()
	output := judgeOutput{
		Version:  "1",
		JudgedAt: time.Now().UTC().Format(time.RFC3339),
		Source:   opts.inputPath,
		Verdicts: make([]judgeFindingVerdict, 0, len(findings)),
	}

	for _, f := range findings {
		allMatches := precedent.Match(records, f, precedent.MatchOptions{
			MinConfidence: 0,
			Limit:         10,
		})
		matches := filterMatchesByConfidence(allMatches, opts.minConfidence)
		if len(allMatches) > 0 {
			output.Summary.PrecedentMatches++
		}

		var v judge.Verdict
		if top, ok := strongestMatch(allMatches); ok && effectiveConfidence(top) > opts.autoApplyThreshold {
			v = verdictFromPrecedent(top)
			v.Reasoning = "Auto-applied high-confidence precedent."
			output.Summary.AutoApplied++
		} else {
			judged, err := judger.JudgeFinding(context.Background(), judge.PromptInput{
				Finding:         f,
				RuleDescription: ruleDescriptions[f.Rule],
				Precedents:      matches,
			})
			if err != nil {
				// For command core reliability, if model fails and precedent exists, use the strongest precedent.
				if len(matches) > 0 {
					v = verdictFromPrecedent(matches[0])
				} else {
					fmt.Fprintf(os.Stderr, "error: judge finding %s:%d (%s): %v\n", f.File, f.Line, f.Rule, err)
					return 2
				}
			} else {
				output.Summary.LLMCalls++
				v = judged
			}
		}

		output.Verdicts = append(output.Verdicts, judgeFindingVerdict{
			RuleID:             f.Rule,
			File:               f.File,
			Line:               f.Line,
			Code:               f.Code,
			Context:            f.Context,
			Verdict:            v.Verdict,
			Reasoning:          v.Reasoning,
			Confidence:         v.Confidence,
			PrecedentDecision:  v.PrecedentDecision,
			PrecedentRationale: v.PrecedentRationale,
			Source:             v.Source,
			InputTokens:        v.InputTokens,
			OutputTokens:       v.OutputTokens,
		})
		if v.Verdict == judge.VerdictAdvisory {
			err := debtStore.Add(debt.Entry{
				RuleID:    f.Rule,
				File:      f.File,
				Line:      f.Line,
				Code:      f.Code,
				Message:   f.Message,
				Reasoning: v.Reasoning,
				CreatedAt: time.Now().UTC(),
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: save advisory debt: %v\n", err)
				return 2
			}
			output.Summary.AdvisoriesTracked++
		}

		record := v.ToPrecedent(f, time.Now().UTC())
		if err := record.AsPrecedent().Validate(); err != nil {
			continue
		}
		saved, err := store.AddOrUpdateJudgment(record.AsPrecedent())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: save precedents: %v\n", err)
			return 2
		}
		records = append(records, saved)
	}
	candidates := law.DetectCandidates(records, opts.lawThreshold)
	if len(candidates) > 0 {
		if err := law.NewCandidateStore(opts.lawCandidatesPath).Save(candidates); err != nil {
			fmt.Fprintf(os.Stderr, "error: save law candidates: %v\n", err)
			return 2
		}
		output.Summary.LawCandidates = len(candidates)
	}

	accumulateJudgeSummary(&output)

	if opts.format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(output); err != nil {
			fmt.Fprintf(os.Stderr, "error: write judge output: %v\n", err)
			return 2
		}
	} else {
		fmt.Fprintf(os.Stdout, "Judged %d findings: guilty=%d not_guilty=%d advisory=%d\n",
			output.Summary.Total, output.Summary.Guilty, output.Summary.NotGuilty, output.Summary.Advisory)
	}

	if output.Summary.Guilty > 0 {
		return 1
	}
	return 0
}

func verdictFromPrecedent(p precedent.Precedent) judge.Verdict {
	v := judge.Verdict{
		Reasoning:          "Applied matching precedent due LLM unavailability.",
		Confidence:         p.Confidence,
		PrecedentDecision:  p.Decision,
		PrecedentRationale: p.Rationale,
		Source:             "precedent",
	}
	if v.Confidence <= 0 {
		v.Confidence = 0.5
	}
	if p.Decision == precedent.DecisionAllow {
		v.Verdict = judge.VerdictNotGuilty
	} else {
		v.Verdict = judge.VerdictGuilty
	}
	return v
}

func accumulateJudgeSummary(output *judgeOutput) {
	output.Summary.Total = len(output.Verdicts)
	output.Summary.Guilty = 0
	output.Summary.NotGuilty = 0
	output.Summary.Advisory = 0
	for _, v := range output.Verdicts {
		switch v.Verdict {
		case judge.VerdictGuilty:
			output.Summary.Guilty++
		case judge.VerdictNotGuilty:
			output.Summary.NotGuilty++
		case judge.VerdictAdvisory:
			output.Summary.Advisory++
		}
	}
}

func parseJudgeOptions(args []string) (judgeOptions, error) {
	opts := judgeOptions{
		format:             "json",
		minConfidence:      0.0,
		autoApplyThreshold: 0.9,
		lawThreshold:       10,
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 >= len(args) {
				return judgeOptions{}, fmt.Errorf("--format requires a value")
			}
			opts.format = args[i+1]
			i++
		case "--precedents":
			if i+1 >= len(args) {
				return judgeOptions{}, fmt.Errorf("--precedents requires a value")
			}
			opts.precedentsPath = args[i+1]
			i++
		case "--debt":
			if i+1 >= len(args) {
				return judgeOptions{}, fmt.Errorf("--debt requires a value")
			}
			opts.debtPath = args[i+1]
			i++
		case "--law-candidates":
			if i+1 >= len(args) {
				return judgeOptions{}, fmt.Errorf("--law-candidates requires a value")
			}
			opts.lawCandidatesPath = args[i+1]
			i++
		case "--law-threshold":
			if i+1 >= len(args) {
				return judgeOptions{}, fmt.Errorf("--law-threshold requires a value")
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil {
				return judgeOptions{}, fmt.Errorf("invalid --law-threshold %q", args[i+1])
			}
			if n <= 0 {
				return judgeOptions{}, fmt.Errorf("--law-threshold must be > 0")
			}
			opts.lawThreshold = n
			i++
		case "--min-confidence":
			if i+1 >= len(args) {
				return judgeOptions{}, fmt.Errorf("--min-confidence requires a value")
			}
			n, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return judgeOptions{}, fmt.Errorf("invalid --min-confidence %q", args[i+1])
			}
			if n < 0 || n > 1 {
				return judgeOptions{}, fmt.Errorf("--min-confidence must be between 0 and 1")
			}
			opts.minConfidence = n
			i++
		case "--auto-apply-threshold":
			if i+1 >= len(args) {
				return judgeOptions{}, fmt.Errorf("--auto-apply-threshold requires a value")
			}
			n, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return judgeOptions{}, fmt.Errorf("invalid --auto-apply-threshold %q", args[i+1])
			}
			if n < 0 || n > 1 {
				return judgeOptions{}, fmt.Errorf("--auto-apply-threshold must be between 0 and 1")
			}
			opts.autoApplyThreshold = n
			i++
		default:
			opts.inputPath = args[i]
		}
	}
	if opts.inputPath == "" {
		return judgeOptions{}, fmt.Errorf("judge requires a findings JSON path")
	}
	if opts.format != "json" && opts.format != "text" {
		return judgeOptions{}, fmt.Errorf("unknown format %q (use json or text)", opts.format)
	}
	return opts, nil
}

func loadFindingsFromJSON(path string) ([]finding.Finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read findings: %w", err)
	}
	var rpt report.JSONReport
	if err := json.Unmarshal(data, &rpt); err != nil {
		return nil, fmt.Errorf("decode findings json: %w", err)
	}
	findings := make([]finding.Finding, 0, len(rpt.Findings))
	for _, f := range rpt.Findings {
		findings = append(findings, finding.Finding{
			Rule:       f.Rule,
			Severity:   finding.Severity(f.Severity),
			File:       f.File,
			Line:       f.Line,
			Code:       f.Code,
			Context:    f.Context,
			Message:    f.Message,
			Suggestion: f.Suggestion,
		})
	}
	return findings, nil
}

func buildRuleDescriptionMap() map[string]string {
	reg := rules.DefaultRegistry()
	all := reg.AllRules()
	out := make(map[string]string, len(all))
	for _, r := range all {
		out[r.ID] = r.Description
	}
	return out
}

func filterMatchesByConfidence(matches []precedent.Precedent, min float64) []precedent.Precedent {
	if len(matches) == 0 || min <= 0 {
		return matches
	}
	out := make([]precedent.Precedent, 0, len(matches))
	for _, p := range matches {
		if effectiveConfidence(p) >= min {
			out = append(out, p)
		}
	}
	return out
}

func strongestMatch(matches []precedent.Precedent) (precedent.Precedent, bool) {
	if len(matches) == 0 {
		return precedent.Precedent{}, false
	}
	return matches[0], true
}

func effectiveConfidence(p precedent.Precedent) float64 {
	if p.Confidence <= 0 {
		return 0.5
	}
	return p.Confidence
}
