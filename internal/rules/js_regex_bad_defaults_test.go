package rules

import "testing"

// --- JSTsIgnore ---

func TestJSTsIgnore_Positive(t *testing.T) {
	checker := &JSTsIgnore{}
	lines := []string{
		`// @ts-ignore`,
		`const x: any = getValue();`,
	}
	findings := checker.CheckLines("src/utils.ts", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 1 {
		t.Errorf("expected line 1, got %d", findings[0].Line)
	}
}

func TestJSTsIgnore_WithSpaces(t *testing.T) {
	checker := &JSTsIgnore{}
	lines := []string{`  //  @ts-ignore  `}
	findings := checker.CheckLines("src/utils.ts", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSTsIgnore_WithExplanation(t *testing.T) {
	checker := &JSTsIgnore{}
	lines := []string{`// @ts-ignore legacy API returns untyped`}
	findings := checker.CheckLines("src/utils.ts", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when explanation present, got %d", len(findings))
	}
}

func TestJSTsIgnore_TsExpectError(t *testing.T) {
	checker := &JSTsIgnore{}
	lines := []string{`// @ts-expect-error testing bad input`}
	findings := checker.CheckLines("src/utils.ts", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for @ts-expect-error, got %d", len(findings))
	}
}

// --- JSEslintDisableNoReason ---

func TestJSEslintDisableNoReason_Positive(t *testing.T) {
	checker := &JSEslintDisableNoReason{}
	lines := []string{
		`// eslint-disable-next-line no-unused-vars`,
		`const x = 42;`,
	}
	findings := checker.CheckLines("src/utils.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSEslintDisableNoReason_BareDisable(t *testing.T) {
	checker := &JSEslintDisableNoReason{}
	lines := []string{`// eslint-disable-next-line`}
	findings := checker.CheckLines("src/utils.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSEslintDisableNoReason_WithReason(t *testing.T) {
	checker := &JSEslintDisableNoReason{}
	lines := []string{`// eslint-disable-next-line no-unused-vars -- needed for API compatibility`}
	findings := checker.CheckLines("src/utils.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when reason present, got %d", len(findings))
	}
}

func TestJSEslintDisableNoReason_BlockDisable(t *testing.T) {
	checker := &JSEslintDisableNoReason{}
	lines := []string{`// eslint-disable no-console`}
	findings := checker.CheckLines("src/utils.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for block disable, got %d", len(findings))
	}
}

// --- JSNoStrictMode ---

func TestJSNoStrictMode_Positive(t *testing.T) {
	checker := &JSNoStrictMode{}
	lines := []string{
		`const http = require('http');`,
		`module.exports = {};`,
	}
	findings := checker.CheckLines("server.cjs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSNoStrictMode_WithStrict(t *testing.T) {
	checker := &JSNoStrictMode{}
	lines := []string{
		`'use strict';`,
		`const http = require('http');`,
	}
	findings := checker.CheckLines("server.cjs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings with 'use strict', got %d", len(findings))
	}
}

func TestJSNoStrictMode_DoubleQuotes(t *testing.T) {
	checker := &JSNoStrictMode{}
	lines := []string{
		`"use strict";`,
		`const http = require('http');`,
	}
	findings := checker.CheckLines("server.cjs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings with double-quote strict, got %d", len(findings))
	}
}

func TestJSNoStrictMode_CommentBeforeStrict(t *testing.T) {
	checker := &JSNoStrictMode{}
	lines := []string{
		`// Server entry point`,
		`'use strict';`,
		`const http = require('http');`,
	}
	findings := checker.CheckLines("server.cjs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings with comment before strict, got %d", len(findings))
	}
}

func TestJSNoStrictMode_EmptyFile(t *testing.T) {
	checker := &JSNoStrictMode{}
	findings := checker.CheckLines("empty.cjs", []string{})
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty file, got %d", len(findings))
	}
}
