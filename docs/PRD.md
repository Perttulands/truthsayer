# Aletheia the Truthsayer — Product Requirements Document

> Development anti-pattern scanner. Detects bad practices that mask problems: silent fallbacks, swallowed errors, bad defaults, mocks in production paths, missing traces.

## 1. Overview

Truthsayer is a CLI tool written in Go that scans codebases for development anti-patterns — code constructs that hide failures, degrade debuggability, or leak test artifacts into production. It uses AST-based analysis for Go source files and regex-based pattern matching for bash scripts, config files, and general text.

Truthsayer operates in two modes:
- **Active (scan)**: On-demand full or partial scan producing structured JSON reports
- **Passive (watch)**: File-system watcher that alerts on newly introduced anti-patterns in real-time

The tool is designed to run standalone with zero external dependencies beyond the Go standard library and `fsnotify`.

## 2. Goals

- Detect all 6 anti-pattern categories defined in the spec (silent fallbacks, missing error context, trace/log gaps, mock leakage, bad defaults, configuration smells)
- Zero false positives on standard Go and bash idioms
- Scan 10k LOC in under 5 seconds
- Produce machine-parseable JSON reports consumable by CI systems and other tools
- Ship as a single static binary with no runtime dependencies
- Support incremental adoption via per-rule enable/disable in TOML config
- Integrate into CI/CD pipelines and git hooks as a quality gate

## 3. Non-Goals

- Auto-fixing code (suggestions only, never auto-apply)
- Supporting languages beyond Go and bash/shell in v1
- Database or persistent storage (JSON files only)
- GUI or web interface
- Real-time collaboration or multi-user features
- LLM-based classification (deferred to v2; spec mentions optional `claude -p` but v1 is deterministic rules only)

## 4. User Stories

### Scanner Core

**US-001**: As a developer, I want to scan a directory for anti-patterns so I can find hidden problems before they reach production.

**US-002**: As a developer, I want to scan a single file so I can check my work before committing.

**US-003**: As a developer, I want findings to include file path, line number, code snippet, and a fix suggestion so I can act on them immediately.

**US-004**: As a developer, I want scan results sorted by severity (error > warning > info) so I can triage effectively.

**US-005**: As a developer, I want the scanner to skip vendor/, node_modules/, and .git/ directories by default so scans are fast and relevant.

### Watch Mode

**US-006**: As a developer, I want to watch a directory for file changes and get instant feedback on new anti-patterns so I catch problems as I code.

**US-007**: As a developer, I want watch mode to only report findings in the changed lines (not the entire file) so I'm not overwhelmed by pre-existing issues.

### Reporting

**US-008**: As a developer, I want to generate a JSON report of all findings so I can feed it into other tools or dashboards.

**US-009**: As a CI engineer, I want the scanner to exit with code 1 when errors are found so I can use it as a pipeline quality gate.

**US-010**: As a developer, I want a human-readable terminal summary after each scan showing counts by severity and category.

### Configuration

**US-011**: As a team lead, I want to configure which rules are enabled/disabled via a TOML config file so I can tailor the scanner to our codebase.

**US-012**: As a developer, I want to exclude specific files or directories from scanning via config so I can skip generated code or legacy modules.

**US-013**: As a developer, I want to override severity levels per rule so I can promote warnings to errors for rules my team cares about.

### Rules & Discovery

**US-014**: As a developer, I want to list all available detection rules with their IDs, descriptions, and severity so I understand what the scanner checks.

**US-015**: As a developer, I want to list only the currently enabled rules so I know what's active in my project.

### Integration

**US-016**: As a developer, I want to run Truthsayer as a git pre-commit hook on staged files so anti-patterns are caught before they're committed.

**US-017**: As a CI engineer, I want to run Truthsayer in a GitHub Actions workflow and have it fail the build on errors.

### Installation & Health

**US-018**: As a developer, I want a `doctor` command that checks my installation, config validity, and reports the active rule count so I can verify everything works.

**US-019**: As a developer, I want `--version` to print the version string so I can verify which build I'm running.

