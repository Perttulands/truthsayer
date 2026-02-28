package cli

import "testing"

func TestCheck_ExitCode2_NoArgs(t *testing.T) {
	code := runCheck(nil)
	if code != 2 {
		t.Errorf("expected exit code 2 for missing file arg, got %d", code)
	}
}

func TestCheck_ExitCode2_DirectoryArg(t *testing.T) {
	dir := t.TempDir()
	code := runCheck([]string{dir})
	if code != 2 {
		t.Errorf("expected exit code 2 for directory arg to check, got %d", code)
	}
}

func TestCheck_ExitCode2_NonexistentFile(t *testing.T) {
	code := runCheck([]string{"/nonexistent/file.go"})
	if code != 2 {
		t.Errorf("expected exit code 2 for nonexistent file, got %d", code)
	}
}

func TestCheck_ExitCode2_BadConfig(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "ok.go", goClean)
	code := runCheck([]string{"--config", "/nonexistent/config.toml", path})
	if code != 2 {
		t.Errorf("expected exit code 2 for bad config, got %d", code)
	}
}
