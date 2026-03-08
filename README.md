# Aletheia the Truthsayer

![Aletheia Banner](banner.png)

*Bare feet. Linen. The Codex under one arm. She reads your code like a confession.*

---

Truthsayer is a static analysis tool that finds the lies your code tells. Not style violations or naming complaints — the specific places where your code swallows an error, catches an exception and does nothing with it, returns HTTP 200 from a catch block, or quietly drops a failure on the floor. The kind of thing that works fine in dev and wakes you up at 3am in production. It covers Go, JavaScript, TypeScript, Python, Rust, Bash, and config files, with 104 rules across 11 categories, and it builds as a single Go binary.

---

Every codebase has a layer of polite lies. Error handlers that handle nothing. Catch blocks that catch and release. Timeouts that were going to be added "later." Somewhere under the clean abstractions, something is swallowing failures and smiling about it.

Aletheia sees both layers. She wears undyed linen — plain, undecorated, what's left when everything ornamental is stripped away. A dark stole across one shoulder with every rule stitched in gold thread. The Codex hangs at her side, heavy enough to use as a weapon if the Red Quill doesn't make the point clearly enough. And the bare feet — Pseudos, the false copy Prometheus's apprentice made, had no feet. Hers are the proof she's real.

She doesn't care about style. She doesn't care about naming conventions. She cares about the specific moment your code decided to lie about what happened. That `catch (e) {}` you wrote at 2am? She found it. She's been waiting.

Aletheia's niche is *failure-hiding anti-patterns* — the silent fallbacks, swallowed exceptions, missing observability, and production-test boundary violations that generic linters ignore.

## Current Status

| Area | Status |
|------|--------|
| Go scanning (AST) | ✅ 21 rules, stdlib parser |
| JS/TS scanning (AST) | ✅ 20 rules via tree-sitter |
| Python scanning (AST) | ✅ 19 rules via tree-sitter |
| Rust scanning (regex) | ✅ 10 rules |
| Bash/Config scanning (regex) | ✅ 12 rules |
| Cross-language rules (security, code-quality) | ✅ 6 rules |
| JS/TS regex rules | ✅ 9 rules |
| Python regex rules | ✅ 7 rules |
| `scan`, `check`, `watch`, `report` | ✅ Working |
| `hook`, `hook install`, `ci`, `ci init` | ✅ Working |
| `rules`, `doctor`, `version` | ✅ Working |
| `judge` (LLM-assisted triage) | ✅ Working — requires `ANTHROPIC_API_KEY` |
| `senate parse` / `senate apply` | ✅ Working |
| `warmup`, `debt` | ✅ Working |
| Precedent system | ✅ File-based, working |
| Parallel scanning | ✅ `--parallel` flag |
| Bead creation | ⚠️ Requires external `br` binary |
| Windows support | ⚠️ Untested — fsnotify should work but no CI coverage |

All 19 internal packages pass tests. Builds clean on Linux with Go 1.25.0 and a C compiler.

## Install

Requires a C compiler (gcc or clang) for tree-sitter parsing.

```bash
go install github.com/perttulands/truthsayer/cmd/truthsayer@latest
```

Or build from source:

```bash
git clone https://github.com/Perttulands/truthsayer.git
cd truthsayer
go build -o truthsayer ./cmd/truthsayer
```

On Debian/Ubuntu, install C toolchain if needed:

```bash
sudo apt-get install build-essential
```

## Usage

```bash
# Scan a directory (all languages)
truthsayer scan .

# Scan specific languages only
truthsayer scan --lang go,python .
truthsayer scan --lang js,ts .

# Scan a single file
truthsayer check path/to/file.go

# Watch for changes
truthsayer watch .

# JSON report
truthsayer scan --format json .

# Reuse past allow/deny decisions from precedents.json
truthsayer scan --use-precedents .

# List all rules
truthsayer rules

# List rules for a specific language
truthsayer rules --lang python

# Check installation
truthsayer doctor
```

### Language Aliases

The `--lang` flag accepts these aliases:

| Alias | Extensions |
|-------|-----------|
| `go` | `.go` |
| `js`, `javascript` | `.js`, `.jsx`, `.mjs`, `.cjs` |
| `ts`, `typescript` | `.ts`, `.tsx` |
| `python`, `py` | `.py`, `.pyi` |
| `bash`, `shell`, `sh` | `.sh`, `.bash` |

## Supported Languages and Rule Count

