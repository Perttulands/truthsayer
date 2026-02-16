package rules

import (
	"testing"
)

func TestPyStarImport(t *testing.T) {
	checker := &PyStarImport{}

	t.Run("triggers on wildcard import", func(t *testing.T) {
		src := `from os.path import *
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "bad-defaults.py-star-import" {
			t.Errorf("expected rule bad-defaults.py-star-import, got %s", findings[0].Rule)
		}
	})

	t.Run("triggers on multiple wildcard imports", func(t *testing.T) {
		src := `from os.path import *
from collections import *
from typing import *
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 3 {
			t.Fatalf("expected 3 findings, got %d", len(findings))
		}
	})

	t.Run("clean on explicit imports", func(t *testing.T) {
		src := `from os.path import join, exists
from collections import defaultdict
import os
import sys
`
		findings := runPyCheckerOnSource(t, checker, "app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("skips test files", func(t *testing.T) {
		src := `from mymodule import *
`
		findings := runPyCheckerOnSource(t, checker, "test_app.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for test file, got %d", len(findings))
		}
	})

	t.Run("skips test directory files", func(t *testing.T) {
		src := `from mymodule import *
`
		findings := runPyCheckerOnSource(t, checker, "tests/conftest.py", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for test dir file, got %d", len(findings))
		}
	})
}
