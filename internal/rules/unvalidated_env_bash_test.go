package rules

import "testing"

func TestUnvalidatedEnvBash_DefaultValue(t *testing.T) {
	checker := &UnvalidatedEnvBash{}
	lines := []string{
		"#!/bin/bash",
		`MODEL="${MODEL:-sonnet}"`,
	}
	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestUnvalidatedEnvBash_Required(t *testing.T) {
	checker := &UnvalidatedEnvBash{}
	lines := []string{
		"#!/bin/bash",
		`MODEL="${MODEL:?"MODEL is required"}"`,
	}
	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for :? pattern, got %d", len(findings))
	}
}

func TestUnvalidatedEnvBash_Comment(t *testing.T) {
	checker := &UnvalidatedEnvBash{}
	lines := []string{
		"#!/bin/bash",
		`# MODEL="${MODEL:-sonnet}"`,
	}
	findings := checker.CheckLines("test.sh", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for comment, got %d", len(findings))
	}
}
