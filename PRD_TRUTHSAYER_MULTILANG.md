# PRD: Truthsayer Multi-Language Extension

**Tech Stack**: Go, Tree-sitter, TOML config, JSON reports

**Source**: PRD-multilang.md — Extending Truthsayer's anti-pattern detection from Go/bash to JavaScript/TypeScript and Python via tree-sitter AST parsing.

**Test Command**: `go test ./... -count=1`

**Repo**: /home/perttu/truthsayer

---

## Sprint 1: Tree-sitter Infrastructure
**Status:** COMPLETE

- [x] **US-100** Add tree-sitter dependencies and verify build
  - Add `github.com/smacker/go-tree-sitter` and language grammars (javascript, typescript, tsx, python) to `go.mod`
  - Create a minimal smoke test that initializes each parser and parses a trivial snippet
  - Verify cgo build works: `go build ./...`
  - **Files**: `go.mod`, `go.sum`, `internal/scanner/treesitter_test.go`
  - **Acceptance**: `go test ./internal/scanner/ -run TestTreeSitter -count=1` passes; parsers initialize for JS, TS, TSX, Python

- [x] **US-101** Implement shared tree-sitter utilities
  - Create `internal/scanner/treesitter.go` with helper functions:
    - `FindNodesByType(root, types...) []*Node`
    - `NodeText(node, source) string`
    - `LineNumber(node) int`
    - `SourceLine(source, line) string`
    - `HasChildOfType(node, childType) bool`
    - `IsInsideFunction(node) bool`
    - `IsTestFile(path) bool` (handles `*_test.py`, `*.spec.ts`, `*.test.js`, etc.)
  - Unit tests for each utility function
  - **Files**: `internal/scanner/treesitter.go`, `internal/scanner/treesitter_test.go`
  - **Acceptance**: All utility functions tested with representative JS/TS/Python AST nodes

- [x] **US-102** Define JSASTChecker and PyASTChecker interfaces + extend Registry
  - Add `JSASTChecker` interface to `internal/rules/rule.go`: `Meta() Rule` + `CheckJSAST(tree, source, path) []Finding`
  - Add `PyASTChecker` interface to `internal/rules/rule.go`: `Meta() Rule` + `CheckPyAST(tree, source, path) []Finding`
  - Extend `Rule` struct if needed (ensure `FileTypes` and `ScanType` fields exist)
  - Add to `internal/rules/registry.go`: `RegisterJSAST()`, `RegisterPyAST()`, `JSASTCheckers()`, `PyASTCheckers()`
  - Tests: register mock checkers, verify retrieval
  - **Files**: `internal/rules/rule.go`, `internal/rules/registry.go`, `internal/rules/registry_test.go`
  - **Acceptance**: Can register and retrieve JS/Python AST checkers from registry

- [x] **US-103** Implement JSScanner with parser pooling
  - Create `internal/scanner/js_scanner.go`:
    - `ParserPool` using `sync.Pool` for thread-safe parser reuse
    - `JSScanner` struct with parser pool and checker list
    - `Scan(path) ([]Finding, []string, error)` — read file, parse with tree-sitter, dispatch to all JS AST checkers
    - Language detection: `.js/.mjs/.cjs` → JavaScript grammar, `.ts` → TypeScript grammar, `.tsx` → TSX grammar, `.jsx` → JavaScript grammar
  - Tests: scan a simple JS file with a no-op checker, verify parser pooling under concurrency
  - **Files**: `internal/scanner/js_scanner.go`, `internal/scanner/js_scanner_test.go`
  - **Acceptance**: JSScanner parses JS/TS files, dispatches to checkers, concurrent test with `-race` passes

- [x] **US-104** Implement PyScanner with parser pooling
  - Create `internal/scanner/py_scanner.go`:
    - `PyScanner` struct with parser pool and checker list
    - `Scan(path) ([]Finding, []string, error)` — same pattern as JSScanner
  - Tests: scan a simple Python file with a no-op checker, verify parser pooling under concurrency
  - **Files**: `internal/scanner/py_scanner.go`, `internal/scanner/py_scanner_test.go`
  - **Acceptance**: PyScanner parses `.py` files, dispatches to checkers, concurrent test with `-race` passes

