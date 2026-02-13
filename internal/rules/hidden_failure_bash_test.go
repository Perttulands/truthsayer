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
