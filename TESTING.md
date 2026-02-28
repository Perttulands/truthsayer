# Test Quality Report — truthsayer

Rubric: `/home/polis/tools/TEST_RUBRIC.md`

---

## Rubric Scores

### BEFORE (2026-02-28, pre-audit baseline)

| Dimension | Score | Notes |
|-----------|-------|-------|
| E2E Realism | 4/5 | 25+ real E2E tests: build binary, scan temp dirs, verify specific rules fire. Covers Go/JS/Python/Bash, config disable/severity/exclude, JSON structure, error cases. Missing: watch/hook/judge/debt/CI/warmup/senate/report E2E. |
| Unit Test Behaviour Focus | 4/5 | Rules tests use `runASTCheckerOnSource` — clean behaviour focus. Tests verify "bad code triggers rule X" / "clean code doesn't". Clear intent. |
| Edge Case & Error Path | 3/5 | Good error paths in E2E and precedent validation. But many rule helper functions at 0%: `isNumericZero`, `isOSStderr`, `containsLenCall`, `callName`, `isTimePkgSelector`. CLI `parseCIOptions` at 35.7%, `filterMatchesByConfidence` at 28.6%. |
| Test Isolation | 4/5 | `t.TempDir()` everywhere, no shared state, no timing deps, deterministic. |
| Regression Value | 3/5 | E2E catches rule detection regressions. Config override tests protect config behavior. But 0% coverage on key helpers means silent regressions. |
| **TOTAL** | **18/25** | **Grade B** |

### AFTER (2026-02-28, post-audit)

| Dimension | Score | Notes |
|-----------|-------|-------|
| E2E Realism | 4/5 | Unchanged — E2E tests were already solid. |
| Unit Test Behaviour Focus | 4/5 | Unchanged — existing pattern is clean. |
| Edge Case & Error Path | 4/5 | Improved: `isNumericZero` now tested (18 cases), `isOSStderr`/`hasStderrWriteCall` now tested (3 Fprintf variants), all magic_number exclusion paths tested (time multiplier, len arithmetic, common call args, return exit codes, octal permissions, negative literals, slice indices). CLI: parseCIOptions 100%, filterMatchesByConfidence 100%, parseScanOptions 100%. Precedent: Validate now covers all branches, normalizeMatchOptions boundary conditions tested. |
| Test Isolation | 4/5 | Unchanged — new tests follow same isolation pattern. |
| Regression Value | 4/5 | Improved: rule helper regressions now caught. Confidence filtering, CI option parsing, check error paths all guarded. Precedent store edge cases covered. |
| **TOTAL** | **20/25** | **Grade B** |

---

## Coverage Delta

| Package | Before | After | Delta |
|---------|--------|-------|-------|
| internal/rules | 85.8% | 88.8% | +3.0% |
| internal/cli | 70.6% | 73.6% | +3.0% |
| internal/precedent | 78.5% | 91.3% | +12.8% |
| Overall (all packages) | ~83.3% | ~85%+ | +~2% |

### Function-level improvements

**Rules (3 target files):**
- `bare_return_on_error.go`: `isNumericZero` 0% -> tested (18 table-driven cases), `isZeroValueExpr` 55.6% -> improved (hex, octal, binary, float zero values), `isAllZeroValueReturn` covered
- `error_path_no_log.go`: `isOSStderr` 0% -> tested (Fprint, Fprintf, Fprintln to os.Stderr), `hasStderrWriteCall` 31.8% -> improved, test file skip tested, multi-error-path tested
- `magic_number.go`: `containsLenCall` 0% -> tested, `callName` 0% -> tested (SplitN, ParseInt), `isTimePkgSelector` 0% -> tested, `isReturnExitCode` 44.4% -> improved, `isLenArithmetic` 28.6% -> tested, `isIndexLiteral` 45.5% -> improved (slice), `isCommonCallArg` 55.6% -> improved, `isOctalLiteral` 66.7% -> improved

**CLI:**
- `parseCIOptions`: 35.7% -> 100%
- `filterMatchesByConfidence`: 28.6% -> 100%
- `strongestMatch`: 100% (maintained)
- `effectiveConfidence`: 66.7% -> 100%
- `parseScanOptions`: 85.7% -> 100%
- `parseJudgeOptions`: 71% -> 78.3%
- `runCheck`: 70.4% -> 92.6%
- `runScan`: 71.4% -> 78.6%