- [x] **US-105** Extend engine routing and walker for JS/TS/Python
  - Modify `internal/engine/engine.go`:
    - Add `jsScanner` and `pyScanner` fields with lazy initialization (`sync.Once`)
    - Route `.js/.jsx/.ts/.tsx/.mjs/.cjs` → JSScanner + RegexScanner
    - Route `.py/.pyi` → PyScanner + RegexScanner
    - Helper: `isJSExt(ext) bool`
  - Update walker's `supportedExts` map to include `.py`, `.pyi`
  - Update default `exclude_dirs` to include `node_modules`, `__pycache__`, `.venv`, `dist`, `build`
  - Tests: engine routes files to correct scanners based on extension
  - **Files**: `internal/engine/engine.go`, `internal/walker/walker.go` (or equivalent), `internal/engine/engine_test.go`
  - **Acceptance**: Engine correctly routes JS/TS files to JSScanner, Python files to PyScanner, Go files unchanged

- [x] **US-106** Tree-sitter parser correctness and edge case tests
  - Test that tree-sitter correctly parses modern syntax:
    - JS: optional chaining, nullish coalescing, top-level await, class fields, decorators
    - TS: generics, type assertions, enums, decorators, satisfies operator
    - Python: match/case (3.10), walrus operator (3.8), f-strings, type hints, async/await
  - Test error-tolerant parsing: syntactically broken files parse without crashing
  - Test race safety: `go test -race` with concurrent parser access
  - **Files**: `internal/scanner/treesitter_parse_test.go`, `testdata/js/syntax_modern.js`, `testdata/js/syntax_modern.ts`, `testdata/python/syntax_modern.py`
  - **Acceptance**: All parse tests pass including broken syntax; `-race` clean

- [x] **US-REVIEW-S1** Review Sprint 1
  - All tree-sitter infrastructure compiles and tests pass
  - `go test ./... -count=1` green
  - `go test -race ./internal/scanner/...` clean
  - No regressions in existing Go/bash rules

---

## Sprint 2: JS/TS Rules
**Status:** IN PROGRESS

- [x] **US-200** JS silent-fallback rules: empty-catch, catch-return-null, floating-promise
  - Implement `internal/rules/js_empty_catch.go` — detect `catch (e) {}` with empty body
  - Implement `internal/rules/js_catch_return_null.go` — detect `.catch(() => null)` and `.catch(() => undefined)`
  - Implement `internal/rules/js_floating_promise.go` — detect unhandled promise-returning calls (fetch, async calls) not awaited/assigned/returned
  - Register all three in `DefaultRegistry()`
  - Test fixtures + unit tests for each (positive + negative cases)
  - **Files**: `internal/rules/js_empty_catch.go`, `internal/rules/js_catch_return_null.go`, `internal/rules/js_floating_promise.go`, `internal/rules/js_empty_catch_test.go`, `internal/rules/js_catch_return_null_test.go`, `internal/rules/js_floating_promise_test.go`, `testdata/js/empty_catch.js`, `testdata/js/empty_catch_negative.js`, `testdata/js/catch_return_null.js`, `testdata/js/catch_return_null_negative.js`, `testdata/js/floating_promise.ts`, `testdata/js/floating_promise_negative.ts`
  - **Acceptance**: Each rule triggers on positive fixtures, silent on negative fixtures

- [x] **US-201** JS silent-fallback rules: callback-err-ignored, optional-chain-silence
  - Implement `internal/rules/js_callback_err_ignored.go` — detect `(err, data) => { ... }` callbacks where `err` is never referenced
  - Implement `internal/rules/js_optional_chain_silence.go` — detect deep optional chaining `a?.b?.c?.d?.e` (>3 levels)
  - Register in `DefaultRegistry()`
  - Test fixtures + unit tests
  - **Files**: `internal/rules/js_callback_err_ignored.go`, `internal/rules/js_optional_chain_silence.go`, corresponding `_test.go` and `testdata/js/` fixtures
  - **Acceptance**: Rules trigger correctly, no false positives on normal callback patterns and reasonable optional chaining

- [x] **US-202** JS error-context rules: rethrow-no-wrap, generic-error-message, promise-reject-non-error, console-error-no-throw
  - Implement `internal/rules/js_rethrow_no_wrap.go` — `catch (e) { throw e }` without wrapping
  - Implement `internal/rules/js_generic_error_message.go` — `throw new Error("failed")` with no interpolation
  - Implement `internal/rules/js_promise_reject.go` — `Promise.reject("string")` or `reject(42)` with non-Error values
  - Implement `internal/rules/js_console_error_no_throw.go` — `console.error(...)` in catch without re-throw
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: 4 rule files + 4 test files + 8 fixture files in `testdata/js/`
  - **Acceptance**: All four rules trigger on anti-patterns, silent on proper error handling

