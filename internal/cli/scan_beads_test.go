package cli

import (
	"strings"
	"testing"
)

type fakeBeadCreator struct {
	calls []beadCall
}

type beadCall struct {
	rule  string
	file  string
	count int
}

func (f *fakeBeadCreator) CreateProblemBead(rule string, file string, count int) (string, error) {
	f.calls = append(f.calls, beadCall{
		rule:  rule,
		file:  file,
		count: count,
	})
	return "bd-test-1", nil
}

func TestScan_CreateBeadsFlag_CreatesProblemBeads(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.go", goWithError)

	fake := &fakeBeadCreator{}
	oldFactory := newProblemBeadCreator
	newProblemBeadCreator = func() problemBeadCreator { return fake }
	defer func() { newProblemBeadCreator = oldFactory }()

	out := captureStdout(t, func() {
		code := runScan([]string{"--create-beads", dir})
		if code != 1 {
			t.Fatalf("expected exit code 1, got %d", code)
		}
	})

	if len(fake.calls) != 1 {
		t.Fatalf("expected exactly 1 bead creation call, got %d", len(fake.calls))
	}
	if fake.calls[0].count != 1 {
		t.Fatalf("expected grouped error count 1, got %d", fake.calls[0].count)
	}
	if !strings.Contains(out, "Beads created: 1") {
		t.Fatalf("expected bead summary in output, got:\n%s", out)
	}
	if !strings.Contains(out, "bd-test-1") {
		t.Fatalf("expected bead ID in output, got:\n%s", out)
	}
}

func TestScan_BeadThreshold_SuppressesSmallGroups(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.go", goWithError)

	fake := &fakeBeadCreator{}
	oldFactory := newProblemBeadCreator
	newProblemBeadCreator = func() problemBeadCreator { return fake }
	defer func() { newProblemBeadCreator = oldFactory }()

	out := captureStdout(t, func() {
		code := runScan([]string{"--create-beads", "--bead-threshold", "1", dir})
		if code != 1 {
			t.Fatalf("expected exit code 1, got %d", code)
		}
	})

	if len(fake.calls) != 0 {
		t.Fatalf("expected 0 bead calls with threshold=1, got %d", len(fake.calls))
	}
	if !strings.Contains(out, "Beads created: 0") {
		t.Fatalf("expected zero bead summary in output, got:\n%s", out)
	}
}