| Language | Parser | Rule Count |
|----------|--------|-----------|
| Go | `go/ast` (stdlib) | 21 AST rules |
| JavaScript/TypeScript | tree-sitter (cgo) | 20 AST + 9 regex rules |
| Python | tree-sitter (cgo) | 19 AST + 7 regex rules |
| Rust | regex | 10 regex rules |
| Bash | regex | 12 regex rules (shared with config/test) |
| Config files (`.toml`, `.yaml`, `.yml`, `.json`, `.env`) | regex | (shared with bash/test rules) |
| Cross-language (security, code-quality) | regex | 6 rules |

**Total: 104 built-in rules** across 11 categories.

## What It Detects

| Category | What it catches |
|----------|----------------|
| **silent-fallback** | Swallowed errors, empty catch blocks, bare except, floating promises, ignored callbacks |
| **error-context** | Generic error messages, unwrapped errors, HTTP 200 on error, raise-from-none |
| **trace-gaps** | Functions without logging, missing request IDs, no unhandled rejection handler |
| **mock-leakage** | Test imports in production, debug guards, jest.mock in source, pytest fixtures in source |
| **bad-defaults** | Missing timeouts, no pipefail, eval usage, mutable default args, star imports |
| **config-smells** | Hardcoded paths, secrets in config, missing .gitignore, unpinned requirements |
| **test-isolation** | Missing cleanup in beforeAll/afterAll, leaked servers and SSE connections |
| **rust** | Unwrap panics, ignored results, unsafe blocks without safety comments |
| **security** | Command injection, SQL injection, hardcoded credentials |
| **code-quality** | Error swallowing, unreachable code, unused variables |
| **test** | Test-specific anti-patterns |

## CLI Reference

### Shared Flag

`--config <path>` is accepted by: `scan`, `check`, `watch`, `report`, `rules`, `doctor`, `hook`, `ci`, `warmup`.

When omitted, defaults to `<scan target>/.truthsayer.toml` (or cwd for `doctor`/`rules`); if missing, defaults are used.

### `scan <path>`

Full directory scan. `<path>` must be a directory.

| Flag | Description |
|---|---|
| `--format <text\|json>` | Output format. Default: `text`. |
| `--lang <langs>` | Enable only listed languages for this invocation. |
| `--parallel [n]` | Parallel worker count. Bare flag uses `runtime.NumCPU()`. |
| `--use-precedents` | Load `<path>/precedents.json` and suppress findings with matching `allow` decisions. |
| `--create-beads` | Create problem beads for error findings. |
| `--bead-threshold <n>` | Only create beads for grouped error findings with count > n. Default `0`. |
| `--config <path>` | Config file path. |

Exit codes: `0` no error findings, `1` error findings, `2` tool error.

### `check <file>`

Scan one file. `<file>` must be a file (directories rejected). Builds engine rooted at `dirname(file)`.

| Flag | Description |
|---|---|
| `--config <path>` | Config file path. |

### `watch <path>`

Watch directory for file changes and rescan changed lines. Uses fsnotify with 100ms debounce. Exits on Ctrl+C with `1` if any error finding seen during session, else `0`.

| Flag | Description |
|---|---|
| `--config <path>` | Config file path. |

### `report <path>`

Scan directory and write JSON report file.

| Flag | Description |
|---|---|
| `--output <file>` | Output file path. Default `truthsayer-report.json`. |
| `--config <path>` | Config file path. |

### `rules`

List rule metadata.

| Flag | Description |
|---|---|
| `--enabled` | List only enabled rules; applies config disables/severity overrides. |
| `--lang <langs>` | Filter by language alias. |
| `--config <path>` | Config file path (used when `--enabled` is set). |

Output columns: `ID`, `SEVERITY`, `FILES`, `DESCRIPTION`.

### `doctor`

Installation/readiness diagnostics. Checks version, Go toolchain, config load, rule counts, tree-sitter parser availability, and local file counts.

| Flag | Description |
|---|---|
| `--config <path>` | Config file path. |

### `judge <findings.json>`

Classify findings as `guilty|not_guilty|advisory` using precedents + LLM fallback.

Requires `ANTHROPIC_API_KEY` for LLM judgments.

