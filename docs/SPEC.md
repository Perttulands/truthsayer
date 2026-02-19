# Truthsayer — Development Anti-Pattern Scanner

## Name
Truthsayer — the one who sees through comfortable lies. Detects bad practices that mask problems: silent fallbacks, swallowed errors, bad defaults, mocks in production paths, missing traces.

## Purpose
Truthsayer scans codebases and configurations for anti-patterns that hide problems. In dev mode, every failure must be visible. Truthsayer ensures:
- Errors are never swallowed
- Fallbacks don't mask broken paths
- Logging and tracing exist where needed
- Mocks don't leak into production
- Defaults are explicit, not silent

## Architecture

### Modes

**Passive (daemon)**: Watches file changes via fsnotify. When code is modified, scans the diff for newly introduced anti-patterns. Alerts in real-time.

**Active (scan)**: On-demand full scan of a repo or directory. Produces a structured report of all findings.

### Detection Categories

#### 1. Silent Fallbacks
- Catch blocks that swallow errors (empty catch, bare except)
- Default values that mask missing configuration
- `|| true` / `2>/dev/null` hiding failures
- Fallback model/service/endpoint without explicit logging
- `set -e` without ERR trap in bash scripts

#### 2. Missing Error Context
- Error messages without variable interpolation (generic "failed")
- Catch/rescue without re-raise or logging
- Functions that return nil/null on error without indication
- HTTP handlers that return 200 on internal errors
- Missing error return values in Go

#### 3. Trace & Log Gaps
- Functions >20 lines with no log/trace statements
- Error paths without structured logging
- Missing request IDs in HTTP handlers
- No timestamps in log output
- stderr not captured in subprocess calls

#### 4. Mock Leakage
- Mock/stub imports outside test files
- Test fixtures referenced in production code
- Hardcoded test values in non-test files
- `if TEST` / `if DEBUG` guards in production paths

#### 5. Bad Defaults
- Magic numbers without named constants
- Implicit timeouts (no timeout = infinite wait)
- Missing `set -euo pipefail` in bash scripts
- Unvalidated environment variables used directly
- Default credentials or tokens

#### 6. Configuration Smells
- Secrets in plaintext config files
- Personal information in committed files (emails, IPs, paths with usernames)
- Hardcoded paths instead of env vars
- Missing .gitignore entries for sensitive files

## Tech Stack
- **Language**: Go (CLI tool)
- **Scanner**: AST-based for Go files, regex-based for bash/config/general
- **Storage**: JSON report files + file-based precedent store (`precedents.json`)
- **Runtime**: CLI tool (no daemon dependency), optional systemd for passive mode
- **Config**: TOML
- **LLM**: Optional `claude -p --model haiku` for ambiguous pattern classification

## Interface

```
truthsayer scan <path>           # Full scan of directory/repo
truthsayer scan --fix <path>     # Scan + suggest fixes (no auto-apply)
truthsayer watch <path>          # Passive mode — watch for changes
truthsayer check <file>          # Single file check
truthsayer report <path>         # Generate structured JSON report
truthsayer rules                 # List all detection rules
truthsayer rules --enabled       # List enabled rules
truthsayer doctor                # Check installation
truthsayer --version             # Print version
```

## Report Format

```json
{
  "scan_time": "2026-02-13T15:00:00Z",
  "path": "/home/user/project",
  "findings": [
    {
      "rule": "silent-fallback.empty-catch",
      "severity": "error",
      "file": "pkg/handler.go",
      "line": 42,
      "code": "catch (err) {}",
      "message": "Empty catch block swallows error silently",
      "suggestion": "Log the error or re-raise: log.Error(\"handler failed\", err)"
    }
  ],
  "summary": {
    "total": 12,
    "errors": 3,
    "warnings": 7,
    "info": 2
  }
}
```

## Precedent Schema

Truthsayer keeps historical violation decisions in a local JSON file (`precedents.json`) so past decisions can inform future judgments.

```json
[
  {
    "rule_id": "silent-fallback.empty-error-check",
    "violation_hash": "3f962cc15b...",
    "decision": "deny",
    "rationale": "Swallowing errors in this code path caused incidents before.",
    "created_at": "2026-02-19T00:00:00Z"
  }
]
```

Field definitions:
- `rule_id`: canonical rule identifier
- `violation_hash`: stable identifier for the violation instance
- `decision`: `allow` or `deny`
- `rationale`: explanation for why this precedent exists
- `created_at`: decision timestamp in RFC3339 format

Implementation package: `internal/precedent`
- `NewStore(path string) *Store`
- `(*Store).Load() ([]Precedent, error)`
- `(*Store).Save([]Precedent) error`
- `(*Store).Add(Precedent) error`
- `(*Store).Query(ruleID, violationHash string) (Precedent, bool, error)`
- `QueryByRule(precedents []Precedent, ruleID string) []Precedent`

## Severity Levels
- **error**: Must fix. Actively hides problems (empty catch, swallowed errors, secrets in code)
- **warning**: Should fix. Degrades debuggability (missing logs, generic error messages)
- **info**: Consider fixing. Style/hygiene (magic numbers, missing comments on complex logic)

## Integration
- **CI/CD**: Exit code 1 if any errors found (use as quality gate)
- **Git hooks**: Pre-commit scan of staged files
- **Argus**: Alert on new findings in watched repos
- **Oathkeeper**: Cross-reference — if Oathkeeper detects commitment language, Truthsayer verifies the backing mechanism has proper error handling

## Success Criteria
- Detects all 6 anti-pattern categories
- Zero false positives on standard Go/bash idioms
- Scan of 10k LOC completes in <5 seconds
- JSON report parseable by other tools
- Works standalone (no external dependencies beyond Go stdlib)
- No personal information in repo (git author anonymized)
