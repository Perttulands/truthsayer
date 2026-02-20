package rules

import "testing"

func TestHiddenFailureBash_OrTrue(t *testing.T) {
	checker := &HiddenFailureBash{}
	lines := []string{
		"#!/bin/bash",
		"set -euo pipefail",
		"cmd || true",
	}
	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 3 {
		t.Errorf("expected line 3, got %d", findings[0].Line)
	}
}

func TestHiddenFailureBash_DevNull(t *testing.T) {
	checker := &HiddenFailureBash{}
	lines := []string{
		"#!/bin/bash",
		"cmd 2>/dev/null",
	}
	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestHiddenFailureBash_Comment(t *testing.T) {
	checker := &HiddenFailureBash{}
	lines := []string{
		"#!/bin/bash",
		"# cmd || true is bad",
	}
	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for comment, got %d", len(findings))
	}
}

func TestHiddenFailureBash_Clean(t *testing.T) {
	checker := &HiddenFailureBash{}
	lines := []string{
		"#!/bin/bash",
		"cmd || { echo 'failed' >&2; exit 1; }",
	}
	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestHiddenFailureBash_TrapLineOrTrueExempt(t *testing.T) {
	checker := &HiddenFailureBash{}
	lines := []string{
		"#!/bin/bash",
		"trap 'rm -f \"$tmp\" || true' EXIT",
	}

	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for trap context, got %d", len(findings))
	}
}

func TestHiddenFailureBash_TrapInvokedCleanupExempt(t *testing.T) {
	checker := &HiddenFailureBash{}
	lines := []string{
		"#!/bin/bash",
		"cleanup() {",
		"  rm -f \"$tmp\" || true",
		"}",
		"trap 'cleanup' EXIT",
	}

	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for trap-invoked cleanup, got %d", len(findings))
	}
}

func TestHiddenFailureBash_OneLevelOnlyDoesNotExemptNestedHelper(t *testing.T) {
	checker := &HiddenFailureBash{}
	lines := []string{
		"#!/bin/bash",
		"helper() {",
		"  rm -f \"$tmp\" || true",
		"}",
		"cleanup() {",
		"  helper",
		"}",
		"trap 'cleanup' EXIT",
	}

	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for nested helper, got %d", len(findings))
	}
	if findings[0].Line != 3 {
		t.Fatalf("expected finding on line 3, got %d", findings[0].Line)
	}
}

func TestHiddenFailureBash_AnnotatedCleanupExempt(t *testing.T) {
	checker := &HiddenFailureBash{}
	lines := []string{
		"#!/bin/bash",
		"# truthsayer:cleanup-context",
		"cleanup() {",
		"  rm -f \"$tmp\" || true",
		"}",
	}

	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for annotated cleanup context, got %d", len(findings))
	}
}

func TestHiddenFailureBash_NameOnlyCleanupNotExempt(t *testing.T) {
	checker := &HiddenFailureBash{}
	lines := []string{
		"#!/bin/bash",
		"cleanup() {",
		"  rm -f \"$tmp\" || true",
		"}",
	}

	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding when only function name suggests cleanup, got %d", len(findings))
	}
}