| Flag | Description |
|---|---|
| `--format <json\|text>` | Output format. Default `json`. |
| `--precedents <path>` | Precedent store path. Default `<input-dir>/precedents.json`. |
| `--debt <path>` | Advisory debt path. Default `<input-dir>/.truthsayer-debt.json`. |
| `--law-candidates <path>` | Law candidate log path. |
| `--law-updates <path>` | Markdown proposal path. Default `<input-dir>/law-updates.md`. |
| `--law-threshold <n>` | Repeated-pattern threshold for proposals. Default `10`. |
| `--metrics <path>` | JSONL cost metrics path. Default `<input-dir>/.truthsayer-cost.jsonl`. |
| `--budget <usd>` | Spend cap. Default `0` (unlimited). |
| `--min-confidence <0..1>` | Minimum precedent confidence included in prompt context. Default `0`. |
| `--auto-apply-threshold <0..1>` | Skip LLM if strongest precedent confidence is above threshold. Default `0.9`. |

### `debt [path]`

List advisory debt entries. If `[path]` is omitted, reads `.truthsayer-debt.json`; if a directory, reads `<path>/.truthsayer-debt.json`.

| Flag | Description |
|---|---|
| `--format <text\|json>` | Output format. Default `text`. |

### `senate parse <file>`

Parse and validate a Senate verdict payload. Accepts JSON object or fenced ` ```json ` block; outputs normalized JSON.

### `senate apply <file> [repo]`

Apply approved amendments from a Senate verdict. `[repo]` defaults to `.`. Only applies when `status=approved`. Writes/updates `<repo>/.truthsayer-amendments.json` and appends audit JSONL to `<repo>/.truthsayer-amendments.audit.jsonl`.

### `warmup <path> [judge options...]`

Run full-repo scan + judge to seed precedent base. Forces `--format text` and `--precedents <path>/precedents.json`.

| Flag | Description |
|---|---|
| `--config <path>` | Config file path. |

Additional args are forwarded to `judge`.

### `hook <path>`

Pre-commit gate over staged files. `<path>` must be a git repo root. Reads staged files via `git diff --cached --name-only --diff-filter=ACM`. Runs scan; optionally gates via full judge pipeline with repo precedents. Falls back to deterministic scan gate if judgment unavailable.

| Flag | Description |
|---|---|
| `--config <path>` | Config file path. |

### `hook install <path>`

Install git pre-commit hook script. `<path>` must contain `.git/hooks`. Writes executable `.git/hooks/pre-commit` that runs `truthsayer hook .`.

### `ci [path]`

CI gate scanning only changed lines. `[path]` defaults to `.`.

Changed-line base resolution order:
1. `origin/$GITHUB_BASE_REF...HEAD` (if env set and ref exists)
2. `$GITHUB_EVENT_BEFORE...HEAD` (if set, non-zero hash, and ref exists)
3. `HEAD~1...HEAD` fallback

| Flag | Description |
|---|---|
| `--bead-threshold <n>` | Same semantics as `scan`; default `0`. |
| `--create-beads` | Accepted as compatibility no-op (CI always creates beads). |
| `--config <path>` | Config file path. |

### `ci init <path>`

Generate GitHub Actions workflow. `<path>` must be a git repo. Writes `.github/workflows/truthsayer.yml`.

### `version`

Print build version (`dev` if unset). Also accessible as `truthsayer --version` or `truthsayer -v`.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No error-severity findings / successful non-gating command |
| 1 | Error-severity findings present or guilty judgment |
| 2 | Tool error (bad args, missing files, invalid config) |

## Configuration

### `.truthsayer.toml`

```toml
[scan]
exclude_dirs = ["vendor", "node_modules", "testdata", "__pycache__", ".venv", "dist", "build"]
exclude_patterns = ["*_generated.go", "*.pb.go", "*.min.js", "*.bundle.js", "*.pyc"]

[scan.languages]
# Per-language enable/disable (all enabled by default when unset)
go = true
javascript = true   # .js, .jsx, .mjs, .cjs
typescript = true   # .ts, .tsx
python = true       # .py, .pyi
bash = true         # .sh, .bash
rust = true         # .rs

[rules.disable]
ids = ["bad-defaults.magic-number"]

