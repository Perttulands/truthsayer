# Truthsayer Multi-Language Extension — Product Requirements Document

> Extending Truthsayer's anti-pattern detection from Go/bash to JavaScript/TypeScript and Python.

## 1. Overview

Truthsayer v1 scans Go source files via `go/ast` and bash/config files via regex patterns. This PRD defines the extension to JavaScript/TypeScript (JS/TS) and Python — two ecosystems with fundamentally different anti-pattern profiles and AST tooling.

The core insight: Truthsayer's architecture already separates concerns cleanly — `Rule` metadata, `ASTChecker`/`RegexChecker` interfaces, a `Registry` that dispatches by file extension, and an `Engine` that routes `.go` files to the Go scanner and everything else to regex. Extending to new languages means:

1. Adding new scanner backends (JS/TS AST, Python AST)
2. Defining language-specific checker interfaces
3. Writing rules that exploit each language's unique anti-pattern surface
4. Keeping the engine, registry, config, and CLI largely unchanged

The goal is **not** to reimplement ESLint or pylint. Truthsayer's niche is *failure-hiding anti-patterns* — the silent fallbacks, swallowed exceptions, missing observability, and production-test boundary violations that generic linters ignore.

## 2. Why These Languages

**JavaScript/TypeScript**: The most common language for web services. Its async model (Promises, async/await) creates an entirely unique class of silent failures: unhandled rejections, floating promises, swallowed `.catch()` blocks, and callback error parameters silently ignored. TypeScript adds type assertion anti-patterns (`as any`, non-null assertions) that defeat the type system's safety guarantees.

**Python**: Dominant in data pipelines, ML infrastructure, and automation scripts — exactly the domains where silent failures cause the most damage. Python's `except: pass`, bare `except`, and `# type: ignore` comments are textbook Truthsayer targets. Its dynamic typing means anti-patterns that would be compile-time errors in Go (wrong return types, missing attributes) become runtime silent failures.

## 3. Architecture

### 3.1 Current State

```
Engine.Scan(file)
  ├── ext == ".go"  → GoScanner (go/ast) + RegexScanner
  └── ext == other  → RegexScanner only

Registry
  ├── ASTChecker    — Go AST rules only
  └── RegexChecker  — line-based rules (bash, config, any file type)
```

### 3.2 Target State

```
Engine.Scan(file)
  ├── ext == ".go"              → GoScanner (go/ast) + RegexScanner
  ├── ext == ".js/.ts/.jsx/…"   → JSScanner (external AST) + RegexScanner
  ├── ext == ".py"              → PyScanner (external AST) + RegexScanner
  └── ext == other              → RegexScanner only

Registry
  ├── ASTChecker      — Go AST rules (unchanged)
  ├── JSASTChecker    — JS/TS AST rules (new)
  ├── PyASTChecker    — Python AST rules (new)
  └── RegexChecker    — line-based rules (all languages)
```

### 3.3 AST Strategy Per Language

#### Go (existing) — Native AST, zero dependencies

Go's `go/ast` is stdlib. Parsing is fast, the AST is stable, and Truthsayer links against it directly. No changes needed.

#### JavaScript/TypeScript — Tree-sitter via Go bindings

**Why not a JS parser in Go?** No production-quality JS/TS parser exists in pure Go. The ecosystem relies on Babel, TypeScript compiler, or SWC — all written in JS/Rust.

