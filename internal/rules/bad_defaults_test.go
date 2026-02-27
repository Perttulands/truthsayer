package rules

import "testing"

func TestMissingPipefail_NoPipefail(t *testing.T) {
	checker := &MissingPipefail{}
	lines := []string{
		"#!/bin/bash",
		"echo hello",
	}
	findings := checker.CheckLines("deploy.sh", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestMissingPipefail_HasPipefail(t *testing.T) {
	checker := &MissingPipefail{}
	lines := []string{
		"#!/bin/bash",
		"set -euo pipefail",
		"echo hello",
	}
	findings := checker.CheckLines("deploy.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestMissingPipefail_NotBashShebang(t *testing.T) {
	checker := &MissingPipefail{}
	lines := []string{
		"#!/bin/sh",
		"echo hello",
	}
	findings := checker.CheckLines("deploy.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-bash shebang, got %d", len(findings))
	}
}

func TestMissingPipefail_EmptyFile(t *testing.T) {
	checker := &MissingPipefail{}
	findings := checker.CheckLines("empty.sh", nil)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for empty file, got %d", len(findings))
	}
}

func TestMissingPipefail_PipefailInComment(t *testing.T) {
	checker := &MissingPipefail{}
	lines := []string{
		"#!/bin/bash",
		"# set -euo pipefail",
		"echo hello",
	}
	findings := checker.CheckLines("deploy.sh", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding when pipefail is in comment, got %d", len(findings))
	}
}
