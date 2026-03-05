# Aletheia the Truthsayer

![Aletheia Banner](banner.png)

*Bare feet. Linen. The Codex under one arm. She reads your code like a confession.*

---

Every codebase has a layer of polite lies. Error handlers that handle nothing. Catch blocks that catch and release. Timeouts that were going to be added "later." Somewhere under the clean abstractions, something is swallowing failures and smiling about it.

Aletheia sees both layers. She wears undyed linen — plain, undecorated, what's left when everything ornamental is stripped away. A dark stole across one shoulder with every rule stitched in gold thread. Eighty-eight of them. The Codex hangs at her side, heavy enough to use as a weapon if the Red Quill doesn't make the point clearly enough. And the bare feet — Pseudos, the false copy Prometheus's apprentice made, had no feet. Hers are the proof she's real.

She doesn't care about style. She doesn't care about naming conventions. She cares about the specific moment your code decided to lie about what happened. That `catch (e) {}` you wrote at 2am? She found it. She's been waiting.

Development anti-pattern scanner for Go, JavaScript/TypeScript, Python, and bash codebases. Detects hidden failures, swallowed errors, bad defaults, mock leakage, missing traces, and configuration smells.

Aletheia's niche is *failure-hiding anti-patterns* — the silent fallbacks, swallowed exceptions, missing observability, and production-test boundary violations that generic linters ignore.

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

## Supported Languages

| Language | Parser | Rule Count |
|----------|--------|-----------|
| Go | `go/ast` (stdlib) | 21 AST + 12 regex |
| JavaScript/TypeScript | tree-sitter (cgo) | 20 AST + 9 regex |
| Python | tree-sitter (cgo) | 19 AST + 7 regex |
| Bash/Config | regex | (shared with above) |

**Total: 88 rules** across 7 categories.

## What It Detects

| Category | What it catches |
|----------|----------------|
| **silent-fallback** | Swallowed errors, empty catch blocks, bare except, floating promises, ignored callbacks |
| **error-context** | Generic error messages, unwrapped errors, HTTP 200 on error, raise-from-none |
| **trace-gaps** | Functions without logging, missing request IDs, no unhandled rejection handler |
| **mock-leakage** | Test imports in production, debug guards, jest.mock in source, pytest fixtures in source |
| **bad-defaults** | Missing timeouts, no pipefail, eval usage, mutable default args, star imports |
| **config-smells** | Hardcoded paths, secrets in config, missing .gitignore, unpinned requirements |
| **test-isolation** | Missing cleanup in beforeAll/afterAll, test-only imports in source |

## CI Integration

```bash
# Install pre-commit hook
truthsayer hook install .

# Generate GitHub Actions workflow
truthsayer ci init .

# Use as quality gate (exits 1 on errors)
truthsayer scan .
```

## Configuration

Create `.truthsayer.toml` in your project root:

```toml
[scan]
exclude_dirs = ["vendor", "node_modules", "testdata", "__pycache__", ".venv", "dist", "build"]
exclude_patterns = ["*_generated.go", "*.pb.go", "*.min.js", "*.bundle.js", "*.pyc"]

[scan.languages]
# Per-language enable/disable (all enabled by default)
go = true
javascript = true   # .js, .jsx, .mjs, .cjs
typescript = true   # .ts, .tsx
python = true       # .py, .pyi
bash = true          # .sh, .bash

[rules.disable]
ids = ["bad-defaults.magic-number"]

[rules]
# Promote a warning to error
"trace-gaps.long-function-no-log" = "error"
```

Set `[scan.languages]` values to `false` to skip scanning for that language entirely. Omitted languages default to enabled.

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
- `created_at`: RFC3339 timestamp (UTC recommended)

When `--use-precedents` is enabled, findings with matching `rule_id` + `violation_hash`
and decision `allow` are suppressed from scan output.

Package: `internal/precedent`

```go
store := precedent.NewStore("precedents.json")

_ = store.Add(precedent.Precedent{
    RuleID:        "error-context.http-200-on-error",
    ViolationHash: "f8d31f4d8e7c...",
    Decision:      precedent.DecisionDeny,
    Rationale:     "Returning 200 in catch hides failures from clients.",
})

match, ok, err := store.Query("error-context.http-200-on-error", "f8d31f4d8e7c...")
_ = match
_ = ok
_ = err
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No error-severity findings |
| 1 | Error-severity findings present |
| 2 | Tool error (bad config, invalid path) |

## Dependencies

None. Standalone tool.

## Part of the Agora

Aletheia was forged in **[Athena's Agora](https://github.com/Perttulands/athena-workspace)** — an autonomous coding system where AI agents build software and a figure in dark robes makes sure none of it is lying.

[Argus](https://github.com/Perttulands/argus-watcher) watches the server. [Oathkeeper](https://github.com/Perttulands/horkos-oathkeeper) watches the promises. [Relay](https://github.com/Perttulands/hermes-relay) carries the messages. Aletheia watches the code. Between the four of them, your silent failures have nowhere to hide.

The [mythology](https://github.com/Perttulands/athena-workspace/blob/main/mythology.md) has the full story.

## License

MIT
