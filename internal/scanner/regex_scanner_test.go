package scanner

import (
	"testing"

	"github.com/perttulands/truthsayer/internal/rules"
)

func TestRegexScanner_MissingPipefail(t *testing.T) {
	s := NewRegexScanner([]rules.RegexChecker{&rules.MissingPipefail{}})

	findings, err := s.Scan(testdataPath("bash/no_pipefail.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("expected findings for missing pipefail, got none")
	}
	f := findings[0]
	if f.Rule != "bad-defaults.missing-pipefail" {
		t.Errorf("expected rule bad-defaults.missing-pipefail, got %s", f.Rule)
	}
}

func TestRegexScanner_ProperBash(t *testing.T) {
	s := NewRegexScanner([]rules.RegexChecker{&rules.MissingPipefail{}})

	findings, err := s.Scan(testdataPath("bash/proper_bash.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings for proper bash, got %d", len(findings))
	}
}