[rules]
# Override severity for a specific rule
"trace-gaps.long-function-no-log" = "error"
```

### Config Files and Persisted Data

| Path | Purpose |
|---|---|
| `.truthsayer.toml` | Primary config (auto-loaded unless `--config` supplied). |
| `precedents.json` | Precedent decision store. |
| `.truthsayer-debt.json` | Advisory debt store. |
| `.truthsayer-law-candidates.json` | Repeated-ruling candidate store. |
| `law-updates.md` | Generated law proposal markdown. |
| `.truthsayer-cost.jsonl` | LLM cost metrics log. |
| `.truthsayer-amendments.json` | Applied Senate amendments store. |
| `.truthsayer-amendments.audit.jsonl` | Append-only amendment audit trail. |
| `truthsayer-report.json` | Default report output file. |
| `.github/workflows/truthsayer.yml` | CI workflow generated by `ci init`. |

### Environment Variables

| Var | Effect |
|---|---|
| `ANTHROPIC_API_KEY` | Required for `judge` LLM calls. |
| `GITHUB_BASE_REF` | CI diff-base selection candidate. |
| `GITHUB_EVENT_BEFORE` | CI diff-base fallback. |

## Outbound Network

`judge` calls `POST https://api.anthropic.com/v1/messages` using your `ANTHROPIC_API_KEY`. No other outbound HTTP. No inbound endpoints.

## Precedents (File-Based)

Truthsayer can store past violation decisions in `precedents.json` so future scans can reuse prior judgments.
Enable it during scans with `--use-precedents`.

Schema:

```json
[
  {
    "rule_id": "error-context.http-200-on-error",
    "violation_hash": "f8d31f4d8e7c...",
    "decision": "deny",
    "rationale": "Returning 200 in catch hides failures from clients.",
    "created_at": "2026-02-19T00:00:00Z"
  }
]
```

- `rule_id`: Truthsayer rule identifier
- `violation_hash`: stable hash for a specific violation instance
- `decision`: `allow` or `deny`
- `rationale`: human explanation for the decision
- `created_at`: RFC3339 timestamp

When `--use-precedents` is enabled, findings with matching `rule_id` + `violation_hash`
and decision `allow` are suppressed from scan output.

## CI Integration

```bash
# Install pre-commit hook
truthsayer hook install .

# Generate GitHub Actions workflow
truthsayer ci init .

# Use as quality gate (exits 1 on errors)
truthsayer scan .
```

## Dependencies

- Go `1.25.0`
- C compiler (gcc or clang) for tree-sitter cgo parsers
- Runtime optional: `git` (for `hook` and `ci`), `br` (for `--create-beads`)

## Part of Polis

Truthsayer is one tool in a larger system. Here's how the pieces connect:

| Tool | What it does | Repo |
|------|-------------|------|
| **Ergon** | Work orchestration | [ergon-work-orchestration](https://github.com/Perttulands/ergon-work-orchestration) |
| **Hermes** | Message relay | [hermes-relay](https://github.com/Perttulands/hermes-relay) |
| **Cerberus** | Access gate | [cerberus-gate](https://github.com/Perttulands/cerberus-gate) |
| **Chiron** | Agent trainer | [chiron-trainer](https://github.com/Perttulands/chiron-trainer) |
| **Learning Loop** | Feedback system | [learning-loop](https://github.com/Perttulands/learning-loop) |
| **Senate** | Governance | [senate](https://github.com/Perttulands/senate) |
| **Beads** | Problem tracking | [beads-polis](https://github.com/Perttulands/beads-polis) |
| **Truthsayer** | Code truth scanner | [truthsayer](https://github.com/Perttulands/truthsayer) |
| **UBS** | Bug scanner | [ultimate_bug_scanner](https://github.com/Perttulands/ultimate_bug_scanner) |
| **Oathkeeper** | Promise enforcement | [horkos-oathkeeper](https://github.com/Perttulands/horkos-oathkeeper) |
| **Argus** | Server watcher | [argus-watcher](https://github.com/Perttulands/argus-watcher) |
| **Utils** | Shared utilities | [polis-utils](https://github.com/Perttulands/polis-utils) |

## Part of the Agora

Aletheia was forged in **[Athena's Agora](https://github.com/Perttulands/athena-workspace)** — an autonomous coding system where AI agents build software and a figure in dark robes makes sure none of it is lying.

[Argus](https://github.com/Perttulands/argus-watcher) watches the server. [Oathkeeper](https://github.com/Perttulands/horkos-oathkeeper) watches the promises. [Relay](https://github.com/Perttulands/hermes-relay) carries the messages. Aletheia watches the code. Between the four of them, your silent failures have nowhere to hide.

The [mythology](https://github.com/Perttulands/athena-workspace/blob/main/mythology.md) has the full story.

## License

MIT