- [x] **US-203** JS error-context rule: http-200-on-error
  - Implement `internal/rules/js_http_200_on_error.go` — detect `res.status(200)` or `res.json(...)` after error detection in Express/Koa/Fastify handlers
  - Heuristic: look for `res.status(200)` inside catch blocks, or `res.json()` following error variable checks
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: `internal/rules/js_http_200_on_error.go`, `internal/rules/js_http_200_on_error_test.go`, `testdata/js/http_200_on_error.js`, `testdata/js/http_200_on_error_negative.js`
  - **Acceptance**: Detects status 200 after error in Express-style handlers, clean on proper error responses

- [x] **US-204** JS trace-gaps + mock-leakage AST rules
  - Implement `internal/rules/js_no_error_handler_express.go` — Express app without error middleware `(err, req, res, next)`
  - Implement `internal/rules/js_test_import_in_src.go` — import from `@testing-library/*`, `vitest`, `jest` in non-test files
  - Implement `internal/rules/js_env_test_check.go` — `process.env.NODE_ENV === 'test'` in production code
  - Implement `internal/rules/js_missing_correlation_id.go` — HTTP middleware chain without correlation ID propagation
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: 4 rule files + 4 test files + 8 fixture files
  - **Acceptance**: All rules detect their anti-patterns accurately

- [x] **US-205** JS bad-defaults AST rules: no-timeout-fetch, any-type-assertion, non-null-assertion, eval-usage
  - Implement `internal/rules/js_no_timeout_fetch.go` — `fetch()` without `AbortController`/`signal`
  - Implement `internal/rules/js_any_assertion.go` — `as any` type assertion (TS files only)
  - Implement `internal/rules/js_non_null_assertion.go` — `variable!` non-null assertion (TS files only)
  - Implement `internal/rules/js_eval_usage.go` — `eval()` or `new Function()` in production code
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: 4 rule files + 4 test files + 8 fixture files (TS fixtures for type-specific rules)
  - **Acceptance**: Rules trigger on anti-patterns; TS-specific rules only fire on `.ts/.tsx` files

- [x] **US-206** JS test-isolation AST rules
  - Implement `internal/rules/js_no_afterall_cleanup.go` — `beforeAll`/`beforeEach` without matching `afterAll`/`afterEach` cleanup
  - Implement `internal/rules/js_test_only_import.go` — imports exclusively used in test setup living in `src/`
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: 2 rule files + 2 test files + 4 fixture files
  - **Acceptance**: Rules detect missing cleanup and misplaced test imports

- [x] **US-207** JS/TS regex rules (all categories)
  - Implement regex rules (can be individual rule files or grouped by category):
    - `mock-leakage.jest-mock-in-src` — `jest.mock(`, `jest.fn(`, `jest.spyOn(` in non-test files
    - `mock-leakage.storybook-in-src` — `.stories.` import in non-story files
    - `bad-defaults.ts-ignore` — `@ts-ignore` without explanation
    - `bad-defaults.eslint-disable-no-reason` — `eslint-disable` without comment
    - `bad-defaults.no-strict-mode` — CommonJS `.cjs` without `'use strict'`
    - `trace-gaps.no-unhandled-rejection` — no `process.on('unhandledRejection')` in entry points
    - `trace-gaps.console-log-in-production` — `console.log` in non-test source
    - `config-smells.hardcoded-api-url` — hardcoded `http://localhost`, `https://api.` URLs in source
    - `config-smells.dotenv-no-example` — `.env` without `.env.example`
  - Register all in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: Rule files in `internal/rules/`, test files, fixtures in `testdata/js/`
  - **Acceptance**: All 9 regex rules trigger correctly on positive fixtures, silent on negative

- [x] **US-208** JS/TS integration test suite
  - Create full integration test: scan `testdata/js/` directory with full engine
  - Verify all expected findings from all JS/TS rules appear
  - Verify no false positives on negative fixtures
  - Test that JS/TS scan doesn't interfere with existing Go scan
  - **Files**: `internal/engine/js_integration_test.go`, verify all `testdata/js/` fixtures are complete
  - **Acceptance**: Integration test passes; all JS/TS rule IDs present in scan results for positive fixtures

- [ ] **US-REVIEW-S2** Review Sprint 2
  - All 21+ JS/TS rules implemented and tested
  - `go test ./... -count=1` green
  - No regressions in Go/bash rules
  - All test fixtures have positive + negative cases

---

## Sprint 3: Python Rules
**Status:** NOT STARTED

