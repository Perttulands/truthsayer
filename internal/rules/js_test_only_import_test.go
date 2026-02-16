package rules

import (
	"testing"
)

func TestJSTestOnlyImport(t *testing.T) {
	checker := &JSTestOnlyImport{}

	t.Run("import from __tests__ directory", func(t *testing.T) {
		src := `import { createMockUser } from '../__tests__/helpers';`
		findings := runJSCheckerOnSource(t, checker, "src/user.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "test-isolation.test-only-import" {
			t.Errorf("expected rule test-isolation.test-only-import, got %s", findings[0].Rule)
		}
	})

	t.Run("import from __mocks__ directory", func(t *testing.T) {
		src := `import { mockDB } from './__mocks__/database';`
		findings := runJSCheckerOnSource(t, checker, "src/db.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("import test-utils module", func(t *testing.T) {
		src := `import { render } from './test-utils';`
		findings := runJSCheckerOnSource(t, checker, "src/component.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("import test helper module", func(t *testing.T) {
		src := `import { setup } from './test-helpers';`
		findings := runJSCheckerOnSource(t, checker, "src/app.ts", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("import module with .test suffix", func(t *testing.T) {
		src := `import { factory } from './user.test';`
		findings := runJSCheckerOnSource(t, checker, "src/api.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("require from __fixtures__", func(t *testing.T) {
		src := `const fixtures = require('./__fixtures__/data');`
		findings := runJSCheckerOnSource(t, checker, "src/loader.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("require mock module", func(t *testing.T) {
		src := `const mock = require('./mock-database');`
		findings := runJSCheckerOnSource(t, checker, "src/service.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("normal import - clean", func(t *testing.T) {
		src := `import { useState } from 'react';
import { db } from './database';
import { utils } from '../lib/utils';
`
		findings := runJSCheckerOnSource(t, checker, "src/app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("normal require - clean", func(t *testing.T) {
		src := `const express = require('express');
const db = require('./database');
`
		findings := runJSCheckerOnSource(t, checker, "src/server.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("test file imports test utils - clean", func(t *testing.T) {
		src := `import { render } from './test-utils';
import { createMockUser } from '../__tests__/helpers';
`
		findings := runJSCheckerOnSource(t, checker, "src/app.test.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for test file, got %d", len(findings))
		}
	})

	t.Run("multiple test imports in source", func(t *testing.T) {
		src := `import { setup } from './test-helpers';
import { mockDB } from './__mocks__/database';
import { utils } from '../lib/utils';
`
		findings := runJSCheckerOnSource(t, checker, "src/init.js", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("import from stubs directory", func(t *testing.T) {
		src := `import { stubApi } from './stubs';`
		findings := runJSCheckerOnSource(t, checker, "src/client.ts", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})
}
