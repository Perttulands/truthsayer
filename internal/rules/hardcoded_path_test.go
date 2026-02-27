package rules

import "testing"

func TestHardcodedPath_LinuxHome(t *testing.T) {
	checker := &HardcodedPath{}
	lines := []string{
		`DATA_DIR=/home/deploy/app/data`,
	}
	findings := checker.CheckLines("config.toml", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestHardcodedPath_MacUsers(t *testing.T) {
	checker := &HardcodedPath{}
	lines := []string{
		`path = "/Users/john/projects/app"`,
	}
	findings := checker.CheckLines("config.yaml", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for macOS path, got %d", len(findings))
	}
}

func TestHardcodedPath_NoMatch(t *testing.T) {
	checker := &HardcodedPath{}
	lines := []string{
		`DATA_DIR=$HOME/app/data`,
		`path = "/var/log/app.log"`,
	}
	findings := checker.CheckLines("config.toml", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestHardcodedPath_CommentSkipped(t *testing.T) {
	checker := &HardcodedPath{}
	lines := []string{
		`# /home/user/some/path`,
		`// /home/user/other/path`,
	}
	findings := checker.CheckLines("config.sh", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for comments, got %d", len(findings))
	}
}
