package rules

import "testing"

func TestNoErrTrap_SetENoTrap(t *testing.T) {
	checker := &NoErrTrap{}
	lines := []string{
		"#!/bin/bash",
		"set -euo pipefail",
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

func TestNoErrTrap_SetEWithTrap(t *testing.T) {
	checker := &NoErrTrap{}
	lines := []string{
		"#!/bin/bash",
		"set -euo pipefail",
		`trap 'echo "Error at $LINENO" >&2' ERR`,
		"echo hello",
	}
	findings := checker.CheckLines("deploy.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings with trap ERR, got %d", len(findings))
	}
}

func TestNoErrTrap_NoSetE(t *testing.T) {
	checker := &NoErrTrap{}
	lines := []string{
		"#!/bin/bash",
		"echo hello",
	}
	findings := checker.CheckLines("deploy.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings without set -e, got %d", len(findings))
	}
}

func TestNoErrTrap_SkipsTestDir(t *testing.T) {
	checker := &NoErrTrap{}
	lines := []string{
		"#!/bin/bash",
		"set -e",
		"echo test",
	}
	findings := checker.CheckLines("project/tests/run.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test dir, got %d", len(findings))
	}
}

func TestNoErrTrap_SetEInComment(t *testing.T) {
	checker := &NoErrTrap{}
	lines := []string{
		"#!/bin/bash",
		"# set -e",
		"echo hello",
	}
	findings := checker.CheckLines("deploy.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when set -e is in comment, got %d", len(findings))
	}
}