## 5. Functional Requirements

### 5.1 Scanner Engine

**FR-001**: The scanner MUST parse Go files using `go/ast` and `go/parser` from the standard library.

**FR-002**: The scanner MUST analyze bash/shell scripts and config files using compiled regex patterns.

**FR-003**: The scanner MUST walk the directory tree recursively, respecting exclude patterns from config.

**FR-004**: The scanner MUST skip `.git/`, `vendor/`, `node_modules/`, and `testdata/` directories by default.

**FR-005**: The scanner MUST process files concurrently using goroutines bounded by `runtime.NumCPU()`.

**FR-006**: The scanner MUST complete a 10k LOC scan in under 5 seconds on commodity hardware.

### 5.2 Detection Rules

**FR-007**: The scanner MUST implement detection for all 6 categories: silent fallbacks, missing error context, trace/log gaps, mock leakage, bad defaults, configuration smells.

**FR-008**: Each rule MUST have a unique ID in the format `category.rule-name` (e.g., `silent-fallback.empty-catch`).

**FR-009**: Each rule MUST have a severity level: `error`, `warning`, or `info`.

**FR-010**: Each rule MUST produce a finding with: rule ID, severity, file path, line number, matched code snippet, human-readable message, and fix suggestion.

**FR-011**: Rules MUST be individually toggleable (enable/disable) via config.

**FR-012**: Rules MUST support severity override via config.

### 5.3 Findings

**FR-013**: Each finding MUST include: `rule` (string), `severity` (enum), `file` (relative path), `line` (int), `code` (string, the matched source), `message` (string), `suggestion` (string).

**FR-014**: Findings MUST be deduplicated — the same rule on the same file:line MUST NOT appear twice.

**FR-015**: Findings MUST be sorted by severity (error first), then by file path, then by line number.

### 5.4 Report Output

**FR-016**: The `report` command MUST produce a JSON file matching the schema in section 8.

**FR-017**: The `scan` command MUST print a human-readable summary to stdout with counts by severity and category.

**FR-018**: The `scan` command MUST accept `--format json` to output JSON to stdout instead of the human-readable summary.

**FR-019**: Exit code MUST be 0 if no errors found, 1 if any error-severity findings exist, 2 for tool failures.

### 5.5 Watch Mode

**FR-020**: The `watch` command MUST use `fsnotify` to monitor file-system events.

**FR-021**: The `watch` command MUST debounce rapid file changes (100ms window) to avoid redundant scans.

**FR-022**: The `watch` command MUST only scan files matching supported extensions (`.go`, `.sh`, `.bash`, `.toml`, `.yaml`, `.yml`, `.json`, `.env`).

**FR-023**: The `watch` command MUST print findings to stdout in real-time as they are detected.

### 5.6 Configuration

**FR-024**: The tool MUST look for config at `.truthsayer.toml` in the scanned directory, then `~/.config/truthsayer/config.toml`, then built-in defaults.

**FR-025**: The tool MUST accept `--config <path>` to override config file location.

**FR-026**: All rules MUST be enabled by default when no config exists.

### 5.7 CLI

**FR-027**: The CLI MUST implement all commands defined in section 7.

**FR-028**: The CLI MUST support `--quiet` flag to suppress non-error output.

**FR-029**: The CLI MUST support `--verbose` flag for debug-level output.

## 6. Technical Architecture

### 6.1 Project Layout

