package rules

import (
	"strings"
	"testing"
)

// --- PyPrintDebug ---

func TestPyPrintDebug_Positive(t *testing.T) {
	checker := &PyPrintDebug{}
	lines := strings.Split("import os\n\ndef process():\n    print(\"hello\")\n    return 42", "\n")
	findings := checker.CheckLines("src/utils.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "trace-gaps.print-debug" {
		t.Fatalf("expected rule trace-gaps.print-debug, got %s", findings[0].Rule)
	}
	if findings[0].Line != 4 {
		t.Fatalf("expected line 4, got %d", findings[0].Line)
	}
}

func TestPyPrintDebug_Multiple(t *testing.T) {
	checker := &PyPrintDebug{}
	lines := strings.Split("def process():\n    print(\"start\")\n    x = compute()\n    print(f\"result: {x}\")\n    return x", "\n")
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestPyPrintDebug_TestFileSkipped(t *testing.T) {
	checker := &PyPrintDebug{}
	lines := []string{`    print("debugging test")`}
	findings := checker.CheckLines("tests/test_main.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings in test file, got %d", len(findings))
	}
}

func TestPyPrintDebug_ScriptDirSkipped(t *testing.T) {
	checker := &PyPrintDebug{}
	lines := []string{`print("running script")`}
	findings := checker.CheckLines("scripts/deploy.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings in script file, got %d", len(findings))
	}
}

func TestPyPrintDebug_BinDirSkipped(t *testing.T) {
	checker := &PyPrintDebug{}
	lines := []string{`print("cli tool")`}
	findings := checker.CheckLines("bin/tool.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings in bin file, got %d", len(findings))
	}
}

func TestPyPrintDebug_CommentsSkipped(t *testing.T) {
	checker := &PyPrintDebug{}
	lines := []string{`# print("this is a comment")`, `def func():`, `    return 1`}
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for commented print, got %d", len(findings))
	}
}

func TestPyPrintDebug_LoggingClean(t *testing.T) {
	checker := &PyPrintDebug{}
	lines := strings.Split("import logging\nlogger = logging.getLogger(__name__)\n\ndef process():\n    logger.info(\"hello\")\n    return 42", "\n")
	findings := checker.CheckLines("src/main.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for logging, got %d", len(findings))
	}
}

// --- PyNoLoggingConfig ---

func TestPyNoLoggingConfig_Positive(t *testing.T) {
	checker := &PyNoLoggingConfig{}
	lines := strings.Split("import sys\n\ndef main():\n    process()\n\nif __name__ == \"__main__\":\n    main()", "\n")
	findings := checker.CheckLines("src/app.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "trace-gaps.no-logging-config" {
		t.Fatalf("expected rule trace-gaps.no-logging-config, got %s", findings[0].Rule)
	}
}

func TestPyNoLoggingConfig_BasicConfigClean(t *testing.T) {
	checker := &PyNoLoggingConfig{}
	lines := strings.Split("import logging\nlogging.basicConfig(level=logging.INFO)\n\nif __name__ == \"__main__\":\n    main()", "\n")
	findings := checker.CheckLines("src/app.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings with basicConfig, got %d", len(findings))
	}
}

func TestPyNoLoggingConfig_GetLoggerClean(t *testing.T) {
	checker := &PyNoLoggingConfig{}
	lines := strings.Split("import logging\nlogger = logging.getLogger(__name__)\n\nif __name__ == \"__main__\":\n    main()", "\n")
	findings := checker.CheckLines("src/app.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings with getLogger, got %d", len(findings))
	}
}

func TestPyNoLoggingConfig_NonEntryPointClean(t *testing.T) {
	checker := &PyNoLoggingConfig{}
	lines := []string{`def helper():`, `    return 42`}
	findings := checker.CheckLines("src/utils.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-entry-point, got %d", len(findings))
	}
}

func TestPyNoLoggingConfig_TestFileSkipped(t *testing.T) {
	checker := &PyNoLoggingConfig{}
	lines := strings.Split("if __name__ == \"__main__\":\n    run_tests()", "\n")
	findings := checker.CheckLines("tests/test_main.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings in test file, got %d", len(findings))
	}
}
