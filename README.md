# 🔍 Truthsayer

![Banner](banner.jpg)

*88 rules. No mercy. No exceptions.*

---

Every codebase is a crime scene. The developers are gone. The git blame points everywhere and nowhere. Somewhere in the wreckage, someone wrote `catch (Exception e) {}` and thought nobody would notice.

Truthsayer noticed.

He walks the Agora in dark robes stitched with golden text — all 88 rules, woven into the fabric. The law is literally part of him. A bronze monocle over one eye, a red quill in his hand, and a cracked mirror that shows two layers at once: the beautiful surface, and the rot underneath. Your linter sees valid syntax and moves on. Truthsayer sees the `except: pass` hiding behind it and starts writing in permanent ink.

It scans your code across 5 languages, matches against 88 anti-pattern rules that were written in blood (metaphorical — mostly), and tells you exactly where you're lying to yourself. Some of the rules are obvious. Some will hurt your feelings. All of them exist because someone shipped the thing they warn about and regretted it at 3am.

It's the friend who tells you there's spinach in your teeth. Except the spinach is an unchecked type assertion and your teeth are production.

## What It Catches

| Category | What you did wrong |
|----------|--------------------|
| **silent-fallback** | Swallowed errors, empty catch blocks, bare except, floating promises. The code equivalent of putting a rug over a hole in the floor. |
| **error-context** | Generic error messages, unwrapped errors, HTTP 200 on error. "Something went wrong" is not a diagnostic. |
| **trace-gaps** | Functions without logging, missing request IDs. When it breaks at 3am, how exactly were you planning to debug it? |
| **mock-leakage** | Test imports in production, jest.mock in source. Congratulations, you're shipping your test harness. |
| **bad-defaults** | Missing timeouts, no pipefail, eval usage, mutable default args. The "works on my machine" starter pack. |
| **config-smells** | Hardcoded paths, secrets in config, unpinned requirements. Future you is going to be very upset with past you. |
| **test-isolation** | Missing cleanup, test-only imports in source. Your tests pass because they're lying, not because they're right. |

## Supported Languages

| Language | Parser | Rules |
|----------|--------|-------|
| Go | `go/ast` (stdlib) | 21 AST + 12 regex |
| JavaScript/TypeScript | tree-sitter (cgo) | 20 AST + 9 regex |
| Python | tree-sitter (cgo) | 19 AST + 7 regex |
| Bash | regex | shared |
| Config files | regex | shared |

## Install

```bash
# You need a C compiler. If that surprises you, tree-sitter has opinions about parsing.
sudo apt-get install build-essential  # Debian/Ubuntu

# Then:
go install github.com/Perttulands/truthsayer/cmd/truthsayer@latest

# Or from source:
git clone https://github.com/Perttulands/truthsayer.git
cd truthsayer
go build -o truthsayer ./cmd/truthsayer
```

## Usage

```bash
# The main event. Scan everything. Fear nothing.
truthsayer scan .

# Be selective about your suffering
truthsayer scan --lang go,python .

# Single file therapy session
truthsayer check path/to/file.go

# Watch mode — real-time judgment
truthsayer watch .

# JSON output for the machines
truthsayer scan --format json .

# Read the law before you break it
truthsayer rules

# Quality gate — exits 1 on errors. Put this in CI.
# Sleep well knowing nothing ships without passing 88 checks.
truthsayer scan .
```

## CI Integration

```bash
# Pre-commit hook (installs to .git/hooks/)
truthsayer hook install .

# GitHub Actions — generates workflow file
truthsayer ci init .
```

## Configuration

Create `.truthsayer.toml` if you want to adjust the rules. You can't disable them all. That's not how laws work.

```toml
[scan]
exclude_dirs = ["vendor", "node_modules", "testdata"]
exclude_patterns = ["*_generated.go", "*.pb.go"]

[rules.disable]
ids = ["bad-defaults.magic-number"]  # Fine, but we're judging you

[rules]
"trace-gaps.long-function-no-log" = "error"  # Promote to error severity
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Clean. Truthsayer respects you today. |
| 1 | Findings. The red quill has been busy. |
| 2 | Tool error. Even the law keeper has bad days. |

## For Agents

This repo includes `AGENTS.md` with operational instructions.

```bash
git clone https://github.com/Perttulands/truthsayer.git
cd truthsayer
go build -o ~/go/bin/truthsayer ./cmd/truthsayer
```

Dependencies: Go 1.21+, C compiler (`gcc`/`clang`) for tree-sitter. On Debian/Ubuntu: `sudo apt-get install build-essential`.

## Part of the Agora

Truthsayer was forged in **[Athena's Agora](https://github.com/Perttulands/athena-workspace)** — an autonomous coding system where AI agents build software under the watch of Greek mythology and cyberpunk engineering.

He's not alone. [Oathkeeper](https://github.com/Perttulands/oathkeeper) checks whether agents kept their promises. [Argus](https://github.com/Perttulands/argus) watches the server with one red eye that never closes. [Relay](https://github.com/Perttulands/relay) carries messages between agents at 26,000/sec with zero lost. Truthsayer watches the code. The red quill marks what the red eye misses.

There are others. The [mythology](https://github.com/Perttulands/athena-workspace/blob/main/mythology.md) has the full story.

## License

MIT
