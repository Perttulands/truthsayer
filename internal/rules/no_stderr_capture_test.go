package rules

import "testing"

func TestNoStderrCapture_FindsMissingCapture(t *testing.T) {
	checker := &NoStderrCapture{}
	lines := []string{
		"cmd := exec.Command(\"ls\")",
		"_ = cmd.Run()",
	}

	findings := checker.CheckLines("main.go", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 1 {
		t.Errorf("expected finding on line 1, got %d", findings[0].Line)
	}
}

func TestNoStderrCapture_AllowsStderrPipe(t *testing.T) {
	checker := &NoStderrCapture{}
	lines := []string{
		"cmd := exec.Command(\"ls\")",
		"stderr, _ := cmd.StderrPipe()",
		"_ = stderr",
	}

	findings := checker.CheckLines("main.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestNoStderrCapture_AllowsCombinedOutput(t *testing.T) {
	checker := &NoStderrCapture{}
	lines := []string{
		"out, err := exec.Command(\"ls\").CombinedOutput()",
		"_, _ = out, err",
	}

	findings := checker.CheckLines("main.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestNoStderrCapture_AllowsStderrAssignment(t *testing.T) {
	checker := &NoStderrCapture{}
	lines := []string{
		"cmd := exec.Command(\"ls\")",
		"cmd.Stderr = os.Stderr",
		"_ = cmd.Run()",
	}

	findings := checker.CheckLines("main.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestNoStderrCapture_IgnoresCommentedCode(t *testing.T) {
	checker := &NoStderrCapture{}
	lines := []string{
		"// cmd := exec.Command(\"ls\")",
	}

	findings := checker.CheckLines("main.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
