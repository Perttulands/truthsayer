package diff

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTracker_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644)

	tr := NewTracker()
	changed, err := tr.Update(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// New file: all lines are "changed"
	if len(changed) != 3 {
		t.Fatalf("expected 3 changed lines, got %d: %v", len(changed), changed)
	}
	for _, line := range []int{1, 2, 3} {
		if !changed[line] {
			t.Errorf("expected line %d to be changed", line)
		}
	}
}

func TestTracker_ModifiedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")

	// Initial content
	os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644)
	tr := NewTracker()
	tr.Update(path) // snapshot

	// Modify line 2 only
	os.WriteFile(path, []byte("line1\nMODIFIED\nline3\n"), 0644)
	changed, err := tr.Update(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !changed[2] {
		t.Error("expected line 2 to be changed")
	}
	if changed[1] || changed[3] {
		t.Error("lines 1 and 3 should not be changed")
	}
}

func TestTracker_AddedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")

	// Initial: 2 lines
	os.WriteFile(path, []byte("line1\nline2\n"), 0644)
	tr := NewTracker()
	tr.Update(path)

	// Add a line in the middle
	os.WriteFile(path, []byte("line1\nNEW\nline2\n"), 0644)
	changed, err := tr.Update(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Line 2 is new, line 3 shifted from old line 2
	if !changed[2] {
		t.Error("expected line 2 (inserted) to be changed")
	}
}

func TestTracker_DeletedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")

	// Initial: 3 lines
	os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644)
	tr := NewTracker()
	tr.Update(path)

	// Delete line 2
	os.WriteFile(path, []byte("line1\nline3\n"), 0644)
	changed, err := tr.Update(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Deletion doesn't add changed lines (no new content to report)
	// But line 2 is now "line3" which moved — not a "changed" line in the new file
	// The key point: no false positives on lines that didn't change content-wise
	if changed[1] {
		t.Error("line 1 should not be changed")
	}
}

func TestTracker_NoChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")

	os.WriteFile(path, []byte("line1\nline2\n"), 0644)
	tr := NewTracker()
	tr.Update(path)

	// Same content written again
	os.WriteFile(path, []byte("line1\nline2\n"), 0644)
	changed, err := tr.Update(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(changed) != 0 {
		t.Fatalf("expected no changed lines, got %d: %v", len(changed), changed)
	}
}

func TestTracker_FileNotFound(t *testing.T) {
	tr := NewTracker()
	_, err := tr.Update("/nonexistent/file.go")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
