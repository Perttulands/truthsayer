# Changelog

All notable changes to Truthsayer.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)

## [Unreleased]

### Changed
- README: restored mythology intro (The Law Keeper), character sigil and visual items, "Part of the Agora" section

### Added
- Pre-commit judgment integration: `truthsayer hook` now executes `scan -> judge`, blocks only on guilty verdicts, and persists precedents from hook judgments
- Advisory tracking with `.truthsayer-debt.json`: advisory verdicts now create debt entries, and `truthsayer debt` lists accumulated advisory items
- Consistent-ruling detection in `truthsayer judge`: repeated same-pattern/same-decision judgments are flagged as law candidates (default threshold 10, configurable) and logged to `.truthsayer-law-candidates.json`
- Automatic law update proposal generation: `truthsayer judge` now writes Senate-ready `law-updates.md` proposals with rule context, pattern evidence, and suggested amendments
- Senate verdict parsing: new `truthsayer senate parse <file>` command validates verdict schema (including amendment actions) from JSON or fenced JSON markdown
- Senate amendment application: new `truthsayer senate apply <file> [repo]` persists approved amendments + audit trail, and scan engine now enforces applied `set_severity`/`disable_rule`/`enable_rule` amendments
- Judgment cost tracking: `truthsayer judge` now logs token/cost metrics to `.truthsayer-cost.jsonl`, reports spend in summary output, and supports `--budget` caps with precedent fallback on exhaustion
- Similar-finding batching in judgment pipeline: findings are grouped by `rule + pattern` so one LLM decision can fan out to matching findings, reducing LLM call volume and surfacing `batches` in summary
- Warmup mode: new `truthsayer warmup <repo>` command runs full-repo `scan -> judge` to build precedent history and reports warmup statistics
- Integration coverage for judgment flow: added end-to-end CLI integration tests for `scan -> judge` precedent persistence and precedent fallback behavior when LLM judgment fails
- `finding.Finding.Context` with ±10-line source windows and highlighted violation line, populated across Go, JS/TS, Python, bash, and config scans
- Pattern hashing for judgments via `precedent.HashPattern`/`precedent.HashFindingPattern`, with normalization of variable names, literals, and whitespace for stable precedent matching
- Precedent lookup API via `Store.Match`/`precedent.Match` using `rule_id + pattern_hash`, confidence thresholds, and confidence-first sorting for ranked match retrieval
- Claude LLM client (`internal/llm`) with Anthropic Messages API integration, auth validation, retry/backoff for 429/5xx responses, and request pacing for rate limiting
- Judgment prompt template builder (`internal/judge.BuildPrompt`) with structured finding/context/precedent payload and strict parseable JSON response schema
- LLM judgment call logic (`internal/judge.LLMJudge`) that builds prompts, calls the LLM client, parses typed verdict JSON, and converts verdicts into precedent records
- New `truthsayer judge` command core: reads findings JSON, retrieves precedent context, performs judgment, outputs verdicts JSON/text, and writes updated precedent records
- Confidence scoring updates in precedent storage: repeated matching decisions increase confidence and seen-count, while overrides decay confidence and reset seen-count
- High-confidence auto-apply in `truthsayer judge`: precedent matches above threshold (default `>0.9`) bypass LLM calls and apply cached decisions directly with `auto_applied` metrics

### Changed
- `hidden-failure-bash` rule upgraded to ERROR severity; downgraded to INFO when line has `# REASON:` comment justifying the suppression
- `silent-fallback.hidden-failure-bash` now exempts `|| true` inside `trap` handlers, one-hop trap-invoked cleanup functions, and functions marked with `# truthsayer:cleanup-context`

## [1.0.0] - 2026-02-13

### Added
- 24 detection rules across Go AST, bash regex, and general regex categories
- **Go AST rules**: ignored-error, unwrapped-error, generic-message, no-timeout, unvalidated-env-go, long-function-no-log, error-path-no-log, mock-import-non-test, bare-return-on-error, http-200-on-error, nil-on-error, no-request-id, magic-number, debug-guard
- **Bash regex rules**: hidden-failure-bash, hardcoded-path, unvalidated-env-bash, no-err-trap, no-stderr-capture
- **General regex rules**: secret-in-config, test-fixture-ref, missing-gitignore
- CLI commands: `scan`, `check`, `report`, `watch`, `rules`, `hook`, `ci`, `doctor`, `--version`
- TOML configuration for enabling/disabling rules, severity overrides, and file exclusions
- Watch mode with diff-aware filtering (US-006, US-007)
- JSON report output (US-008)
- CI quality gate with exit code 1 on errors (US-009)
- Human-readable terminal summary (US-010)
- Git pre-commit hook generation (US-016)
- GitHub Actions workflow generation via `ci init` (US-017)
- Doctor command for installation health checks (US-018)
- `--version` flag with ldflags support (US-019)
- README and LICENSE

### Fixed
- Zero errors and zero warnings on self-scan