**Approach**: Use [tree-sitter](https://tree-sitter.github.io/) via the Go bindings ([go-tree-sitter](https://github.com/tree-sitter/tree-sitter/tree/master/lib/binding_go) or [smacker/go-tree-sitter](https://github.com/smacker/go-tree-sitter)). Tree-sitter is a C library with Go bindings that provides incremental, error-tolerant parsing for 100+ languages including JavaScript, TypeScript, TSX, and JSX.

Benefits:
- **Single binary**: tree-sitter compiles into the Go binary via cgo. No Node.js runtime required.
- **Error-tolerant**: Parses syntactically broken files without crashing — critical for scanning WIP code.
- **Fast**: C-speed parsing, handles 10k LOC files in milliseconds.
- **Unified**: Same library handles JS, TS, JSX, TSX with different grammar modules.

Trade-offs:
- **cgo dependency**: Build requires C compiler. Acceptable — tree-sitter is well-maintained and cross-compiles cleanly.
- **AST shape differs from go/ast**: JS/TS rules work with tree-sitter's generic `Node` type (navigated by node kind strings) rather than Go's typed AST. Rules use tree-sitter queries (S-expression pattern matching) or manual traversal.

**Alternative considered**: Shell out to a Node.js-based parser (e.g., `npx @babel/parser`). Rejected — adds a runtime dependency, slower, defeats single-binary goal.

#### Python — Tree-sitter via same Go bindings

Python AST parsing follows the same tree-sitter strategy. The `tree-sitter-python` grammar is mature and handles Python 3.6–3.13 syntax.

**Why not Python's `ast` module via subprocess?** Requires Python runtime on the host. Version-dependent (Python 3.8 `ast` differs from 3.12). Slow for bulk scanning. Tree-sitter keeps the single-binary promise.

### 3.4 New Interfaces

```go
// JSASTChecker is implemented by rules that analyze JS/TS AST nodes.
type JSASTChecker interface {
    Meta() Rule
    CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding
}

// PyASTChecker is implemented by rules that analyze Python AST nodes.
type PyASTChecker interface {
    Meta() Rule
    CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding
}
```

Both interfaces receive the tree-sitter `Tree` and raw source bytes (for snippet extraction). The `path` argument enables test-file detection (e.g., skip rules in `*_test.py`, `*.spec.ts`).

### 3.5 Scanner Implementations

```go
// internal/scanner/js_scanner.go
type JSScanner struct {
    parser   *sitter.Parser
    checkers []rules.JSASTChecker
}

func (s *JSScanner) Scan(path string) ([]finding.Finding, []string, error) {
    source, _ := os.ReadFile(path)
    tree, _ := s.parser.ParseCtx(ctx, nil, source)
    defer tree.Close()

    var findings []finding.Finding
    for _, checker := range s.checkers {
        findings = append(findings, checker.CheckJSAST(tree, source, path)...)
    }
    lines := strings.Split(string(source), "\n")
    return findings, lines, nil
}
```

```go
// internal/scanner/py_scanner.go
type PyScanner struct {
    parser   *sitter.Parser
    checkers []rules.PyASTChecker
}
// Same pattern as JSScanner with CheckPyAST dispatch.
```

### 3.6 Engine Routing Changes

```go
// internal/engine/engine.go — extended routing
func (e *Engine) scanFile(path string) []finding.Finding {
    ext := filepath.Ext(path)
    var findings []finding.Finding

    switch {
    case ext == ".go":
        results, lines, _ := e.goScanner.Scan(path)
        findings = append(findings, results...)
        findings = append(findings, e.regexScanner.ScanLines(path, lines)...)

    case isJSExt(ext): // .js, .jsx, .ts, .tsx, .mjs, .cjs
        results, lines, _ := e.jsScanner.Scan(path)
        findings = append(findings, results...)
        findings = append(findings, e.regexScanner.ScanLines(path, lines)...)

    case ext == ".py":
        results, lines, _ := e.pyScanner.Scan(path)
        findings = append(findings, results...)
        findings = append(findings, e.regexScanner.ScanLines(path, lines)...)

    default:
        results, _ := e.regexScanner.Scan(path)
        findings = append(findings, results...)
    }
    return findings
}
```

### 3.7 Registry Extensions

```go
func (r *Registry) RegisterJSAST(c JSASTChecker)  { ... }
func (r *Registry) RegisterPyAST(c PyASTChecker)  { ... }
func (r *Registry) JSASTCheckers() []JSASTChecker  { ... }
func (r *Registry) PyASTCheckers() []PyASTChecker  { ... }
```

`DefaultRegistry()` gains new blocks:

```go
// JS/TS AST rules
reg.RegisterJSAST(&EmptyCatchBlock{})
reg.RegisterJSAST(&FloatingPromise{})
// ...

// Python AST rules
reg.RegisterPyAST(&BareExcept{})
reg.RegisterPyAST(&ExceptPass{})
// ...
```

## 4. JavaScript/TypeScript Rules

JS/TS anti-patterns fall into the same 6 categories but with language-specific manifestations. The async execution model is the biggest differentiator — promises create an entire failure-hiding surface that doesn't exist in Go.

### 4.1 Silent Fallbacks

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `silent-fallback.empty-catch` | error | AST | `catch (e) {}` or `catch (e) { /* empty */ }` — error silently swallowed |
| `silent-fallback.catch-return-null` | error | AST | `.catch(() => null)` or `.catch(() => undefined)` — promise error masked |
| `silent-fallback.floating-promise` | error | AST | Promise-returning call without `await`, `.then()`, or `.catch()` — rejection silently lost |
| `silent-fallback.callback-err-ignored` | warning | AST | Callback pattern `(err, data) => { ... }` where `err` parameter is never referenced |
| `silent-fallback.optional-chain-silence` | info | AST | Deep optional chaining `a?.b?.c?.d?.e` (>3 levels) — masks structural failures as undefined |

**What makes these unique**: Go forces `if err != nil` at every call site. JS/TS has three separate error channels (throw/catch, Promise rejection, callback err), and all three can be silently ignored. The `floating-promise` rule alone catches more production incidents than any other single anti-pattern.

### 4.2 Missing Error Context

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `error-context.rethrow-no-wrap` | warning | AST | `catch (e) { throw e }` without adding context — stack trace preserved but no business context |
| `error-context.generic-error-message` | warning | AST | `throw new Error("failed")` or `throw new Error("error")` — no variable interpolation |
| `error-context.promise-reject-non-error` | warning | AST | `Promise.reject("string")` or `reject(42)` — non-Error rejections lack stack traces |
| `error-context.console-error-no-throw` | warning | AST | `console.error(...)` in catch block without re-throw — error logged but swallowed |
| `error-context.http-200-on-error` | error | AST | Express/Koa/Fastify handler: `res.status(200)` or `res.json(...)` after error detection |

### 4.3 Trace & Log Gaps

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `trace-gaps.no-error-handler-express` | warning | AST | Express app without `app.use((err, req, res, next) => ...)` error middleware |
| `trace-gaps.no-unhandled-rejection` | warning | Regex | No `process.on('unhandledRejection', ...)` in entry point files |
| `trace-gaps.console-log-in-production` | info | Regex | `console.log` used for debugging in non-test source — should use structured logger |
| `trace-gaps.missing-correlation-id` | warning | AST | HTTP middleware chain without correlation/request ID propagation |

### 4.4 Mock Leakage

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `mock-leakage.jest-mock-in-src` | error | Regex | `jest.mock(`, `jest.fn(`, `jest.spyOn(` in non-test files |
| `mock-leakage.test-import-in-src` | error | AST | Import from `@testing-library/*`, `vitest`, `jest` in production source |
| `mock-leakage.env-test-check` | warning | AST | `process.env.NODE_ENV === 'test'` guard in production code |
| `mock-leakage.storybook-in-src` | info | Regex | `.stories.` import or Storybook decorator in non-story file |

### 4.5 Bad Defaults

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `bad-defaults.no-timeout-fetch` | error | AST | `fetch()` without `AbortController` / `signal` — can hang indefinitely |
| `bad-defaults.any-type-assertion` | warning | AST (TS) | `as any` type assertion — defeats TypeScript's type safety |
| `bad-defaults.non-null-assertion` | info | AST (TS) | `variable!` non-null assertion — runtime null possible despite compile-time override |
| `bad-defaults.ts-ignore` | warning | Regex | `@ts-ignore` without explanation comment — hides type errors |
| `bad-defaults.eslint-disable-no-reason` | info | Regex | `eslint-disable` without comment justification |
| `bad-defaults.eval-usage` | error | AST | `eval()` or `new Function()` in production code — security and debuggability risk |
| `bad-defaults.no-strict-mode` | info | Regex | CommonJS file without `'use strict'` (only `.cjs` files, ESM is strict by default) |

### 4.6 Configuration Smells

Existing regex rules (`secret-in-config`, `hardcoded-path`, `missing-gitignore`) already apply to JS/TS config files (`.json`, `.yaml`, `.env`). New JS-specific additions:

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `config-smells.hardcoded-api-url` | warning | Regex | Hardcoded `http://localhost`, `https://api.` URLs in source (not config) files |
| `config-smells.dotenv-no-example` | info | Regex | `.env` file exists but no `.env.example` — team members don't know required vars |

### 4.7 Test Isolation (existing category, extended)

The existing `test-isolation.*` regex rules already target JS/TS test files. AST-based versions can be more precise:

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `test-isolation.no-afterall-cleanup` | warning | AST | `beforeAll`/`beforeEach` creating resources without matching `afterAll`/`afterEach` (AST version of existing regex rule) |
| `test-isolation.test-only-import` | info | AST | `import { ... } from '...'` where the module is exclusively used in test setup but lives in `src/` |

## 5. Python Rules

Python's anti-pattern surface is shaped by its dynamic typing, its exception model (bare `except`, `except Exception`), and its scripting heritage (scripts that silently continue after failures).

### 5.1 Silent Fallbacks

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `silent-fallback.bare-except` | error | AST | `except:` without specifying exception type — catches SystemExit, KeyboardInterrupt, everything |
| `silent-fallback.except-pass` | error | AST | `except SomeError: pass` — exception explicitly silenced |
| `silent-fallback.except-broad` | warning | AST | `except Exception:` — overly broad, catches too much |
| `silent-fallback.subprocess-no-check` | warning | AST | `subprocess.run(...)` or `subprocess.call(...)` without `check=True` — exit code silently ignored |
| `silent-fallback.getattr-silent-default` | info | AST | `getattr(obj, 'attr', None)` in contexts where missing attribute indicates a bug, not optional feature |
| `silent-fallback.dict-get-none` | info | AST | `dict.get(key)` without explicit default where None would cause downstream failure |

**What makes these unique**: Python's `except:` catches *everything*, including `SystemExit` and `KeyboardInterrupt`. This is far more dangerous than Go's explicit error returns — a `except: pass` can literally prevent a program from being killed. The `subprocess` rules are Python's equivalent of bash's missing `set -e`.

### 5.2 Missing Error Context

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `error-context.raise-from-none` | warning | AST | `raise NewError() from None` — explicitly discards exception chain |
| `error-context.bare-raise-different` | warning | AST | `except ErrorA: raise ErrorB()` without `from` — original traceback lost |
| `error-context.generic-exception` | warning | AST | `raise Exception("failed")` — base Exception without specific type |
| `error-context.string-exception` | error | AST | `raise "error message"` — Python 2 idiom that's a TypeError in Python 3 |
| `error-context.log-and-raise` | info | AST | `logging.error(e); raise` — duplicate error reporting (once in log, once in traceback) |

### 5.3 Trace & Log Gaps

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `trace-gaps.print-debug` | info | Regex | `print()` used for debugging in non-test/non-script files — should use `logging` |
| `trace-gaps.no-logging-config` | warning | Regex | Python package with no `logging.basicConfig()` or `logging.getLogger()` in entry points |
| `trace-gaps.silent-request` | warning | AST | `requests.get/post()` without response status check (`response.raise_for_status()` or status_code check) |

### 5.4 Mock Leakage

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `mock-leakage.unittest-import` | error | AST | `from unittest.mock import ...` or `import mock` in non-test file |
| `mock-leakage.pytest-fixture-in-src` | error | Regex | `@pytest.fixture` decorator in non-test file |
| `mock-leakage.debug-flag` | warning | AST | `if __debug__:` or `if DEBUG:` guard in production code with side effects |

### 5.5 Bad Defaults

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `bad-defaults.mutable-default-arg` | error | AST | `def f(items=[])` — mutable default argument shared across calls |
| `bad-defaults.no-timeout-requests` | error | AST | `requests.get/post()` without `timeout=` parameter — blocks indefinitely |
| `bad-defaults.star-import` | warning | AST | `from module import *` — namespace pollution, hides dependencies |
| `bad-defaults.type-ignore-bare` | warning | Regex | `# type: ignore` without specific error code — blanket type suppression |
| `bad-defaults.noqa-bare` | info | Regex | `# noqa` without specific code — blanket linter suppression |
| `bad-defaults.global-state` | warning | AST | Module-level mutable global (`ITEMS = []`, `CACHE = {}`) without documentation |
| `bad-defaults.no-encoding-open` | info | AST | `open(file)` without explicit `encoding=` parameter — platform-dependent encoding |

### 5.6 Configuration Smells

Existing regex rules cover `.yaml`, `.json`, `.env`. Python-specific additions:

| ID | Severity | Scan Type | Description |
|---|---|---|---|
| `config-smells.hardcoded-credentials-py` | error | Regex | `password = "..."`, `api_key = "..."`, `secret = "..."` string literals in Python source |
| `config-smells.requirements-unpinned` | warning | Regex | `requirements.txt` entries without version pins (`==`) — non-reproducible builds |

## 6. Configuration Extensions

### 6.1 TOML Config Changes

The existing `.truthsayer.toml` format extends naturally. No schema-breaking changes.

```toml
[scan]
# Extended extension list (auto-detected, but can be overridden)
exclude_dirs = ["vendor", "node_modules", "testdata", "__pycache__", ".venv", "dist", "build"]
exclude_patterns = [
    "*_generated.go", "*.pb.go",        # Go
    "*.min.js", "*.bundle.js",           # JS bundled output
    "*.pyc",                             # Python bytecode
]

# NEW: Max file size increase for JS bundles that sneak in
max_file_size = 2097152  # 2MB

[scan.languages]
# NEW: Per-language enable/disable (all enabled by default)
# Set to false to skip all rules for a language
go = true
javascript = true   # Covers .js, .jsx, .mjs, .cjs
typescript = true   # Covers .ts, .tsx
python = true
bash = true          # Covers .sh, .bash

[rules]
# Severity overrides work the same way — rule IDs are namespaced per language
# "silent-fallback.empty-catch" = "warning"        # JS rule
# "silent-fallback.bare-except" = "warning"         # Python rule

[rules.disable]
# Disable specific rules (same mechanism, new rule IDs)
# ids = ["bad-defaults.any-type-assertion", "trace-gaps.print-debug"]
```

### 6.2 Config Struct Changes

```go
type ScanConfig struct {
    ExcludeDirs     []string       `toml:"exclude_dirs"`
    ExcludePatterns []string       `toml:"exclude_patterns"`
    MaxFileSize     int64          `toml:"max_file_size"`
    Languages       LanguageConfig `toml:"languages"`
}

type LanguageConfig struct {
    Go         *bool `toml:"go"`          // default: true
    JavaScript *bool `toml:"javascript"`  // default: true
    TypeScript *bool `toml:"typescript"`  // default: true
    Python     *bool `toml:"python"`      // default: true
    Bash       *bool `toml:"bash"`        // default: true
}
```

Using `*bool` allows distinguishing "not set" (default true) from "explicitly false".

### 6.3 Walker Changes

The walker already supports JS/TS extensions in `supportedExts`. Add Python:

```go
var supportedExts = map[string]bool{
    // existing...
    ".py":  true,  // NEW
    ".pyi": true,  // NEW: Python type stubs
}
```

## 7. CLI Changes

### 7.1 No Breaking Changes

All existing commands work unchanged. The extensions are additive.

### 7.2 New Flags

```
truthsayer scan --lang go,python <path>     # Scan only specific languages
truthsayer scan --lang js,ts <path>         # JS/TS shorthand aliases
truthsayer rules --lang python              # List only Python rules
truthsayer rules --lang js                  # List only JS/TS rules
```

The `--lang` flag is a convenience filter. Without it, all enabled languages are scanned (backward compatible).

Language aliases:
- `go` → `.go`
- `js`, `javascript` → `.js`, `.jsx`, `.mjs`, `.cjs`
- `ts`, `typescript` → `.ts`, `.tsx`
- `python`, `py` → `.py`, `.pyi`
- `bash`, `shell`, `sh` → `.sh`, `.bash`

### 7.3 Doctor Command Extension

`truthsayer doctor` gains language-specific checks:

```
$ truthsayer doctor
✓ Config loaded from .truthsayer.toml
✓ 47 rules active (21 Go, 16 JS/TS, 10 Python)
✓ Go AST parser: available (stdlib)
✓ JS/TS AST parser: available (tree-sitter)
✓ Python AST parser: available (tree-sitter)
✓ All scanners operational
```

### 7.4 Rules Command Extension

```
$ truthsayer rules --lang python
ID                                          SEVERITY   DESCRIPTION
silent-fallback.bare-except                 error      except: without specifying exception type
silent-fallback.except-pass                 error      except SomeError: pass — exception silenced
silent-fallback.subprocess-no-check         warning    subprocess.run() without check=True
bad-defaults.mutable-default-arg            error      Mutable default argument def f(items=[])
bad-defaults.no-timeout-requests            error      requests.get() without timeout parameter
...
```

## 8. New Dependencies

| Dependency | Purpose | Impact |
|---|---|---|
| `github.com/smacker/go-tree-sitter` | Tree-sitter Go bindings | cgo; compiles C code into binary |
| `github.com/smacker/go-tree-sitter/javascript` | JS grammar | ~200KB compiled |
| `github.com/smacker/go-tree-sitter/typescript/typescript` | TS grammar | ~300KB compiled |
| `github.com/smacker/go-tree-sitter/typescript/tsx` | TSX grammar | ~300KB compiled |
| `github.com/smacker/go-tree-sitter/python` | Python grammar | ~200KB compiled |

**Binary size impact**: Estimated +2–3MB for tree-sitter core + grammars. Total binary stays under 20MB.

**Build impact**: Requires C compiler (gcc/clang). CI needs `build-essential` or equivalent. Cross-compilation possible via `zig cc` as C cross-compiler.

**Alternative**: If the cgo requirement is unacceptable, an alternative approach uses tree-sitter's WASM bindings via `wazero` (pure-Go WebAssembly runtime). This is 2–3× slower but avoids cgo entirely. Recommendation: start with cgo, evaluate WASM only if distribution becomes a problem.

## 9. Project Layout Changes

```
truthsayer/
├── internal/
│   ├── rules/
│   │   ├── rule.go                      # Extended with JSASTChecker, PyASTChecker
│   │   ├── registry.go                  # Extended with RegisterJSAST, RegisterPyAST
│   │   ├── js_empty_catch.go            # JS: silent-fallback.empty-catch
│   │   ├── js_floating_promise.go       # JS: silent-fallback.floating-promise
│   │   ├── js_catch_return_null.go      # JS: silent-fallback.catch-return-null
│   │   ├── js_no_timeout_fetch.go       # JS: bad-defaults.no-timeout-fetch
│   │   ├── js_any_assertion.go          # TS: bad-defaults.any-type-assertion
│   │   ├── js_promise_reject.go         # JS: error-context.promise-reject-non-error
│   │   ├── js_eval_usage.go             # JS: bad-defaults.eval-usage
│   │   ├── js_http_200_on_error.go      # JS: error-context.http-200-on-error
│   │   ├── js_rethrow_no_wrap.go        # JS: error-context.rethrow-no-wrap
│   │   ├── py_bare_except.go            # PY: silent-fallback.bare-except
│   │   ├── py_except_pass.go            # PY: silent-fallback.except-pass
│   │   ├── py_mutable_default.go        # PY: bad-defaults.mutable-default-arg
│   │   ├── py_no_timeout_requests.go    # PY: bad-defaults.no-timeout-requests
│   │   ├── py_subprocess_no_check.go    # PY: silent-fallback.subprocess-no-check
│   │   ├── py_star_import.go            # PY: bad-defaults.star-import
│   │   ├── py_raise_from_none.go        # PY: error-context.raise-from-none
│   │   ├── py_generic_exception.go      # PY: error-context.generic-exception
│   │   ├── ...                          # (remaining rule files)
│   │   └── (existing Go/bash rules)     # Unchanged
│   ├── scanner/
│   │   ├── go_scanner.go                # Unchanged
│   │   ├── js_scanner.go                # NEW: tree-sitter JS/TS scanner
│   │   ├── py_scanner.go                # NEW: tree-sitter Python scanner
│   │   ├── regex_scanner.go             # Unchanged
│   │   └── treesitter.go               # NEW: shared tree-sitter utilities
│   └── ...
├── testdata/
│   ├── go/                              # Existing
│   ├── bash/                            # Existing
│   ├── js/                              # NEW: JS/TS test fixtures
│   │   ├── empty_catch.js
│   │   ├── floating_promise.ts
│   │   ├── proper_error_handling.ts
│   │   ├── any_assertion.ts
│   │   ├── no_timeout_fetch.js
│   │   └── ...
│   ├── python/                          # NEW: Python test fixtures
│   │   ├── bare_except.py
│   │   ├── except_pass.py
│   │   ├── proper_exception.py
│   │   ├── mutable_default.py
│   │   ├── subprocess_no_check.py
│   │   └── ...
│   └── config/                          # Existing
└── ...
```

### File Naming Convention

Rule files follow the pattern `{lang}_{rule_name}.go`:
- `js_empty_catch.go` — JS/TS rule (JS AST checker)
- `py_bare_except.go` — Python rule (Python AST checker)
- No prefix — Go or cross-language (existing convention)

## 10. Shared Tree-sitter Utilities

Many operations are common across JS and Python rules. A shared utility package avoids duplication:

```go
// internal/scanner/treesitter.go

// FindNodesByType walks the tree and returns all nodes matching any of the given types.
func FindNodesByType(root *sitter.Node, types ...string) []*sitter.Node { ... }

// NodeText extracts the source text for a given node.
func NodeText(node *sitter.Node, source []byte) string { ... }

// LineNumber returns the 1-indexed line number for a node.
func LineNumber(node *sitter.Node) int { return int(node.StartPoint().Row) + 1 }

// SourceLine extracts a single source line from the source bytes.
func SourceLine(source []byte, line int) string { ... }

// HasChildOfType checks if a node has any direct child of the given type.
func HasChildOfType(node *sitter.Node, childType string) bool { ... }

// IsInsideFunction checks if a node is inside a function/method definition.
func IsInsideFunction(node *sitter.Node) bool { ... }

// IsTestFile determines if a path is a test file for JS/TS or Python.
func IsTestFile(path string) bool { ... }
```

## 11. Rule Implementation Examples

### 11.1 JS: Empty Catch Block (AST)

```go
// internal/rules/js_empty_catch.go
type EmptyCatchBlock struct{}

func (e *EmptyCatchBlock) Meta() Rule {
    return Rule{
        ID:       "silent-fallback.empty-catch",
        Category: "silent-fallback",
        Name:     "Empty catch block",
        Description: "catch block with no error handling — error silently swallowed",
        Severity: finding.SeverityError,
        FileTypes: []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
        ScanType: ScanTypeAST,
    }
}

func (e *EmptyCatchBlock) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
    var findings []finding.Finding
    // Find all catch_clause nodes
    catchNodes := scanner.FindNodesByType(tree.RootNode(), "catch_clause")
    for _, node := range catchNodes {
        body := node.ChildByFieldName("body")
        if body == nil || body.NamedChildCount() == 0 {
            // Empty catch body or only comments
            line := scanner.LineNumber(node)
            findings = append(findings, finding.Finding{
                Rule:       e.Meta().ID,
                Severity:   e.Meta().Severity,
                File:       path,
                Line:       line,
                Code:       scanner.SourceLine(source, line),
                Message:    "Empty catch block silently swallows error",
                Suggestion: "Handle the error: log it, rethrow it, or add a comment explaining why it's safe to ignore",
            })
        }
    }
    return findings
}
```

### 11.2 JS: Floating Promise (AST)

```go
// internal/rules/js_floating_promise.go — simplified sketch
func (f *FloatingPromise) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
    // Strategy: find expression_statement nodes whose child is a call_expression
    // returning a Promise (heuristic: async function calls, fetch(), .then() chains
    // not assigned, not awaited, not returned, not void'd).
    //
    // This is inherently heuristic without type information. Focus on high-confidence
    // patterns: fetch(), axios.*(), unassigned async calls.
    ...
}
```

### 11.3 Python: Bare Except (AST)

```go
// internal/rules/py_bare_except.go
type BareExcept struct{}

func (b *BareExcept) Meta() Rule {
    return Rule{
        ID:       "silent-fallback.bare-except",
        Category: "silent-fallback",
        Name:     "Bare except clause",
        Description: "except: without exception type catches SystemExit, KeyboardInterrupt, and all errors",
        Severity: finding.SeverityError,
        FileTypes: []string{".py"},
        ScanType: ScanTypeAST,
    }
}

func (b *BareExcept) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
    var findings []finding.Finding
    // Find except_clause nodes without a type argument
    exceptNodes := scanner.FindNodesByType(tree.RootNode(), "except_clause")
    for _, node := range exceptNodes {
        // A bare except has no type child — just "except:"
        if !scanner.HasChildOfType(node, "identifier") &&
           !scanner.HasChildOfType(node, "tuple") &&
           !scanner.HasChildOfType(node, "attribute") {
            line := scanner.LineNumber(node)
            findings = append(findings, finding.Finding{
                Rule:       b.Meta().ID,
                Severity:   b.Meta().Severity,
                File:       path,
                Line:       line,
                Code:       scanner.SourceLine(source, line),
                Message:    "Bare except: catches everything including SystemExit and KeyboardInterrupt",
                Suggestion: "Specify the exception type: except ValueError: or except Exception: at minimum",
            })
        }
    }
    return findings
}
```

### 11.4 Python: Mutable Default Argument (AST)

```go
// internal/rules/py_mutable_default.go
func (m *MutableDefaultArg) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
    // Find function_definition nodes, inspect default_parameter children,
    // check if the default value is a list, dict, or set literal.
    // Tree-sitter node types: "list", "dictionary", "set" as default values.
    ...
}
```

## 12. Rule ID Namespacing Strategy

Rule IDs use the format `category.rule-name`. The same category can contain rules for different languages. Language is determined by the rule's `FileTypes` field, not the ID.

This means `silent-fallback.empty-catch` (JS) and `silent-fallback.empty-error-check` (Go) coexist naturally. When rule names could collide across languages (rare — most anti-patterns have language-specific names), suffix with the language:

- `bad-defaults.no-timeout` → Go (existing)
- `bad-defaults.no-timeout-fetch` → JS
- `bad-defaults.no-timeout-requests` → Python

The `truthsayer rules --lang python` filter uses `FileTypes`, not ID prefix. This keeps IDs semantic rather than encoding language metadata.

## 13. Test Strategy

### 13.1 Test Fixture Structure

Every rule gets:
1. **Positive fixture**: File that MUST trigger the rule (with expected line numbers)
2. **Negative fixture**: File that MUST NOT trigger the rule (proper pattern)
3. **Edge case fixtures**: Borderline cases (e.g., `except Exception as e: logging.error(e); raise` should NOT trigger `except-pass`)

```
testdata/
├── js/
│   ├── empty_catch.js              # try { x() } catch (e) {} ← triggers
│   ├── empty_catch_negative.js     # try { x() } catch (e) { log(e); throw e } ← clean
│   ├── floating_promise.ts         # fetch('/api') on its own line ← triggers
│   ├── floating_promise_negative.ts # await fetch('/api') ← clean
│   ├── any_assertion.ts            # const x = value as any ← triggers
│   ├── any_assertion_negative.ts   # const x = value as SpecificType ← clean
│   └── ...
├── python/
│   ├── bare_except.py              # except: pass ← triggers
│   ├── bare_except_negative.py     # except ValueError: handle() ← clean
│   ├── mutable_default.py          # def f(items=[]): ← triggers
│   ├── mutable_default_negative.py # def f(items=None): ← clean
│   └── ...
```

### 13.2 Unit Test Pattern

Each rule file gets a corresponding `*_test.go`:

```go
func TestEmptyCatchBlock(t *testing.T) {
    checker := &EmptyCatchBlock{}

    t.Run("triggers on empty catch", func(t *testing.T) {
        findings := scanJSFixture(t, checker, "testdata/js/empty_catch.js")
        assertFindingCount(t, findings, 1)
        assertRuleID(t, findings[0], "silent-fallback.empty-catch")
        assertSeverity(t, findings[0], finding.SeverityError)
        assertLine(t, findings[0], 3)  // known line in fixture
    })

    t.Run("clean on proper catch", func(t *testing.T) {
        findings := scanJSFixture(t, checker, "testdata/js/empty_catch_negative.js")
        assertFindingCount(t, findings, 0)
    })
}
```

Test helpers (`scanJSFixture`, `scanPyFixture`) handle tree-sitter parser setup, keeping individual tests clean.

### 13.3 Integration Tests

```go
func TestFullJSScan(t *testing.T) {
    reg := rules.DefaultRegistry()
    engine := engine.New(reg)
    result, err := engine.Scan("testdata/js")
    require.NoError(t, err)
    // Verify known findings from all JS fixtures
    assertContainsRule(t, result.Findings, "silent-fallback.empty-catch")
    assertContainsRule(t, result.Findings, "silent-fallback.floating-promise")
}

func TestFullPythonScan(t *testing.T) {
    reg := rules.DefaultRegistry()
    engine := engine.New(reg)
    result, err := engine.Scan("testdata/python")
    require.NoError(t, err)
    assertContainsRule(t, result.Findings, "silent-fallback.bare-except")
    assertContainsRule(t, result.Findings, "bad-defaults.mutable-default-arg")
}
```

### 13.4 Cross-Language Tests

Verify that scanning a mixed-language project produces correct findings for each language without interference:

```go
func TestMixedLanguageScan(t *testing.T) {
    // testdata/mixed/ contains .go, .js, .ts, .py files
    result, _ := engine.Scan("testdata/mixed")
    goFindings := filterByExt(result.Findings, ".go")
    jsFindings := filterByExt(result.Findings, ".js", ".ts")
    pyFindings := filterByExt(result.Findings, ".py")
    // Each group has expected findings; no cross-contamination
}
```

### 13.5 Language-Disable Tests

```go
func TestDisablePython(t *testing.T) {
    cfg := config.Load(...)
    cfg.Scan.Languages.Python = boolPtr(false)
    // Scan testdata/python/ → 0 findings
}
```

### 13.6 Benchmarks

```go
func BenchmarkJSScan10k(b *testing.B) {
    // Generate or use 10k LOC JS corpus
    for i := 0; i < b.N; i++ {
        engine.Scan("testdata/bench/js-10k")
    }
}

func BenchmarkPythonScan10k(b *testing.B) {
    for i := 0; i < b.N; i++ {
        engine.Scan("testdata/bench/python-10k")
    }
}
```

Performance target: <5 seconds for 10k LOC per language (same as Go target).

### 13.7 Tree-sitter Parser Tests

Dedicated tests verify that tree-sitter correctly parses representative JS/TS/Python files:

```go
func TestTreeSitterJSParse(t *testing.T) {
    // Verify parser doesn't error on modern JS syntax:
    // optional chaining, nullish coalescing, top-level await, decorators
}

func TestTreeSitterPythonParse(t *testing.T) {
    // Verify parser handles:
    // match/case (3.10), walrus operator (3.8), f-strings, type hints
}
```

### 13.8 Race Detection

```bash
go test -race ./internal/scanner/...  # Verify tree-sitter parsers are safe under concurrency
```

Tree-sitter parsers are NOT thread-safe — each goroutine in the worker pool must have its own parser instance. Tests must verify this.

## 14. Performance Considerations

### 14.1 Parser Pooling

Tree-sitter parsers are not thread-safe. The engine must maintain a pool of parsers per language:

```go
type ParserPool struct {
    pool sync.Pool
    lang *sitter.Language
}

func (p *ParserPool) Get() *sitter.Parser {
    if parser, ok := p.pool.Get().(*sitter.Parser); ok {
        return parser
    }
    parser := sitter.NewParser()
    parser.SetLanguage(p.lang)
    return parser
}

func (p *ParserPool) Put(parser *sitter.Parser) {
    p.pool.Put(parser)
}
```

### 14.2 Lazy Initialization

Don't initialize JS/TS or Python parsers unless files of that type are encountered:

```go
func (e *Engine) jsScanner() *JSScanner {
    e.jsOnce.Do(func() {
        e.js = scanner.NewJSScanner(e.reg.JSASTCheckers())
    })
    return e.js
}
```

### 14.3 Source Reuse

Same pattern as existing Go scanner: read source once, pass to both AST and regex scanners.

## 15. Migration & Backward Compatibility

### 15.1 Zero Breaking Changes

- All existing rule IDs are stable
- Existing `.truthsayer.toml` configs work without modification
- CLI commands and flags are unchanged
- JSON report schema is unchanged (new findings just have new rule IDs)
- Exit codes unchanged

### 15.2 Opt-Out

Teams that don't want JS/TS or Python scanning can disable via config:

```toml
[scan.languages]
javascript = false
typescript = false
python = false
```

Or via CLI: `truthsayer scan --lang go .`

### 15.3 Version Bump

This is a minor version bump (v1 → v1.x or v2.0 if cgo is considered breaking). The cgo requirement is the only potentially breaking change for users who build from source.

Recommendation: Release as **v2.0** with clear documentation about the cgo build requirement. Provide pre-built binaries for all major platforms to minimize impact.

## 16. Implementation Sequence

Ordered by dependency and value. Each phase is independently shippable.

### Phase 1: Tree-sitter Infrastructure (1 sprint)
1. Add tree-sitter Go bindings and language grammars to `go.mod`
2. Implement `internal/scanner/treesitter.go` (shared utilities)
3. Implement `JSASTChecker` and `PyASTChecker` interfaces in `rule.go`
4. Extend `Registry` with `RegisterJSAST`, `RegisterPyAST`, and retrieval methods
5. Implement `internal/scanner/js_scanner.go` and `internal/scanner/py_scanner.go`
6. Extend engine routing in `engine.go`
7. Parser pooling and lazy initialization
8. Tests: parser initialization, basic parse, concurrent safety

### Phase 2: JS/TS Core Rules (1 sprint)
9. `silent-fallback.empty-catch` (AST) + fixture + test
10. `silent-fallback.floating-promise` (AST) + fixture + test
11. `silent-fallback.catch-return-null` (AST) + fixture + test
12. `error-context.promise-reject-non-error` (AST) + fixture + test
13. `bad-defaults.no-timeout-fetch` (AST) + fixture + test
14. `bad-defaults.any-type-assertion` (AST) + fixture + test
15. `bad-defaults.eval-usage` (AST) + fixture + test
16. JS/TS regex rules: `jest-mock-in-src`, `ts-ignore`, `eslint-disable-no-reason`
17. Integration test: full JS/TS testdata scan

### Phase 3: Python Core Rules (1 sprint)
18. `silent-fallback.bare-except` (AST) + fixture + test
19. `silent-fallback.except-pass` (AST) + fixture + test
20. `bad-defaults.mutable-default-arg` (AST) + fixture + test
21. `bad-defaults.no-timeout-requests` (AST) + fixture + test
22. `silent-fallback.subprocess-no-check` (AST) + fixture + test
23. `error-context.raise-from-none` (AST) + fixture + test
24. `error-context.generic-exception` (AST) + fixture + test
25. Python regex rules: `type-ignore-bare`, `noqa-bare`, `requirements-unpinned`
26. Integration test: full Python testdata scan

### Phase 4: Extended Rules + Polish (1 sprint)
27. Remaining JS/TS rules (categories 3–6)
28. Remaining Python rules (categories 3–6)
29. `--lang` CLI flag implementation
30. Config `[scan.languages]` support
31. `doctor` command updates
32. Cross-language integration tests
33. Benchmarks (10k LOC per language)
34. README and docs update
35. CI workflow update (cgo build matrix)

## 17. Success Metrics

| Metric | Target |
|---|---|
| JS/TS rule count | ≥15 AST + ≥5 regex rules |
| Python rule count | ≥12 AST + ≥4 regex rules |
| False positive rate | 0% on popular open-source projects (Express, Flask/Django snippets) |
| Parse success rate | 100% on syntactically valid files (ES2024, Python 3.8–3.13) |
| Scan performance | <5s for 10k LOC per language |
| Binary size | <20MB static binary |
| Test coverage | ≥80% for new scanner and rule code |
| Backward compatibility | 100% — no existing behavior changes |

## 18. Open Questions

1. **Tree-sitter vs WASM**: Start with cgo bindings or pure-Go WASM runtime? **Recommendation**: cgo. Performance matters for CI pipelines. Evaluate WASM if distribution complaints arise.

2. **TypeScript type-aware rules**: Some rules (e.g., `floating-promise`) benefit from type information that tree-sitter doesn't provide. Should we optionally invoke `tsc` for type-checked analysis? **Recommendation**: Defer. Heuristic detection catches 80%+ of cases. Type-aware analysis is a v3 feature.

3. **Python version detection**: Should rules adapt based on detected Python version (e.g., `match/case` is 3.10+)? **Recommendation**: No. Tree-sitter parses all syntax regardless. Rules should detect anti-patterns across all modern Python versions.

4. **Monorepo support**: Projects with Go, JS, and Python in the same repo. Any special handling? **Recommendation**: No. The engine already walks all files. Each file gets the appropriate scanner. It just works.

5. **Custom rule DSL**: Should users be able to define custom tree-sitter query rules in config? **Recommendation**: Defer to v3. Powerful but complex — needs query validation, documentation, and security review.

## 19. Non-Goals

- **Not a linter replacement**: Truthsayer does not replace ESLint, pylint, or golangci-lint. It detects failure-hiding patterns that generic linters miss or deprioritize.
- **Not a type checker**: No type inference or flow analysis. Rules are syntactic/structural.
- **Not a security scanner**: While some rules overlap with security (e.g., `eval()`, secrets), Truthsayer's focus is debuggability and failure visibility, not vulnerability detection.
- **No auto-fix**: Consistent with v1 philosophy — suggestions only, never auto-apply.
- **No IDE plugins**: CLI-first. IDE integration is via existing mechanisms (pre-commit hooks, CI feedback).
