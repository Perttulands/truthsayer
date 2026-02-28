package cli

import "testing"

func TestParseCIOptions_DefaultPath(t *testing.T) {
	opts, err := parseCIOptions(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.path != "." {
		t.Errorf("expected default path '.', got %q", opts.path)
	}
}

func TestParseCIOptions_CustomPath(t *testing.T) {
	opts, err := parseCIOptions([]string{"/repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.path != "/repo" {
		t.Errorf("expected path '/repo', got %q", opts.path)
	}
}

func TestParseCIOptions_BeadThresholdValid(t *testing.T) {
	opts, err := parseCIOptions([]string{"--bead-threshold", "5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.beadThreshold != 5 {
		t.Errorf("expected bead threshold 5, got %d", opts.beadThreshold)
	}
}

func TestParseCIOptions_BeadThresholdMissing(t *testing.T) {
	_, err := parseCIOptions([]string{"--bead-threshold"})
	if err == nil {
		t.Fatal("expected error for missing --bead-threshold value")
	}
}

func TestParseCIOptions_BeadThresholdInvalid(t *testing.T) {
	_, err := parseCIOptions([]string{"--bead-threshold", "abc"})
	if err == nil {
		t.Fatal("expected error for invalid --bead-threshold value")
	}
}

func TestParseCIOptions_BeadThresholdNegative(t *testing.T) {
	_, err := parseCIOptions([]string{"--bead-threshold", "-1"})
	if err == nil {
		t.Fatal("expected error for negative --bead-threshold value")
	}
}

func TestParseCIOptions_CreateBeadsNoOp(t *testing.T) {
	// --create-beads is accepted for compatibility but does nothing
	opts, err := parseCIOptions([]string{"--create-beads", "/repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.path != "/repo" {
		t.Errorf("expected path '/repo', got %q", opts.path)
	}
}
