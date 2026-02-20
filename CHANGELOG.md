# Changelog

All notable changes to Truthsayer.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)

## [Unreleased]

### Added
- Judgment system design and prototype `scripts/judge.sh`: LLM layer that reads scan findings from stdin, checks against a precedents file, calls an LLM for unprecedented findings, and outputs verdicts
- `state/precedents.json` for storing judgment precedents across runs
- `.truthsayer.toml` project config: exclude test files from scan, suppress self-referential false positives
- Banner image (`banner.jpg`) for README

### Changed
- README: mythology-forward rewrite — Athena's Forge → The Agora framing, spicy standalone voice
- README: "For Agents" section added covering install, what-this-is, and runtime usage for agent consumers
- AGENTS.md: changelog ground rule added (agents must update CHANGELOG on every significant commit)
- `bd` (beads) replaces `br` (beads_rust) as the agent workflow CLI throughout AGENTS.md

### Fixed
- 2026-02-20: strengthened CLI error handling for scan-reported swallowed errors in `internal/cli/doctor.go` and `internal/cli/judge.go` by making doctor's fallback explicit and logging judge LLM failures before precedent fallback, so operational failures are visible instead of silent.

## [2.0.0] - 2026-02-16

Multi-language support: JS/TS and Python rule coverage added via tree-sitter AST parsing.

### Added
- **Tree-sitter infrastructure** (Sprint 1): shared utilities, `JSScanner` and `PyScanner` with parser pooling, `JSASTChecker`/`PyASTChecker` interfaces, engine routing and walker extended for JS/TS/Python, parser correctness and edge-case tests
- **28 JS/TS rules** (Sprint 2):
  - *silent-fallback*: `empty-catch`, `catch-return-null`, `floating-promise`, `callback-err-ignored`, `optional-chain-silence`
  - *error-context*: `rethrow-no-wrap`, `generic-error-message`, `promise-reject-non-error`, `console-error-no-throw`, `http-200-on-error`
  - *trace-gaps + mock-leakage*: `no-error-handler-express`, `test-import-in-src`, `env-test-check`, `missing-correlation-id`
  - *bad-defaults*: `no-timeout-fetch`, `any-type-assertion`, `non-null-assertion`, `eval-usage`
  - *test-isolation*: `no-afterall-cleanup`, `test-only-import`
  - *regex*: `jest-mock-in-src`, `storybook-in-src`, `ts-ignore`, `eslint-disable-no-reason`, `no-strict-mode`, `no-unhandled-rejection`, `console-log-in-production`, `hardcoded-api-url`, `dotenv-no-example`
  - JS/TS integration test suite (US-208)
- **26 Python rules** (Sprint 3):
  - *silent-fallback*: `bare-except`, `except-pass`, `except-broad`, `subprocess-no-check`, `getattr-silent-default`, `dict-get-none`
  - *error-context*: `raise-from-none`, `bare-raise-different`, `generic-exception`, `string-exception`, `log-and-raise`
  - *bad-defaults*: `mutable-default-arg`, `no-timeout-requests`, `star-import`, `global-state`, `no-encoding-open`
  - *mock-leakage + trace-gaps*: `unittest-import`, `debug-flag`, `silent-request`
  - *regex*: `print-debug`, `no-logging-config`, `pytest-fixture-in-src`, `type-ignore-bare`, `noqa-bare`, `hardcoded-credentials-py`, `requirements-unpinned`
  - Python integration test suite (US-306)
- **Cross-language integration** (Sprint 4):
  - Cross-language integration tests (US-403)
  - Doctor command extended to report multi-language parser status (US-402)
  - CLI `--lang` flag for `scan` and `rules` commands to filter by language (US-401)
  - Config extensions for per-language enable/disable in `.truthsayer.toml` (US-400)
  - Performance benchmarks for multi-language scan (US-404)
  - Shared language resolution extracted to `lang.go` (US-405a)
  - Documentation and CI updates (US-405)

## [1.1.0] - 2026-02-15

Post-launch improvements: new rules, false positive reduction, CLI ergonomics, performance.

### Added
- `--create-beads` flag: automatically creates a problem bead for each scan finding, wiring Truthsayer into the agent workflow
- **Test-isolation rules**: detect leaked file handles, unclosed resources, and test-only state that escapes test scope
- **Security and code-quality rules** (via `security_regex_rules.go`): `code-quality.unused-variable`, `code-quality.unreachable-code`, `security.sql-injection`, and related fixtures
- `--parallel` flag for scan: controls worker pool size for large repos
- mtime-based scan cache: skip files unchanged since last scan, significantly faster on re-runs
- 3 new Go AST rules: `goroutine-no-context`, `defer-in-loop`, `error-string-compare`

### Changed
- `hidden-failure-bash`: upgraded from WARNING to ERROR severity; downgraded back to INFO when the line includes a `# REASON:` comment explaining the suppression
- Reduced false positives across `magic-number`, `long-function-no-log`, `unvalidated-env-go`, and `empty-error-check` rules
- CLI output improved: cleaner summary format, better grouping

### Fixed
- Zero errors and zero warnings on self-scan (production code)
- `ignored-error` rule: skip test files to avoid false positives in test helpers
- Duplicate `main.go` removed (reverted erroneous removal, then correctly resolved)
- License holder name corrected

## [1.0.0] - 2026-02-13

Initial release.

### Added
- **24 detection rules** across Go AST, bash regex, and general regex categories
- **Go AST rules**: `ignored-error`, `unwrapped-error`, `generic-message`, `no-timeout`, `unvalidated-env-go`, `long-function-no-log`, `error-path-no-log`, `mock-import-non-test`, `bare-return-on-error`, `http-200-on-error`, `nil-on-error`, `no-request-id`, `magic-number`, `debug-guard`
- **Bash regex rules**: `hidden-failure-bash`, `hardcoded-path`, `unvalidated-env-bash`, `no-err-trap`, `no-stderr-capture`
- **General regex rules**: `secret-in-config`, `test-fixture-ref`, `missing-gitignore`
- CLI commands: `scan`, `check`, `report`, `watch`, `rules`, `hook`, `ci`, `doctor`, `--version`
- TOML configuration for enabling/disabling rules, severity overrides, and file exclusions (US-011–US-013)
- Scan results sorted by severity: error > warning > info (US-004)
- Findings include file path, line number, code snippet, and suggestion (US-003)
- Automatic skip of `vendor/`, `node_modules/`, `.git/` directories (US-005)
- Single-file scanning via `check` command (US-002)
- Watch mode for directory with instant feedback on file changes (US-006)
- Diff-aware filtering in watch mode: report only changed lines (US-007)
- JSON report output via `report` command (US-008)
- CI quality gate: exit code 1 when error-severity findings found (US-009)
- Human-readable terminal summary with severity and category counts (US-010)
- `rules` command: list all available rules with IDs, descriptions, severity (US-014)
- `rules --enabled` flag: list only currently active rules (US-015)
- Git pre-commit hook generation via `hook` command (US-016)
- GitHub Actions workflow generation via `ci init` (US-017)
- Doctor command for installation health checks and config validation (US-018)
- `--version` flag with ldflags support (US-019)
- README and LICENSE
