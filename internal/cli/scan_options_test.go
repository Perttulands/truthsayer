package cli

import "testing"

func TestParseScanOptions_FormatFlag(t *testing.T) {
	opts, err := parseScanOptions([]string{"--format", "json", "/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.format != "json" {
		t.Errorf("expected format json, got %q", opts.format)
	}
	if opts.path != "/path" {
		t.Errorf("expected path /path, got %q", opts.path)
	}
}

func TestParseScanOptions_FormatMissingValue(t *testing.T) {
	_, err := parseScanOptions([]string{"--format"})
	if err == nil {
		t.Fatal("expected error for --format without value")
	}
}

func TestParseScanOptions_LangMissingValue(t *testing.T) {
	_, err := parseScanOptions([]string{"--lang"})
	if err == nil {
		t.Fatal("expected error for --lang without value")
	}
}

func TestParseScanOptions_UnknownFormat(t *testing.T) {
	_, err := parseScanOptions([]string{"--format", "xml", "/path"})
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestParseScanOptions_ParallelBareFlag(t *testing.T) {
	opts, err := parseScanOptions([]string{"--parallel", "/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.parallel < 1 {
		t.Errorf("expected parallel >= 1, got %d", opts.parallel)
	}
	if opts.path != "/path" {
		t.Errorf("expected path /path, got %q", opts.path)
	}
}

func TestParseScanOptions_ParallelWithValue(t *testing.T) {
	opts, err := parseScanOptions([]string{"--parallel", "4", "/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.parallel != 4 {
		t.Errorf("expected parallel 4, got %d", opts.parallel)
	}
}

func TestParseScanOptions_ParallelZeroErrors(t *testing.T) {
	_, err := parseScanOptions([]string{"--parallel", "0", "/path"})
	if err == nil {
		t.Fatal("expected error for parallel=0")
	}
}

func TestParseScanOptions_BeadThresholdMissingValue(t *testing.T) {
	_, err := parseScanOptions([]string{"--bead-threshold"})
	if err == nil {
		t.Fatal("expected error for --bead-threshold without value")
	}
}

func TestParseScanOptions_BeadThresholdInvalid(t *testing.T) {
	_, err := parseScanOptions([]string{"--bead-threshold", "abc"})
	if err == nil {
		t.Fatal("expected error for invalid bead threshold")
	}
}

func TestParseScanOptions_BeadThresholdNegative(t *testing.T) {
	_, err := parseScanOptions([]string{"--bead-threshold", "-1"})
	if err == nil {
		t.Fatal("expected error for negative bead threshold")
	}
}

func TestParseScanOptions_CreateBeadsFlag(t *testing.T) {
	opts, err := parseScanOptions([]string{"--create-beads", "/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.createBeads {
		t.Error("expected createBeads to be true")
	}
}

func TestParseScanOptions_UsePrecedentsFlagSetsField(t *testing.T) {
	opts, err := parseScanOptions([]string{"--use-precedents", "/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.usePrecedents {
		t.Error("expected usePrecedents to be true")
	}
}

func TestParseScanOptions_LangFlagSetsField(t *testing.T) {
	opts, err := parseScanOptions([]string{"--lang", "go,python", "/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.lang != "go,python" {
		t.Errorf("expected lang go,python, got %q", opts.lang)
	}
}

func TestScan_ExitCode2_FileNotDir(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "file.go", goClean)
	code := runScan([]string{path})
	if code != 2 {
		t.Errorf("expected exit code 2 for file arg to scan, got %d", code)
	}
}

func TestScan_ExitCode2_InvalidLang(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "ok.go", goClean)
	code := runScan([]string{"--lang", "fortran", dir})
	if code != 2 {
		t.Errorf("expected exit code 2 for invalid lang, got %d", code)
	}
}

func TestScan_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.go", goWithError)

	out := captureStdout(t, func() {
		code := runScan([]string{"--format", "json", dir})
		if code != 1 {
			t.Fatalf("expected exit code 1, got %d", code)
		}
	})

	if len(out) == 0 {
		t.Fatal("expected JSON output, got empty")
	}
	// Should be valid JSON
	if out[0] != '{' {
		t.Fatalf("expected JSON output starting with '{', got %q", out[:1])
	}
}
