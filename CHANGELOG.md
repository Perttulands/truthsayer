# Changelog

All notable changes to Truthsayer.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)

## [Unreleased]

### Added
- `finding.Finding.Context` with ±10-line source windows and highlighted violation line, populated across Go, JS/TS, Python, bash, and config scans
- Pattern hashing for judgments via `precedent.HashPattern`/`precedent.HashFindingPattern`, with normalization of variable names, literals, and whitespace for stable precedent matching
- Precedent lookup API via `Store.Match`/`precedent.Match` using `rule_id + pattern_hash`, confidence thresholds, and confidence-first sorting for ranked match retrieval

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