```
truthsayer/
├── cmd/
│   └── truthsayer/
│       └── main.go              # Entry point, CLI dispatch
├── internal/
│   ├── cli/
│   │   ├── root.go              # Root command, global flags
│   │   ├── scan.go              # scan command
│   │   ├── watch.go             # watch command
│   │   ├── check.go             # check command (single file)
│   │   ├── report.go            # report command (JSON output)
│   │   ├── rules.go             # rules listing command
│   │   └── doctor.go            # doctor command
│   ├── config/
│   │   ├── config.go            # TOML config loading and merging
│   │   └── config_test.go
│   ├── engine/
│   │   ├── engine.go            # Scan orchestrator, concurrency
│   │   ├── engine_test.go
│   │   ├── walker.go            # Directory tree walker
│   │   └── walker_test.go
│   ├── rules/
│   │   ├── registry.go          # Rule registry, enable/disable
│   │   ├── registry_test.go
│   │   ├── rule.go              # Rule interface and types
│   │   ├── silent_fallback.go   # Category 1 rules
│   │   ├── error_context.go     # Category 2 rules
│   │   ├── trace_gaps.go        # Category 3 rules
│   │   ├── mock_leakage.go      # Category 4 rules
│   │   ├── bad_defaults.go      # Category 5 rules
│   │   └── config_smells.go     # Category 6 rules
│   ├── scanner/
│   │   ├── go_scanner.go        # Go AST-based scanner
│   │   ├── go_scanner_test.go
│   │   ├── regex_scanner.go     # Regex-based scanner for bash/config
│   │   └── regex_scanner_test.go
│   ├── finding/
│   │   ├── finding.go           # Finding struct, sorting, dedup
│   │   └── finding_test.go
│   ├── report/
│   │   ├── json.go              # JSON report generation
│   │   ├── json_test.go
│   │   ├── terminal.go          # Human-readable terminal output
│   │   └── terminal_test.go
│   └── watcher/
│       ├── watcher.go           # fsnotify wrapper, debounce
│       └── watcher_test.go
├── testdata/                    # Fixture files for testing rules
│   ├── go/                      # Go source fixtures
│   ├── bash/                    # Bash script fixtures
│   └── config/                  # Config file fixtures
├── docs/
│   ├── SPEC.md
│   └── PRD.md
├── .truthsayer.toml             # Example/default config
├── .gitignore
└── go.mod
```

### 6.2 Dependency Graph

```
cmd/truthsayer/main.go
  └── internal/cli/*
        ├── internal/config
        ├── internal/engine
        │     ├── internal/scanner/go_scanner
        │     ├── internal/scanner/regex_scanner
        │     ├── internal/rules/registry
        │     └── internal/finding
        ├── internal/report
        └── internal/watcher
```

### 6.3 External Dependencies

| Dependency | Purpose | Justification |
|---|---|---|
| `github.com/BurntSushi/toml` | TOML config parsing | De facto Go TOML library, stdlib-quality |
| `github.com/fsnotify/fsnotify` | File system watching | Required for watch mode, cross-platform |

All other functionality uses Go standard library only (`go/ast`, `go/parser`, `go/token`, `os`, `path/filepath`, `regexp`, `encoding/json`, `flag`).

### 6.4 Concurrency Model

The scan engine dispatches file processing across a worker pool. Each file is scanned independently:

```
engine.Run(path)
  ├── walker.Walk(path) → []FilePath
  └── pool(NumCPU workers)
        ├── worker: scanner.ScanGo(file) → []Finding
        ├── worker: scanner.ScanRegex(file) → []Finding
        └── ...
  └── collect, dedup, sort → []Finding
```

Files are dispatched to workers via a buffered channel. Workers determine the scanner type based on file extension. Findings are collected via a results channel, then deduplicated and sorted.

## 7. CLI Interface

```
truthsayer scan <path>                  # Full recursive scan
truthsayer scan --format json <path>    # JSON output to stdout
truthsayer scan --fix <path>            # Include fix suggestions in output
truthsayer check <file>                 # Single file scan
truthsayer watch <path>                 # Watch mode with fsnotify
truthsayer report <path>                # Generate JSON report to file
truthsayer report --output <file> <path>  # Specify report output path
truthsayer rules                        # List all rules
truthsayer rules --enabled              # List enabled rules only
truthsayer doctor                       # Check installation and config
truthsayer --version                    # Print version
truthsayer --help                       # Print usage
```

### Global Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--config` | string | auto-detect | Path to config file |
| `--quiet` | bool | false | Suppress non-error output |
| `--verbose` | bool | false | Debug-level output |