- [ ] **US-300** Python silent-fallback rules: bare-except, except-pass, except-broad
  - Implement `internal/rules/py_bare_except.go` — `except:` without exception type
  - Implement `internal/rules/py_except_pass.go` — `except SomeError: pass`
  - Implement `internal/rules/py_except_broad.go` — `except Exception:` (overly broad)
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: 3 rule files + 3 test files + 6 fixture files in `testdata/python/`
  - **Acceptance**: Each rule triggers on its anti-pattern; `except ValueError:` with handling is clean

- [ ] **US-301** Python silent-fallback rules: subprocess-no-check, getattr-silent-default, dict-get-none
  - Implement `internal/rules/py_subprocess_no_check.go` — `subprocess.run()`/`subprocess.call()` without `check=True`
  - Implement `internal/rules/py_getattr_silent_default.go` — `getattr(obj, 'attr', None)` in bug-indicating contexts
  - Implement `internal/rules/py_dict_get_none.go` — `dict.get(key)` without explicit default where None causes downstream failure
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: 3 rule files + 3 test files + 6 fixture files
  - **Acceptance**: Rules trigger correctly; `subprocess.run(..., check=True)` is clean

- [ ] **US-302** Python error-context rules: raise-from-none, bare-raise-different, generic-exception, string-exception, log-and-raise
  - Implement `internal/rules/py_raise_from_none.go` — `raise X from None`
  - Implement `internal/rules/py_bare_raise_different.go` — `except ErrorA: raise ErrorB()` without `from`
  - Implement `internal/rules/py_generic_exception.go` — `raise Exception("failed")`
  - Implement `internal/rules/py_string_exception.go` — `raise "error"` (Python 2 idiom)
  - Implement `internal/rules/py_log_and_raise.go` — `logging.error(e); raise` duplicate reporting
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: 5 rule files + 5 test files + 10 fixture files
  - **Acceptance**: All five rules trigger on anti-patterns, silent on proper exception handling

- [ ] **US-303** Python bad-defaults AST rules: mutable-default-arg, no-timeout-requests, star-import, global-state, no-encoding-open
  - Implement `internal/rules/py_mutable_default.go` — `def f(items=[])` mutable default
  - Implement `internal/rules/py_no_timeout_requests.go` — `requests.get/post()` without `timeout=`
  - Implement `internal/rules/py_star_import.go` — `from module import *`
  - Implement `internal/rules/py_global_state.go` — module-level mutable globals (`ITEMS = []`, `CACHE = {}`)
  - Implement `internal/rules/py_no_encoding_open.go` — `open(file)` without `encoding=`
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: 5 rule files + 5 test files + 10 fixture files
  - **Acceptance**: All rules trigger correctly; `def f(items=None)` and `requests.get(url, timeout=30)` are clean

- [ ] **US-304** Python mock-leakage + trace-gaps AST rules
  - Implement `internal/rules/py_unittest_import.go` — `from unittest.mock import ...` in non-test files
  - Implement `internal/rules/py_debug_flag.go` — `if __debug__:` or `if DEBUG:` in production code with side effects
  - Implement `internal/rules/py_silent_request.go` — `requests.get/post()` without `raise_for_status()` or status check
  - Register in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: 3 rule files + 3 test files + 6 fixture files
  - **Acceptance**: Rules detect anti-patterns; test files with mock imports are ignored

- [ ] **US-305** Python regex rules (all categories)
  - Implement regex rules:
    - `trace-gaps.print-debug` — `print()` in non-test/non-script files
    - `trace-gaps.no-logging-config` — no `logging.basicConfig()` or `logging.getLogger()` in entry points
    - `mock-leakage.pytest-fixture-in-src` — `@pytest.fixture` in non-test files
    - `bad-defaults.type-ignore-bare` — `# type: ignore` without specific error code
    - `bad-defaults.noqa-bare` — `# noqa` without specific code
    - `config-smells.hardcoded-credentials-py` — `password = "..."`, `api_key = "..."` literals
    - `config-smells.requirements-unpinned` — `requirements.txt` without `==` version pins
  - Register all in `DefaultRegistry()`, test fixtures + unit tests
  - **Files**: Rule files in `internal/rules/`, test files, fixtures in `testdata/python/`
  - **Acceptance**: All 7 regex rules trigger correctly

- [ ] **US-306** Python integration test suite
  - Create full integration test: scan `testdata/python/` directory with full engine
  - Verify all expected findings from all Python rules appear
  - Verify no false positives on negative fixtures
  - Test that Python scan doesn't interfere with existing Go/JS scans
  - **Files**: `internal/engine/py_integration_test.go`, verify all `testdata/python/` fixtures complete
  - **Acceptance**: Integration test passes; all Python rule IDs present in results for positive fixtures

