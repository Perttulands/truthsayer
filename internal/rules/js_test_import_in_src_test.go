package rules

import (
	"testing"
)

func TestJSTestImportInSrc(t *testing.T) {
	checker := &JSTestImportInSrc{}

	t.Run("import from @testing-library/react", func(t *testing.T) {
		src := `
import { render, screen } from '@testing-library/react';
import React from 'react';

export function App() {
  return <div>Hello</div>;
}
`
		findings := runJSCheckerOnSource(t, checker, "src/App.jsx", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "mock-leakage.js-test-import-in-src" {
			t.Errorf("wrong rule: %s", findings[0].Rule)
		}
	})

	t.Run("require vitest in source", func(t *testing.T) {
		src := `
const { describe, it } = require('vitest');
module.exports = { handler };
`
		findings := runJSCheckerOnSource(t, checker, "src/handler.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("import from jest", func(t *testing.T) {
		src := `
import { jest } from '@jest/globals';
export const mock = jest.fn();
`
		findings := runJSCheckerOnSource(t, checker, "src/utils.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("normal imports in source", func(t *testing.T) {
		src := `
import express from 'express';
import { useState } from 'react';
const lodash = require('lodash');
`
		findings := runJSCheckerOnSource(t, checker, "src/server.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("test file skipped", func(t *testing.T) {
		src := `
import { render } from '@testing-library/react';
import { describe, it } from 'vitest';
`
		findings := runJSCheckerOnSource(t, checker, "App.test.jsx", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for test file, got %d", len(findings))
		}
	})

	t.Run("@testing-library subpackages", func(t *testing.T) {
		src := `
import { fireEvent } from '@testing-library/dom';
import userEvent from '@testing-library/user-event';
`
		findings := runJSCheckerOnSource(t, checker, "src/component.js", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("@vitest/spy in source", func(t *testing.T) {
		src := `
import { spyOn } from '@vitest/spy';
`
		findings := runJSCheckerOnSource(t, checker, "src/utils.ts", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})
}
