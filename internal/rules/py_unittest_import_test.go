package rules

import (
	"testing"
)

func TestPyUnittestImport(t *testing.T) {
	checker := &PyUnittestImport{}

	t.Run("detects from unittest.mock import", func(t *testing.T) {
		src := `from unittest.mock import patch, MagicMock
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "mock-leakage.py-unittest-import" {
			t.Errorf("expected rule mock-leakage.py-unittest-import, got %s", findings[0].Rule)
		}
	})

	t.Run("detects from unittest import mock", func(t *testing.T) {
		src := `from unittest import mock
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("detects import mock", func(t *testing.T) {
		src := `import mock
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("detects import unittest.mock", func(t *testing.T) {
		src := `import unittest.mock
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("detects import mock as mk", func(t *testing.T) {
		src := `import mock as mk
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("detects from pytest import fixture", func(t *testing.T) {
		src := `from pytest import fixture
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		src := `from unittest.mock import patch
import mock
`
		findings := runPyCheckerOnSource(t, checker, "tests/test_service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings in test file, got %d", len(findings))
		}
	})

	t.Run("clean on regular imports", func(t *testing.T) {
		src := `import os
import sys
from pathlib import Path
from collections import defaultdict
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("no false positive on module containing mock", func(t *testing.T) {
		src := `import mockery
from mocking import helper
`
		findings := runPyCheckerOnSource(t, checker, "app/service.py", src)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})
}