- [ ] **US-REVIEW-S3** Review Sprint 3
  - All 15+ Python rules implemented and tested
  - `go test ./... -count=1` green
  - No regressions in Go/bash/JS rules

---

## Sprint 4: Config, CLI, Integration & Polish
**Status:** NOT STARTED

- [ ] **US-400** Config extensions: per-language enable/disable
  - Add `LanguageConfig` struct with `*bool` fields for Go, JavaScript, TypeScript, Python, Bash
  - Parse `[scan.languages]` section from `.truthsayer.toml`
  - Wire language config into engine: skip scanner initialization for disabled languages
  - Update default `exclude_dirs` to include `node_modules`, `__pycache__`, `.venv`, `dist`, `build`
  - Update `max_file_size` default consideration and `exclude_patterns` for `.min.js`, `.bundle.js`, `.pyc`
  - Tests: disable Python → scan produces no Python findings; disable JS → no JS findings
  - **Files**: `internal/config/config.go`, `internal/engine/engine.go`, `internal/config/config_test.go`
  - **Acceptance**: Language enable/disable works via TOML config; defaults are all-enabled

- [ ] **US-401** CLI --lang flag for scan and rules commands
  - Add `--lang` flag to `scan` command: comma-separated language filter (`go,python`, `js,ts`, etc.)
  - Add `--lang` flag to `rules` command: list only rules matching specified languages
  - Implement language aliases: `js`/`javascript`, `ts`/`typescript`, `py`/`python`, `sh`/`bash`/`shell`
  - Map aliases to file extensions, filter rules by `FileTypes` field
  - Tests: `--lang python` shows only Python rules; `--lang go,js` scans only Go and JS files
  - **Files**: `cmd/scan.go` (or equivalent CLI file), `cmd/rules.go`, CLI test files
  - **Acceptance**: `--lang` flag works for both scan and rules commands with all aliases

- [ ] **US-402** Doctor command extension for multi-language
  - Extend `truthsayer doctor` to show:
    - Rule count per language: "47 rules active (21 Go, 16 JS/TS, 10 Python)"
    - Parser status: "JS/TS AST parser: available (tree-sitter)"
    - Parser status: "Python AST parser: available (tree-sitter)"
  - Verify each parser can initialize successfully
  - Tests: doctor output includes language-specific checks
  - **Files**: `cmd/doctor.go` (or equivalent), doctor test file
  - **Acceptance**: `truthsayer doctor` shows per-language rule counts and parser availability

- [ ] **US-403** Cross-language integration tests
  - Create `testdata/mixed/` directory with `.go`, `.js`, `.ts`, `.py` files containing known anti-patterns
  - Integration test: scan mixed directory, verify findings per language with no cross-contamination
  - Test language-disable: disable Python via config, rescan, verify no Python findings
  - Test `--lang` flag: scan with `--lang go` produces only Go findings
  - **Files**: `internal/engine/integration_test.go`, `testdata/mixed/` fixtures
  - **Acceptance**: Mixed-language scan produces correct findings for each language; disable/filter works

- [ ] **US-404** Performance benchmarks
  - Create benchmark corpus: `testdata/bench/js-10k/` and `testdata/bench/python-10k/` (generated or curated, ~10k LOC each)
  - Implement benchmarks: `BenchmarkJSScan10k`, `BenchmarkPythonScan10k`
  - Run and verify performance target: <5 seconds for 10k LOC per language
  - Run `go test -race ./internal/scanner/...` to confirm race-free under load
  - **Files**: `internal/engine/benchmark_test.go`, `testdata/bench/` corpora (can be generated by test setup)
  - **Acceptance**: Benchmarks pass; <5s per 10k LOC; race-free

- [ ] **US-405** Documentation and CI updates
  - Update README.md: document multi-language support, new rule categories, cgo build requirement
  - Document new config options (`[scan.languages]`, new exclude patterns)
  - Document `--lang` CLI flag and language aliases
  - Update CI workflow: add `build-essential` / C compiler to build matrix
  - Verify cross-platform build (Linux amd64 at minimum)
  - **Files**: `README.md`, CI config files (`.github/workflows/` or equivalent)
  - **Acceptance**: README documents all new features; CI builds and tests pass with cgo

- [ ] **US-REVIEW-S4** Review Sprint 4
  - All config/CLI changes working
  - Full test suite green: `go test ./... -count=1`
  - `go test -race ./...` clean
  - Benchmarks within targets
  - Documentation complete
  - No regressions across all languages
