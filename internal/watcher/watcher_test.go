package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew_InvalidPath(t *testing.T) {
	_, err := New("/nonexistent/path", 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestNew_ValidPath(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Close()
}

func TestEvents_FileCreate(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Close()

	events := w.Events()

	// Create a Go file — should trigger an event
	path := filepath.Join(dir, "test.go")
	if err := os.WriteFile(path, []byte("package main"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	select {
	case got := <-events:
		if got != path {
			t.Fatalf("expected %s, got %s", path, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for file event")
	}
}

func TestEvents_FileWrite(t *testing.T) {
	dir := t.TempDir()

	// Create file before starting watcher
	path := filepath.Join(dir, "test.sh")
	if err := os.WriteFile(path, []byte("#!/bin/bash"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	w, err := New(dir, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Close()

	events := w.Events()

	// Modify the file
	if err := os.WriteFile(path, []byte("#!/bin/bash\necho hi"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	select {
	case got := <-events:
		if got != path {
			t.Fatalf("expected %s, got %s", path, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for write event")
	}
}

func TestEvents_IgnoresUnsupportedExt(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Close()

	events := w.Events()

	// Create a .txt file — should NOT trigger
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a .go file — should trigger
	goPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goPath, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-events:
		if got != goPath {
			t.Fatalf("expected %s, got %s", goPath, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for .go event")
	}
}

func TestEvents_Debounce(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Close()

	events := w.Events()

	path := filepath.Join(dir, "test.go")

	// Write rapidly 5 times within debounce window
	for i := range 5 {
		if err := os.WriteFile(path, []byte("package main // "+string(rune('a'+i))), 0644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Should get exactly one event after debounce
	select {
	case got := <-events:
		if got != path {
			t.Fatalf("expected %s, got %s", path, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for debounced event")
	}

	// Should NOT get a second event immediately
	select {
	case extra := <-events:
		t.Fatalf("unexpected second event: %s", extra)
	case <-time.After(500 * time.Millisecond):
		// OK — no extra events
	}
}
