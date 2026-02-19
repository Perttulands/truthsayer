# Truthsayer Judgment System

## Architecture

```
truthsayer scan . → findings.json → truthsayer judge → verdicts.json
                                          ↑
                                    precedents.json
                                          ↓
                                    law-updates.md (proposals)
```

## Components

### 1. `truthsayer scan --format json .`
Existing. Produces raw findings with file, line, rule, severity, snippet.

### 2. `truthsayer judge` (NEW)
Takes findings.json + source context. For each finding:
- Reads the surrounding code (±10 lines)
- Checks precedents.json for matching patterns
- If precedent exists with high confidence → apply automatically
- If no precedent → call LLM with finding + context + the rule description
- LLM returns: verdict (guilty/not-guilty/advisory), reasoning, suggested precedent

### 3. `precedents.json` (NEW)
Accumulated judgments. Schema:
```json
{
  "rule": "silent-fallback.hidden-failure-bash",
  "pattern": "|| true in trap/cleanup context",
  "verdict": "not-guilty",
  "reasoning": "Error suppression in cleanup/trap handlers is intentional defensive coding",
  "confidence": 0.95,
  "seen_count": 47,
  "first_seen": "2026-02-19",
  "last_seen": "2026-02-19",
  "repos": ["argus", "learning-loop"]
}
```

### 4. Law Updates
When the judge consistently rules the same way on a pattern, it proposes a rule update:
- "Rule X produces false positives in context Y → refine rule"
- "Pattern Z is always guilty → promote to error severity"
- Written to `law-updates.md` for human review

## Verdicts

| Verdict | Meaning | Action |
|---------|---------|--------|
| **guilty** | Real issue, must fix | Blocks commit |
| **not-guilty** | False positive or intentional | Passes, creates precedent |
| **advisory** | Real but minor, fix when convenient | Passes, tracked as tech debt |

## Pre-commit Flow

```
git commit →
  truthsayer hook . →
    scan staged files →
    check findings against precedents →
    any unprecedented findings? →
      yes → judge (LLM) → verdict →
        guilty? → block commit
        not-guilty? → record precedent, pass
        advisory? → record debt, pass
      no → all precedented → apply cached verdicts → pass/block
```

## Precedent Lifecycle

1. First encounter: LLM judges, creates precedent with confidence 0.5
2. Same pattern seen again: confidence increases
3. Confidence > 0.9: precedent is "established" — no LLM call needed
4. If a human overrides a precedent: confidence resets, LLM re-judges
5. Precedents with 0 hits in 90 days: archived

## Cost Control

- Established precedents (confidence > 0.9) skip LLM entirely
- Batch similar findings into one LLM call
- Use haiku/fast model for judgment (not opus)
- Cache by rule+pattern hash
- Target: <$0.01 per commit after warmup period