### Scan Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--format` | string | `text` | Output format: `text` or `json` |
| `--fix` | bool | false | Include fix suggestions |
| `--severity` | string | `all` | Minimum severity: `error`, `warning`, `info` |

### Report Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--output` | string | `truthsayer-report.json` | Report output file path |

## 8. Report Format

### JSON Report Schema

```json
{
  "version": "1",
  "scan_time": "2026-02-13T15:00:00Z",
  "path": "/projects/myapp",
  "config": ".truthsayer.toml",
  "duration_ms": 342,
  "findings": [
    {
      "rule": "silent-fallback.empty-catch",
      "severity": "error",
      "file": "pkg/handler.go",
      "line": 42,
      "code": "if err != nil {\n\treturn nil\n}",
      "message": "Error returned as nil without logging or wrapping",
      "suggestion": "Return the error or log it: return fmt.Errorf(\"handler failed: %w\", err)"
    }
  ],
  "summary": {
    "total": 12,
    "errors": 3,
    "warnings": 7,
    "info": 2,
    "files_scanned": 87,
    "duration_ms": 342
  }
}
```

### Terminal Output Format

```
truthsayer scan results — /projects/myapp
═══════════════════════════════════════════

ERROR  silent-fallback.empty-catch
  pkg/handler.go:42
  Error returned as nil without logging or wrapping
  → Return the error or log it: return fmt.Errorf("handler failed: %w", err)

WARN   trace-gaps.no-request-id
  internal/api/server.go:18
  HTTP handler missing request ID propagation
  → Add request ID via middleware or context

──────────────────────────────────────────
Summary: 3 errors, 7 warnings, 2 info (87 files scanned in 342ms)
```

## 9. Detection Rules Schema

### Rule Definition

Each rule is defined as a Go struct implementing the `Rule` interface:

```go
type Rule struct {
    ID          string   // "category.rule-name"
    Category    string   // "silent-fallback", "error-context", etc.
    Name        string   // Human-readable name
    Description string   // What this rule detects
    Severity    Severity // error, warning, info
    FileTypes   []string // [".go"], [".sh", ".bash"], ["*"]
    ScanType    ScanType // AST or Regex
}

type Severity string
const (
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
    SeverityInfo    Severity = "info"
)

type ScanType int
const (
    ScanTypeAST   ScanType = iota
    ScanTypeRegex
)
```

### Rule Interface

```go
type Checker interface {
    Check(ctx *ScanContext) []Finding
}

// For AST-based rules (Go files)
type ASTChecker interface {
    Checker
    CheckAST(fset *token.FileSet, node ast.Node) []Finding
}

// For regex-based rules (bash, config, general)
type RegexChecker interface {
    Checker
    CheckLines(path string, lines []string) []Finding
}
```

### Initial Rule Set

