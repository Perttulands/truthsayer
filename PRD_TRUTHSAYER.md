# Truthsayer PRD — Ralph Execution Format

**Tech Stack**: Go, AST scanning, regex patterns, TOML config, JSON reports

**Reference**: docs/PRD.md (full PRD), docs/SPEC.md (spec)


## Sprint 1: Core Scanner Infrastructure
**Status:** IN PROGRESS

- [x] **US-001**: As a developer, I want to scan a directory for anti-patterns so I can find hidden problems before they reach production.
- [x] **US-002**: As a developer, I want to scan a single file so I can check my work before committing.
- [x] **US-003**: As a developer, I want findings to include file path, line number, code snippet, and a fix suggestion so I can act on them immediately.
- [x] **US-004**: As a developer, I want scan results sorted by severity (error > warning > info) so I can triage effectively.
- [x] **US-005**: As a developer, I want the scanner to skip vendor/, node_modules/, and .git/ directories by default so scans are fast and relevant.
- [ ] **US-REVIEW-S1**: Review Sprint 1 — run tests, verify integration, fix issues.

## Sprint 2: Detection Rules Engine
**Status:** NOT STARTED

- [ ] **US-006**: As a developer, I want to watch a directory for file changes and get instant feedback on new anti-patterns so I catch problems as I code.
- [ ] **US-007**: As a developer, I want watch mode to only report findings in the changed lines (not the entire file) so I'm not overwhelmed by pre-existing issues.
- [ ] **US-008**: As a developer, I want to generate a JSON report of all findings so I can feed it into other tools or dashboards.
- [ ] **US-009**: As a CI engineer, I want the scanner to exit with code 1 when errors are found so I can use it as a pipeline quality gate.
- [ ] **US-010**: As a developer, I want a human-readable terminal summary after each scan showing counts by severity and category.
- [ ] **US-REVIEW-S2**: Review Sprint 2 — run tests, verify integration, fix issues.

## Sprint 3: CLI & Reporting
**Status:** NOT STARTED

- [ ] **US-011**: As a team lead, I want to configure which rules are enabled/disabled via a TOML config file so I can tailor the scanner to our codebase.
- [ ] **US-012**: As a developer, I want to exclude specific files or directories from scanning via config so I can skip generated code or legacy modules.
- [ ] **US-013**: As a developer, I want to override severity levels per rule so I can promote warnings to errors for rules my team cares about.
- [ ] **US-014**: As a developer, I want to list all available detection rules with their IDs, descriptions, and severity so I understand what the scanner checks.
- [ ] **US-015**: As a developer, I want to list only the currently enabled rules so I know what's active in my project.
- [ ] **US-REVIEW-S3**: Review Sprint 3 — run tests, verify integration, fix issues.

## Sprint 4: Integration & Polish
**Status:** NOT STARTED

- [ ] **US-016**: As a developer, I want to run Truthsayer as a git pre-commit hook on staged files so anti-patterns are caught before they're committed.
- [ ] **US-017**: As a CI engineer, I want to run Truthsayer in a GitHub Actions workflow and have it fail the build on errors.
- [ ] **US-018**: As a developer, I want a `doctor` command that checks my installation, config validity, and reports the active rule count so I can verify everything works.
- [ ] **US-019**: As a developer, I want `--version` to print the version string so I can verify which build I'm running.
- [ ] **US-REVIEW-S4**: Review Sprint 4 — run tests, verify integration, fix issues.
