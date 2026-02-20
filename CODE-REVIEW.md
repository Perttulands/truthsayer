# Code Review

Reviewed: 2026-02-20

## Dead Code / Unused Files

- **`progress_truthsayer.txt`, `progress_truthsayer_multilang.txt`** — Large agent progress logs (36K and 50K) committed to the repo root. These are bead-workflow artifacts, not source code. Should be gitignored or deleted.
- **`ralph-debug.log`, `ralph-error.log`, `ralph-truthsayer.log`** — Committed log files. Logs don't belong in VCS.
- **`truthsayer-report.json`** (64K) — Scan output artifact committed to root. Should be gitignored. `.gitignore` does cover `*.json` inside `state/` but not root-level reports.
- **`truthsayer`** (9.9M binary) — Compiled binary committed to root. Not gitignored.
- **`PRD_TRUTHSAYER.md`, `PRD_TRUTHSAYER_MULTILANG.md`, `PRD-multilang.md`** — Three PRD files in the root alongside `docs/PRD.md`. At least two are duplicates or superseded; no cross-references to distinguish them.

## Unregistered Rules

- **`security_regex_rules.go`** defines `UnusedVariable`, `UnreachableCode`, `SQLInjection`, and related rules. These *are* registered in `registry.go` (lines 303–307), but they are not listed in the README or CHANGELOG under any release. Users cannot discover them from documentation.

## Missing Error Handling

- **`internal/cli/report.go`** — `defer f.Close()` discards the close error. For a report-writing command, a failed flush/close could silently produce a truncated JSON file.
- **`internal/scanner/go_scanner.go`** and **`internal/scanner/regex_scanner.go`** — `defer f.Close()` on read-only files; acceptable in context but inconsistent with the project's otherwise careful error wrapping.
- **`internal/cli/watch.go`** — `defer w.Close()` on fsnotify watcher; error discarded. Low impact.

## Inconsistencies Between Docs and Code

- **CHANGELOG (before this fix)** was missing the entire v2.0.0 multi-language expansion (28 JS/TS rules + 26 Python rules + cross-language Sprint 4). The `[Unreleased]` section had only 2 entries despite ~50 commits landing since v1.0.0.
- **README rule count** — README does not list an exact rule count, which avoids becoming stale. However, `docs/PRD.md` still references the original 24-rule scope without acknowledging the multi-language expansion.
- **`scripts/judge.sh`** uses `claude-haiku` as the default model via `TRUTHSAYER_JUDGE_MODEL`. Claude Haiku is a real model but the judge script assumes a `claude` CLI binary is available and in PATH — this dependency is undocumented (no install step in README or JUDGMENT.md).
- **`JUDGMENT.md`** describes `truthsayer judge` as a subcommand of the binary, but the prototype is a standalone bash script at `scripts/judge.sh`. The design doc and the implementation are misaligned.

## TODO / FIXME / HACK

- No `TODO`, `FIXME`, or `HACK` comments in production Go code. `context_todo.go` is a rule that *detects* `context.TODO()` — not itself a TODO.
- `js_regex_config_smells_test.go:67` has a sample string containing `// TODO: replace http://localhost:3000` as a test fixture — harmless.

## Other Observations

- **`internal/rules/bad_defaults.go`** — File exists at 1.5K but appears to be a category placeholder rather than a rule implementation. Worth confirming it's not accidentally empty.
- **`goroutine_no_context.go`** (4.5K) is the largest rule file. Its detection of goroutines launched without a context parameter is inherently heuristic and likely to produce false positives in goroutines that don't need a context (e.g., background workers with their own lifecycle). No tests visible in the rule listing — worth checking.
- **Walker dot-directory exclusion** in `internal/engine/walker.go`: logic `excludeDirs[name] || strings.HasPrefix(name, ".") && name != "."` relies on Go operator precedence (`&&` binds tighter than `||`). Correct, but fragile to read — parentheses would be clearer.
