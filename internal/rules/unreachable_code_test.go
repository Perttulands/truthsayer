package rules

import "testing"

func TestUnreachableCode_SkipsGoRawStringContent(t *testing.T) {
	checker := &UnreachableCode{}
	lines := []string{
		`func buildPrompt() string {`,
		"	return fmt.Sprintf(`",
		`CURRENT AGENT (version %d):`,
		`System Prompt: %s`,
		`User Message: %s`,
		"`, version, prompt, msg)",
		`}`,
	}
	findings := checker.CheckLines("test.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings inside raw string, got %d", len(findings))
	}
}

func TestUnreachableCode_StillDetectsRealUnreachable(t *testing.T) {
	checker := &UnreachableCode{}
	lines := []string{
		`func foo() {`,
		`	return`,
		`	x := 1`,
		`}`,
	}
	findings := checker.CheckLines("test.go", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 3 {
		t.Errorf("expected line 3, got %d", findings[0].Line)
	}
}

func TestUnreachableCode_NonGoFileUnaffected(t *testing.T) {
	checker := &UnreachableCode{}
	lines := []string{
		`function foo() {`,
		`	return;`,
		`	const x = 1;`,
		`}`,
	}
	findings := checker.CheckLines("test.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for JS, got %d", len(findings))
	}
}

func TestUnreachableCode_MultipleRawStringsInFile(t *testing.T) {
	checker := &UnreachableCode{}
	lines := []string{
		`func a() string {`,
		"	return `",
		`return something`,
		`break here`,
		"`",
		`}`,
		``,
		`func b() string {`,
		"	return `",
		`continue loop`,
		`panic now`,
		"`",
		`}`,
	}
	findings := checker.CheckLines("test.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings inside raw strings, got %d", len(findings))
	}
}

func TestUnreachableCode_SwitchCaseNotUnreachable(t *testing.T) {
	checker := &UnreachableCode{}
	lines := []string{
		`func run() int {`,
		`	switch cmd {`,
		`	case "scan":`,
		`		return runScan()`,
		`	case "check":`,
		`		return runCheck()`,
		`	case "watch":`,
		`		return runWatch()`,
		`	default:`,
		`		return 1`,
		`	}`,
		`}`,
	}
	findings := checker.CheckLines("test.go", lines)
	if len(findings) != 0 {
		for _, f := range findings {
			t.Logf("unexpected finding at line %d: %s", f.Line, f.Code)
		}
		t.Fatalf("expected 0 findings in switch-case, got %d", len(findings))
	}
}

func TestUnreachableCode_MultiLineReturnNotUnreachable(t *testing.T) {
	checker := &UnreachableCode{}
	lines := []string{
		`func cost() float64 {`,
		`	return inputCost +`,
		`		outputCost`,
		`}`,
	}
	findings := checker.CheckLines("test.go", lines)
	if len(findings) != 0 {
		for _, f := range findings {
			t.Logf("unexpected finding at line %d: %s", f.Line, f.Code)
		}
		t.Fatalf("expected 0 findings for multi-line return, got %d", len(findings))
	}
}

func TestUnreachableCode_MultiLineReturnOrChain(t *testing.T) {
	checker := &UnreachableCode{}
	lines := []string{
		`func shouldRetry(status int) bool {`,
		`	return status == 429 ||`,
		`		status == 500 ||`,
		`		status == 502 ||`,
		`		status == 503`,
		`}`,
	}
	findings := checker.CheckLines("test.go", lines)
	if len(findings) != 0 {
		for _, f := range findings {
			t.Logf("unexpected finding at line %d: %s", f.Line, f.Code)
		}
		t.Fatalf("expected 0 findings for multi-line || return, got %d", len(findings))
	}
}

func TestGoRawStringLines(t *testing.T) {
	lines := []string{
		`code here`,           // 0: false
		"return fmt.Sprintf(`", // 1: false (opens raw string)
		`inside raw string`,   // 2: true
		`more inside`,         // 3: true
		"`, arg1, arg2)",      // 4: true (closes raw string)
		`code again`,          // 5: false
	}
	result := goRawStringLines(lines)
	expected := []bool{false, false, true, true, true, false}
	for i, want := range expected {
		if result[i] != want {
			t.Errorf("line %d: got %v, want %v", i, result[i], want)
		}
	}
}
