package rules

import "testing"

func TestPySubprocessNoCheck_Run(t *testing.T) {
	src := `
import subprocess
subprocess.run(["ls", "-la"])
`
	findings := runPyCheckerOnSource(t, &PySubprocessNoCheck{}, "deploy.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.py-subprocess-no-check" {
		t.Errorf("wrong rule: %s", findings[0].Rule)
	}
}

func TestPySubprocessNoCheck_Call(t *testing.T) {
	src := `
import subprocess
subprocess.call(["make", "build"])
`
	findings := runPyCheckerOnSource(t, &PySubprocessNoCheck{}, "build.py", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestPySubprocessNoCheck_NegativeCheckTrue(t *testing.T) {
	src := `
import subprocess
subprocess.run(["ls", "-la"], check=True)
`
	findings := runPyCheckerOnSource(t, &PySubprocessNoCheck{}, "deploy.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPySubprocessNoCheck_NegativeCheckTrueWithOtherArgs(t *testing.T) {
	src := `
import subprocess
subprocess.run(["ls"], capture_output=True, check=True, timeout=30)
`
	findings := runPyCheckerOnSource(t, &PySubprocessNoCheck{}, "deploy.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestPySubprocessNoCheck_NegativeCheckOutput(t *testing.T) {
	src := `
import subprocess
subprocess.check_output(["ls"])
subprocess.check_call(["ls"])
`
	findings := runPyCheckerOnSource(t, &PySubprocessNoCheck{}, "deploy.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for check_output/check_call, got %d", len(findings))
	}
}

func TestPySubprocessNoCheck_NegativeOtherModule(t *testing.T) {
	src := `
import os
os.run(["ls"])
`
	findings := runPyCheckerOnSource(t, &PySubprocessNoCheck{}, "deploy.py", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-subprocess, got %d", len(findings))
	}
}

func TestPySubprocessNoCheck_Multiple(t *testing.T) {
	src := `
import subprocess
subprocess.run(["make", "clean"])
subprocess.run(["make", "build"], check=True)
subprocess.call(["deploy.sh"])
`
	findings := runPyCheckerOnSource(t, &PySubprocessNoCheck{}, "build.py", src)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}