| ID | Category | Severity | Scan Type | Description |
|---|---|---|---|---|
| `silent-fallback.empty-error-check` | silent-fallback | error | AST | `if err != nil { return nil }` without logging |
| `silent-fallback.ignored-error` | silent-fallback | error | AST | Error return value assigned to `_` |
| `silent-fallback.bare-return-on-error` | silent-fallback | warning | AST | Named return with `err != nil` returns zero values silently |
| `silent-fallback.hidden-failure-bash` | silent-fallback | error | Regex | `\|\| true`, `2>/dev/null` in shell scripts |
| `silent-fallback.no-err-trap` | silent-fallback | warning | Regex | `set -e` without `trap ... ERR` in bash |
| `error-context.generic-message` | error-context | warning | AST | `errors.New("failed")` without variable context |
| `error-context.unwrapped-error` | error-context | warning | AST | `return err` without `fmt.Errorf("...: %w", err)` wrapping |
| `error-context.http-200-on-error` | error-context | error | AST | HTTP handler writes 200 status after error check |
| `error-context.nil-on-error` | error-context | error | AST | Function returns `nil, nil` or `"", nil` when error path detected |
| `trace-gaps.long-function-no-log` | trace-gaps | warning | AST | Function >20 lines with no log/trace call |
| `trace-gaps.error-path-no-log` | trace-gaps | warning | AST | Error branch without structured logging |
| `trace-gaps.no-request-id` | trace-gaps | warning | AST | HTTP handler without request ID in context |
| `trace-gaps.no-stderr-capture` | trace-gaps | info | Regex | `exec.Command` without stderr pipe |
| `mock-leakage.mock-import-non-test` | mock-leakage | error | AST | Mock/testify import in non-`_test.go` file |
| `mock-leakage.test-fixture-ref` | mock-leakage | error | Regex | `testdata/` or `fixture` path in non-test file |
| `mock-leakage.debug-guard` | mock-leakage | warning | AST | `if debug` / `if testing` guard in production code |
| `bad-defaults.magic-number` | bad-defaults | info | AST | Numeric literal >1 in function body without const |
| `bad-defaults.no-timeout` | bad-defaults | error | AST | `http.Client{}` or `net.Dial` without explicit timeout |
| `bad-defaults.missing-pipefail` | bad-defaults | error | Regex | Bash script without `set -euo pipefail` |
| `bad-defaults.unvalidated-env` | bad-defaults | warning | AST | `os.Getenv()` used directly without validation |
| `config-smells.secret-in-config` | config-smells | error | Regex | Patterns like `password=`, `token=`, `secret=` with inline values |
| `config-smells.hardcoded-path` | config-smells | warning | Regex | Absolute paths like `/home/`, `/Users/`, `/opt/` in config |
| `config-smells.missing-gitignore` | config-smells | warning | Regex | `.env` file present but not in `.gitignore` |

## 10. Integration Points

### CI/CD (GitHub Actions)

```yaml
- name: Truthsayer scan
  run: |
    truthsayer scan --format json . > truthsayer-report.json
    truthsayer scan .  # exits 1 on errors
```

### Git Pre-Commit Hook

```bash
#!/bin/bash
staged=$(git diff --cached --name-only --diff-filter=ACM)
if [ -n "$staged" ]; then
  echo "$staged" | xargs -I{} truthsayer check {}
fi
```

### Exit Codes

| Code | Meaning |
|---|---|
| 0 | Scan completed, no error-severity findings |
| 1 | Scan completed, error-severity findings present |
| 2 | Tool error (bad config, invalid path, internal failure) |

## 11. Configuration (TOML)

### Default Config: `.truthsayer.toml`

```toml
# Truthsayer configuration

[scan]
# Directories to exclude from scanning (always excludes .git)
exclude_dirs = ["vendor", "node_modules", "testdata", ".git"]

# File patterns to exclude
exclude_patterns = ["*_generated.go", "*.pb.go", "*.min.js"]

# Max file size to scan (bytes) — skip large generated files
max_file_size = 1048576  # 1MB

[rules]
# Override severity: rule_id = "new_severity"
# Example: promote a warning to error
# "trace-gaps.long-function-no-log" = "error"

[rules.disable]
# List of rule IDs to disable
# ids = ["bad-defaults.magic-number"]

[watch]
# Debounce interval in milliseconds
debounce_ms = 100

# File extensions to watch
extensions = [".go", ".sh", ".bash", ".toml", ".yaml", ".yml", ".json", ".env"]

[report]
# Default report output path
output = "truthsayer-report.json"
```

## 12. Testing Strategy

### Unit Tests

Every rule gets a dedicated test with fixture files in `testdata/`:

```
testdata/
├── go/
│   ├── empty_error_check.go        # Should trigger silent-fallback.empty-error-check
│   ├── proper_error_handling.go     # Should NOT trigger (negative case)
│   ├── ignored_error.go            # Should trigger silent-fallback.ignored-error
│   └── ...
├── bash/
│   ├── no_pipefail.sh              # Should trigger bad-defaults.missing-pipefail
│   ├── proper_bash.sh              # Should NOT trigger
│   └── ...
└── config/
    ├── secret_in_config.toml       # Should trigger config-smells.secret-in-config
    └── clean_config.toml           # Should NOT trigger
```

