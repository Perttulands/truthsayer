# 🔍 Truthsayer — The Law Keeper

![Banner](banner.jpg)


_It sees the bugs your linter pretends don't exist._

---

Every codebase has a layer of lies. The swallowed exception that silently corrupts data three hours later. The test that passes because it's mocking the thing it's supposed to test. The fallback that "handles" errors by pretending they didn't happen. Your linter sees valid syntax and moves on. Truthsayer sees the truth underneath.

Truthsayer is a multi-language anti-pattern scanner that detects **failure-hiding patterns** — the kind of bugs that are technically valid code but practically ticking time bombs. It uses tree-sitter AST parsing for deep analysis and regex patterns for the stuff that doesn't need a parse tree.

**88 rules. 5 languages. Zero tolerance for code that lies about being fine.**

## What It Catches

| Category | Examples |
|----------|---------|
| **silent-fallback** | Swallowed errors, empty catch blocks, bare except, floating promises, ignored callbacks |
| **error-context** | Generic error messages, unwrapped errors, HTTP 200 on error, raise-from-none |
| **trace-gaps** | Functions without logging, missing request IDs, no unhandled rejection handler |
| **mock-leakage** | Test imports in production, debug guards, jest.mock in source, pytest fixtures in source |
| **bad-defaults** | Missing timeouts, no pipefail, eval usage, mutable default args, star imports |
| **config-smells** | Hardcoded paths, secrets in config, missing .gitignore, unpinned requirements |
| **test-isolation** | Missing cleanup in beforeAll/afterAll, test-only imports in source |

## Supported Languages

| Language | Parser | Rules |
|----------|--------|-------|
| Go | `go/ast` (stdlib) | 21 AST + 12 regex |
| JavaScript/TypeScript | tree-sitter (cgo) | 20 AST + 9 regex |
| Python | tree-sitter (cgo) | 19 AST + 7 regex |
| Bash | regex | shared |
| Config files | regex | shared |

## Install

Requires a C compiler (gcc or clang) for tree-sitter parsing.

```bash
go install github.com/Perttulands/truthsayer/cmd/truthsayer@latest
```

Or build from source:

```bash
git clone https://github.com/Perttulands/truthsayer.git
cd truthsayer
go build -o truthsayer ./cmd/truthsayer
```

On Debian/Ubuntu:
```bash
sudo apt-get install build-essential
```

## Usage

```bash
# Scan a directory (all languages)
truthsayer scan .

# Specific languages
truthsayer scan --lang go,python .

# Single file
truthsayer check path/to/file.go

# Watch mode
truthsayer watch .

# JSON output
truthsayer scan --format json .

# List all rules
truthsayer rules

# Rules for one language
truthsayer rules --lang python

# Verify installation
truthsayer doctor
```

### Language Aliases

| Alias | Extensions |
|-------|-----------|
| `go` | `.go` |
| `js`, `javascript` | `.js`, `.jsx`, `.mjs`, `.cjs` |
| `ts`, `typescript` | `.ts`, `.tsx` |
| `python`, `py` | `.py`, `.pyi` |
| `bash`, `shell`, `sh` | `.sh`, `.bash` |

## CI Integration

```bash
# Pre-commit hook
truthsayer hook install .

# GitHub Actions workflow
truthsayer ci init .

# Quality gate (exits 1 on errors)
truthsayer scan .
```

## Configuration

Create `.truthsayer.toml` in your project root:

```toml
[scan]
exclude_dirs = ["vendor", "node_modules", "testdata", "__pycache__", ".venv"]
exclude_patterns = ["*_generated.go", "*.pb.go", "*.min.js"]

[scan.languages]
go = true
javascript = true
typescript = true
python = true
bash = true

[rules.disable]
ids = ["bad-defaults.magic-number"]

[rules]
"trace-gaps.long-function-no-log" = "error"
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No error-severity findings |
| 1 | Error-severity findings present |
| 2 | Tool error (bad config, invalid path) |

## Part of [Athena's Agora](https://github.com/Perttulands/athena-workspace)

Truthsayer is part of the quality gate — it runs before Centurion to catch the things that tests can't. See the [mythology](https://github.com/Perttulands/athena-workspace/blob/main/mythology.md) for the full story.

## License

MIT