**Precedent:**
- `Validate`: 64.7% -> ~100% (all 8 validation branches tested)
- `Path()`: 0% -> 100%
- `NewStore`: 66.7% -> 100% (empty, whitespace, custom paths)
- `normalizeMatchOptions`: 60% -> 100% (both bounds tested)
- `precedenceConfidence`: 66.7% -> 100% (zero, negative, positive)
- `Load`: 71.4% -> improved (empty file, invalid JSON, invalid record)
- `AddOrUpdateJudgment`: 79.3% -> improved (load error, initial confidence, LastSeen)

---

## Honest Assessment

### What's genuinely good

1. **E2E tests are REAL.** They build the actual binary, create known-bad code in temp dirs, and assert specific rules fire. This is not smoke-test theater.
2. **Rule tests test behaviour, not internals.** The `runASTCheckerOnSource` pattern is clean — write bad code, check if rule fires.
3. **Test isolation is excellent.** `t.TempDir()` everywhere, no shared mutable state, no flaky timing.
4. **Precedent store tests are thorough.** Round-trip persistence, validation, matching with confidence/limit, AddOrUpdateJudgment confidence evolution.

### Remaining gaps

1. **`root.go:Run` at 11.5% and `printUsage` at 0%.** These read `os.Args` directly — hard to unit test. The E2E tests cover them via the binary, which is the right approach.
2. **`runWatch` at 32.7%.** Filesystem watcher tests are inherently hard. The existing watcher package tests (83.3%) cover the core logic.
3. **`ciDiffBase` at 40%.** Requires specific git branch states — tested indirectly via CI integration tests.
4. **`cmd/truthsayer` and `main.go` at 0%.** Thin entry points, covered by E2E.
5. **`internal/cost` at 72.2%.** Cost tracker has some untested paths.
6. **`internal/llm` at 82.9%.** Real LLM calls can't be unit-tested without mocking.

### E2E Audit

The E2E tests (tests/e2e/e2e_test.go) are **meaningful**. They:
- Test 4 languages (Go, JS, Python, Bash) with real anti-pattern code
- Verify specific rule IDs fire (not just "something found")
- Test config: disable rules, severity override, exclude patterns/dirs, explicit config path
- Test error cases: nonexistent path, file-instead-of-dir, bad config, bad severity, unknown command
- Verify JSON structure fields (version, scan_time, severity values, finding count consistency)
- Test clean code exit 0 and empty directory
- Test multi-language concurrent scan
- Test severity sorting, finding field population
- Test REASON comment suppression

Missing from E2E: watch, hook, judge, warmup, debt, senate, report, CI commands. These are tested at the unit/integration level within internal/cli.

---

## Changelog

### 2026-02-28 — Agent: Apollo (Claude Opus)

- Added: 10 tests for `bare_return_on_error.go` covering numeric zeros (hex, octal, binary, float), false zero value, REASON comment suppression, non-zero values, interface declarations, plus `isNumericZero` table-driven test (18 cases)
- Added: 8 tests for `error_path_no_log.go` covering test file skip, log call exemption, single-line return exemption, Fprint/Fprintln/Fprintf to stderr, non-error check ignored, multiple error paths
- Added: 13 tests for `magic_number.go` covering negative literals, slice indices, return exit codes, octal permissions (both styles), small/large comparisons, len arithmetic, common call args (SplitN, ParseInt), time multipliers, make calls, float literals, per-line dedup, non-func declarations
- Added: 12 tests for `parseScanOptions` (format, lang, parallel, bead-threshold, create-beads, use-precedents error/valid paths), plus scan error paths (file-not-dir, invalid lang, JSON format output)
- Added: 7 tests for `parseCIOptions` (default path, custom path, bead-threshold valid/missing/invalid/negative, create-beads no-op)
- Added: 11 tests for judge helpers (`filterMatchesByConfidence`, `strongestMatch`, `effectiveConfidence`, `parseJudgeOptions` error paths)
- Added: 4 tests for `runCheck` error paths (no args, directory arg, nonexistent file, bad config)
- Added: 22 tests for precedent store gaps (NewStore empty/whitespace/custom, Path, all 8 Validate branches, Load empty/invalid/bad-record, Save creates dirs, Add sets created_at, Query no-match, normalizeMatchOptions bounds, precedenceConfidence zero/negative/positive, Match empty/empty-pattern, Store.Match load error, AddOrUpdateJudgment load error/LastSeen/initial confidence)
- Coverage delta: rules 85.8% -> 88.8%, cli 70.6% -> 73.6%, precedent 78.5% -> 91.3%