Each test asserts:
1. The expected rule ID fires on the positive fixture
2. The correct line number is reported
3. The finding includes a non-empty message and suggestion
4. The negative fixture produces zero findings

### Integration Tests

- Scan the `testdata/` directory as a whole, verify combined report matches expected findings
- Test config loading with overrides (disable rules, change severity)
- Test exit codes: 0 for clean, 1 for errors, 2 for bad input
- Test `--format json` output parses as valid JSON matching the schema

### Benchmarks

- Benchmark scanning of 10k LOC Go corpus (generated or snapshot)
- Assert <5 second completion
- Profile memory usage to verify no leaks in watch mode

### Test Commands

```bash
go test ./...                           # All unit tests
go test ./internal/rules/... -v         # Rule tests with output
go test ./internal/engine/... -bench .  # Engine benchmarks
go test -race ./...                     # Race condition detection
```

## 13. Success Metrics

| Metric | Target | How to Measure |
|---|---|---|
| Category coverage | 6/6 categories with ≥2 rules each | `truthsayer rules \| count by category` |
| False positive rate | 0% on Go stdlib idioms | Scan Go stdlib subset, verify 0 false positives |
| Scan performance | <5s for 10k LOC | Benchmark test on generated corpus |
| Report validity | 100% valid JSON | Schema validation in integration tests |
| Binary size | <15MB static binary | `ls -la` after `go build` |
| Test coverage | ≥80% line coverage | `go test -coverprofile` |
| Zero dependencies at runtime | stdlib + toml + fsnotify only | `go list -m all` |

## 14. Open Questions

1. **Watch mode scope**: Should watch mode scan only changed lines (diff-aware) or the entire changed file? Diff-aware is more useful but significantly more complex. **Recommendation**: v1 scans the full changed file; v2 adds diff-aware filtering.

2. **Rule extensibility**: Should users be able to define custom regex rules in config? **Recommendation**: Defer to v2. Keep v1 rules compiled-in for performance and simplicity.

3. **Baseline/suppress**: Should there be a mechanism to baseline existing findings so only new ones are reported? **Recommendation**: Yes, add `truthsayer baseline` command that snapshots current findings. Implement after core scan works.

4. **LLM integration**: The spec mentions optional `claude -p --model haiku` for ambiguous patterns. **Recommendation**: Defer to v2. v1 is fully deterministic.

5. **Severity exit codes**: Should warnings also cause non-zero exit? **Recommendation**: No. Only error-severity triggers exit code 1. Add `--fail-on warning` flag if teams want stricter gating.

## 15. Implementation Sequence

This PRD is designed for sequential task-based implementation. Build order:

1. **Project scaffold** — `cmd/`, `internal/` layout, go.mod dependencies
2. **Finding types** — `internal/finding/` structs, sorting, dedup
3. **Rule interface & registry** — `internal/rules/` types, registration, enable/disable
4. **Go AST scanner** — `internal/scanner/go_scanner.go` with first rule (`silent-fallback.empty-error-check`)
5. **Regex scanner** — `internal/scanner/regex_scanner.go` with first rule (`bad-defaults.missing-pipefail`)
6. **Scan engine** — `internal/engine/` walker + concurrent dispatch + collection
7. **Terminal reporter** — `internal/report/terminal.go` human-readable output
8. **JSON reporter** — `internal/report/json.go` structured output
9. **Config loader** — `internal/config/` TOML parsing, merging, validation
10. **CLI commands: scan, check** — `internal/cli/scan.go`, `check.go`
11. **CLI commands: report, rules, doctor** — remaining CLI commands
12. **Silent fallback rules (remaining)** — complete category 1
13. **Error context rules** — category 2
14. **Trace gap rules** — category 3
15. **Mock leakage rules** — category 4
16. **Bad defaults rules (remaining)** — complete category 5
17. **Config smell rules** — category 6
18. **Watch mode** — `internal/watcher/` fsnotify integration
19. **CLI commands: watch, version** — watch command + global flags
20. **Integration tests & benchmarks** — end-to-end verification
21. **Example config & documentation** — `.truthsayer.toml`, help text polish
