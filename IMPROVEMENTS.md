# ⚖️ Aletheia the Truthsayer — Improvements from the Agora

_Five decrees from Athena's council chamber, where marble columns meet neon circuitry._

---

## 1. The Oracle's Memory — Persistent Scan Cache

**Problem:** Aletheia's `Engine.fileCache` (`internal/engine/engine.go:44-48`) lives only in RAM and dies with the process. Every `truthsayer scan` cold-starts, re-parsing every file with tree-sitter even when nothing changed. On large repos this is pure wasted computation — the Oracle forgets everything between invocations.

**Implementation Plan:**
- **Files:** `internal/engine/engine.go`, new `internal/cache/disk.go`
- Create `internal/cache/disk.go` implementing a bolt/bbolt or flat-file cache keyed by `(filepath, mtime_ns, file_size)` → serialized `[]finding.Finding`
- In `Engine.getCachedFindings()` / `putCachedFindings()`, add a disk cache layer that falls through from memory → disk → full scan
- Add `--no-cache` flag in `internal/cli/scan.go` and store cache at `.truthsayer-cache` (gitignored)
- Cache versioning: embed a `cacheVersion` const tied to rule registry hash so cache auto-invalidates when rules change

**Expected Impact:** 5-10× faster repeated scans on large codebases. Watch mode (`internal/watcher/watcher.go`) already benefits from the in-memory cache; disk persistence extends this across process restarts.

---

## 2. The Shield of Sarissa — Autofix Engine

**Problem:** Aletheia identifies 88 anti-patterns but offers only `Suggestion` strings (`internal/finding/finding.go:34`). Developers must manually apply every fix. A scanner that only diagnoses but never heals is a physician without hands.

**Implementation Plan:**
- **Files:** new `internal/autofix/autofix.go`, `internal/finding/finding.go`, rule files (start with 5-10 high-value rules)
- Extend `Finding` with an optional `Fix *Fix` field containing `Replacement string`, `StartLine int`, `EndLine int`
- Create `internal/autofix/autofix.go` with `Apply(path string, fixes []Fix) error` that applies non-overlapping fixes in reverse-line order
- Implement fixes for the lowest-hanging rules first: `JSEmptyCatch` (insert `console.error(e)`), `PyExceptPass` (insert `raise`), `MissingPipefail` (prepend `set -euo pipefail`), `PyBareExcept` (replace with `except Exception`)
- Add `truthsayer fix [--dry-run]` command in `internal/cli/` that scans then applies
- Dry-run mode shows unified diffs without writing

**Expected Impact:** Transforms Aletheia from a passive scanner into an active healer. Even 10 auto-fixable rules dramatically reduce friction in CI remediation workflows.

---

## 3. The Panopticon Protocol — SARIF & IDE Integration

**Problem:** Aletheia outputs only terminal text or basic JSON (`internal/report/json.go`). Neither format integrates with GitHub Code Scanning, VS Code Problems panel, or any standard toolchain. The Oracle speaks, but no one in the agora can hear.

**Implementation Plan:**
- **Files:** new `internal/report/sarif.go`, `internal/cli/scan.go`, `internal/cli/report.go`
- Implement SARIF v2.1.0 output (the OASIS standard consumed by GitHub Advanced Security, VS Code SARIF Viewer, Azure DevOps)
- Map `Finding.Rule` → SARIF `reportingDescriptor`, `Finding.Severity` → SARIF level, `Finding.File/Line` → SARIF `physicalLocation`
- Add `--format sarif` flag alongside existing `json` and `terminal` in `internal/cli/scan.go`
- In `internal/cli/ci.go`, update the generated GitHub Actions workflow to upload SARIF via `github/codeql-action/upload-sarif`

**Expected Impact:** Findings surface directly in GitHub PR annotations and IDE problem panels. Zero-friction integration with the broader security tooling ecosystem.

---

## 4. The Loom of Arachne — Custom Rule Plugin System

**Problem:** All 88 rules are hardcoded in `internal/rules/registry.go` (`DefaultRegistry()`). Users cannot add project-specific rules without forking Aletheia. Every codebase has domain-specific anti-patterns (e.g., "never call `db.Exec` without a transaction in our repo") that the built-in rules can't know about.

**Implementation Plan:**
- **Files:** `internal/config/config.go`, new `internal/rules/custom.go`, `internal/rules/registry.go`
- Add `[rules.custom]` section to `.truthsayer.toml` config supporting regex-based custom rules:
  ```toml
  [[rules.custom]]
  id = "project.no-raw-db-exec"
  pattern = 'db\.Exec\('
  file_types = [".go"]
  severity = "error"
  message = "Use db.ExecContext with transaction"
  suggestion = "Wrap in tx.ExecContext(ctx, ...)"
  ```
- Create `internal/rules/custom.go` that parses these into `RegexChecker` implementations and registers them in the registry
- Load custom rules in `internal/cli/scan.go` after `config.Load()`, before engine creation
- Support `exclude_paths` per custom rule for scoping

**Expected Impact:** Turns Aletheia from a fixed scanner into an extensible platform. Teams can encode institutional knowledge as rules without Go code or recompilation.

---

## 5. The Threads of Moirai — Incremental Git-Aware Scanning

**Problem:** `truthsayer scan .` walks the entire directory tree (`internal/engine/walker.go`) and scans every file. In CI, only changed files matter. The `internal/diff/diff.go` `Tracker` does LCS-based line diffing but only works within a single process session (watch mode). There's no git integration — the Fates' threads are unwoven.

**Implementation Plan:**
- **Files:** new `internal/git/changed.go`, `internal/engine/engine.go`, `internal/cli/scan.go`, `internal/diff/diff.go`
- Create `internal/git/changed.go` with `ChangedFiles(base string) ([]string, error)` using `git diff --name-only <base>...HEAD` and `ChangedLines(base, path string) (map[int]bool, error)` using `git diff -U0`
- Add `--since <ref>` flag to `truthsayer scan` (e.g., `--since main`, `--since HEAD~3`)
- When `--since` is set, replace `Walk()` with `git.ChangedFiles()`, then use `finding.FilterByLines()` (already exists at `internal/finding/finding.go:51`) with the git-derived changed lines
- In `internal/cli/ci.go`, default the generated workflow to `--since ${{ github.event.pull_request.base.sha }}`

**Expected Impact:** CI scan time drops from O(repo) to O(diff). On a 10k-file repo with a 5-file PR, this is a 2000× reduction in files scanned. Combined with SARIF output (#3), this creates a tight feedback loop: only new sins are judged.

---

_"The unexamined code is not worth shipping." — Socrates, if he'd been a Site Reliability Engineer_
