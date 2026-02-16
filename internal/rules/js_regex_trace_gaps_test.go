package rules

import "testing"

// --- JSNoUnhandledRejection ---

func TestJSNoUnhandledRejection_Positive(t *testing.T) {
	checker := &JSNoUnhandledRejection{}
	lines := []string{
		`const express = require('express');`,
		`const app = express();`,
		`app.use(middleware);`,
		`app.listen(3000);`,
	}
	findings := checker.CheckLines("src/server.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSNoUnhandledRejection_WithHandler(t *testing.T) {
	checker := &JSNoUnhandledRejection{}
	lines := []string{
		`const express = require('express');`,
		`const app = express();`,
		`process.on('unhandledRejection', (err) => { console.error(err); process.exit(1); });`,
		`app.listen(3000);`,
	}
	findings := checker.CheckLines("src/server.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings with handler, got %d", len(findings))
	}
}

func TestJSNoUnhandledRejection_NotEntryPoint(t *testing.T) {
	checker := &JSNoUnhandledRejection{}
	lines := []string{
		`export function calculateTotal(items) {`,
		`  return items.reduce((sum, item) => sum + item.price, 0);`,
		`}`,
	}
	findings := checker.CheckLines("src/utils.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for non-entry point, got %d", len(findings))
	}
}

func TestJSNoUnhandledRejection_TestFileSkipped(t *testing.T) {
	checker := &JSNoUnhandledRejection{}
	lines := []string{
		`const app = createServer();`,
		`app.listen(3000);`,
	}
	findings := checker.CheckLines("src/server.test.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSNoUnhandledRejection_CreateServer(t *testing.T) {
	checker := &JSNoUnhandledRejection{}
	lines := []string{
		`const server = createServer(handler);`,
		`server.listen(8080);`,
	}
	findings := checker.CheckLines("src/index.ts", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

// --- JSConsoleLogInProduction ---

func TestJSConsoleLogInProduction_Positive(t *testing.T) {
	checker := &JSConsoleLogInProduction{}
	lines := []string{
		`function handleRequest(req) {`,
		`  console.log(req.body);`,
		`  return process(req);`,
		`}`,
	}
	findings := checker.CheckLines("src/handler.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 2 {
		t.Errorf("expected line 2, got %d", findings[0].Line)
	}
}

func TestJSConsoleLogInProduction_TestFileSkipped(t *testing.T) {
	checker := &JSConsoleLogInProduction{}
	lines := []string{`console.log('debugging test');`}
	findings := checker.CheckLines("src/handler.test.ts", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings in test file, got %d", len(findings))
	}
}

func TestJSConsoleLogInProduction_CommentSkipped(t *testing.T) {
	checker := &JSConsoleLogInProduction{}
	lines := []string{`// console.log('debug');`}
	findings := checker.CheckLines("src/handler.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for comment, got %d", len(findings))
	}
}

func TestJSConsoleLogInProduction_ConsoleError(t *testing.T) {
	checker := &JSConsoleLogInProduction{}
	lines := []string{`console.error('Something failed');`}
	findings := checker.CheckLines("src/handler.js", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for console.error, got %d", len(findings))
	}
}

func TestJSConsoleLogInProduction_Multiple(t *testing.T) {
	checker := &JSConsoleLogInProduction{}
	lines := []string{
		`console.log('a');`,
		`console.log('b');`,
		`console.log('c');`,
	}
	findings := checker.CheckLines("src/handler.js", lines)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
}
