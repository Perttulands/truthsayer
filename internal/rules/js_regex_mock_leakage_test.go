package rules

import "testing"

// --- JSJestMockInSrc ---

func TestJSJestMockInSrc_Positive(t *testing.T) {
	checker := &JSJestMockInSrc{}
	lines := []string{
		`const mock = jest.fn(() => 42);`,
		`jest.mock('./module');`,
		`const spy = jest.spyOn(obj, 'method');`,
	}
	findings := checker.CheckLines("src/utils.js", lines)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
}

func TestJSJestMockInSrc_TestFileSkipped(t *testing.T) {
	checker := &JSJestMockInSrc{}
	lines := []string{
		`jest.mock('./module');`,
		`const mock = jest.fn();`,
	}
	findings := checker.CheckLines("src/utils.test.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings in test file, got %d", len(findings))
	}
}

func TestJSJestMockInSrc_TestDirSkipped(t *testing.T) {
	checker := &JSJestMockInSrc{}
	lines := []string{`jest.mock('./module');`}
	findings := checker.CheckLines("__tests__/setup.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings in __tests__ dir, got %d", len(findings))
	}
}

func TestJSJestMockInSrc_CommentSkipped(t *testing.T) {
	checker := &JSJestMockInSrc{}
	lines := []string{`// jest.mock('./module');`}
	findings := checker.CheckLines("src/utils.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for comment, got %d", len(findings))
	}
}

func TestJSJestMockInSrc_NoMatch(t *testing.T) {
	checker := &JSJestMockInSrc{}
	lines := []string{
		`import { jest } from '@jest/globals';`,
		`const result = calculate();`,
	}
	findings := checker.CheckLines("src/utils.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for non-mock code, got %d", len(findings))
	}
}

// --- JSStorybookInSrc ---

func TestJSStorybookInSrc_Positive(t *testing.T) {
	checker := &JSStorybookInSrc{}
	lines := []string{
		`import { Default } from './Button.stories.tsx';`,
	}
	findings := checker.CheckLines("src/App.tsx", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSStorybookInSrc_Require(t *testing.T) {
	checker := &JSStorybookInSrc{}
	lines := []string{
		`const stories = require('./Button.stories.js');`,
	}
	findings := checker.CheckLines("src/utils.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSStorybookInSrc_StoryFileSkipped(t *testing.T) {
	checker := &JSStorybookInSrc{}
	lines := []string{
		`import { Default } from './Button.stories.tsx';`,
	}
	findings := checker.CheckLines("src/Button.stories.tsx", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings in story file, got %d", len(findings))
	}
}

func TestJSStorybookInSrc_TestFileSkipped(t *testing.T) {
	checker := &JSStorybookInSrc{}
	lines := []string{
		`import { Default } from './Button.stories.tsx';`,
	}
	findings := checker.CheckLines("src/Button.test.tsx", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings in test file, got %d", len(findings))
	}
}

func TestJSStorybookInSrc_NoMatch(t *testing.T) {
	checker := &JSStorybookInSrc{}
	lines := []string{
		`import { Button } from './Button';`,
		`import utils from '../utils';`,
	}
	findings := checker.CheckLines("src/App.tsx", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for non-story imports, got %d", len(findings))
	}
}
