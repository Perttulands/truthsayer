# Truthsayer

Development anti-pattern scanner for Go and bash codebases. Detects hidden failures, swallowed errors, bad defaults, mock leakage, missing traces, and configuration smells.

## Install

```bash
go install github.com/perttulands/truthsayer/cmd/truthsayer@latest
```

Or build from source:

```bash
git clone https://github.com/Perttulands/truthsayer.git
cd truthsayer
go build -o truthsayer ./cmd/truthsayer
```

## Usage

```bash
# Scan a directory
truthsayer scan .

# Scan a single file
truthsayer check path/to/file.go

# Watch for changes
truthsayer watch .

# JSON report
truthsayer scan --format json .

# List rules
truthsayer rules

# Check installation
truthsayer doctor
```

## What It Detects

24 rules across 6 categories:

| Category | What it catches |
|----------|----------------|
| **silent-fallback** | Swallowed errors, ignored return values, bare returns on error |
| **error-context** | Generic error messages, unwrapped errors, HTTP 200 on error |
| **trace-gaps** | Functions without logging, missing request IDs, no stderr capture |
| **mock-leakage** | Test imports in production, debug guards, fixture references |
| **bad-defaults** | Missing timeouts, no pipefail, magic numbers, unvalidated env vars |
| **config-smells** | Hardcoded paths, secrets in config, missing .gitignore entries |

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
exclude_dirs = ["vendor", "node_modules", "testdata"]
exclude_patterns = ["*_generated.go", "*.pb.go"]

[rules.disable]
ids = ["bad-defaults.magic-number"]

[rules]
# Promote a warning to error
"trace-gaps.long-function-no-log" = "error"
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No error-severity findings |
| 1 | Error-severity findings present |
| 2 | Tool error (bad config, invalid path) |

## License

MIT
