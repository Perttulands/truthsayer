package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

var (
	testListenPattern      = regexp.MustCompile(`\.listen\s*\(`)
	testEventSourcePattern = regexp.MustCompile(`\bnew\s+(?:window\.)?EventSource\s*\(`)
	testBeforeHookPattern  = regexp.MustCompile(`\bbefore(?:Each)?\s*\(`)
	testAfterHookPattern   = regexp.MustCompile(`\bafter(?:Each)?\s*\(`)
	testCleanupPattern     = regexp.MustCompile(`\bcleanup\s*\(`)
	testClosePattern       = regexp.MustCompile(`\.close\s*\(`)
	testAbortPattern       = regexp.MustCompile(`\.abort\s*\(`)
	testResourcePattern    = regexp.MustCompile(`(?:\.listen\s*\(|\bnew\s+(?:window\.)?EventSource\s*\(|\bsetInterval\s*\(|\bcreateConnection\s*\(|\bconnect\s*\(|\bnew\s+\w*Connection\b)`)
)

var testIsolationFileTypes = []string{
	".go",
	".js",
	".jsx",
	".ts",
	".tsx",
	".mjs",
	".cjs",
	".sh",
	".bash",
}

// TestLeakedServer detects test servers created with .listen() without cleanup close hooks.
type TestLeakedServer struct{}

func (t *TestLeakedServer) Meta() Rule {
	return Rule{
		ID:          "test-isolation.test-leaked-server",
		Category:    "test-isolation",
		Name:        "Leaked test server",
		Description: "Test file starts server with .listen() but does not close it in after/afterEach/cleanup",
		Severity:    finding.SeverityWarning,
		FileTypes:   testIsolationFileTypes,
		ScanType:    ScanTypeRegex,
	}
}

func (t *TestLeakedServer) CheckLines(path string, lines []string) []finding.Finding {
	if !isTestIsolationFile(path) {
		return nil
	}
	if hasCleanupCallInHooks(lines, testClosePattern) {
		return nil
	}

	var findings []finding.Finding
	for i, line := range lines {
		if !testListenPattern.MatchString(line) {
			continue
		}
		findings = append(findings, finding.Finding{
			Rule:       t.Meta().ID,
			Severity:   t.Meta().Severity,
			File:       path,
			Line:       i + 1,
			Code:       line,
			Message:    "Test server started with .listen() without cleanup close hook",
			Suggestion: "Close the server in after()/afterEach() (or cleanup) to prevent leaked handles between tests",
		})
	}
	return findings
}

// TestLeakedSSE detects EventSource usage in tests without cleanup close/abort hooks.
type TestLeakedSSE struct{}

func (t *TestLeakedSSE) Meta() Rule {
	return Rule{
		ID:          "test-isolation.test-leaked-sse",
		Category:    "test-isolation",
		Name:        "Leaked SSE connection in tests",
		Description: "Test file creates EventSource/SSE connection without close or abort in cleanup hooks",
		Severity:    finding.SeverityWarning,
		FileTypes:   testIsolationFileTypes,
		ScanType:    ScanTypeRegex,
	}
}

func (t *TestLeakedSSE) CheckLines(path string, lines []string) []finding.Finding {
	if !isTestIsolationFile(path) {
		return nil
	}
	if hasCleanupCallInHooks(lines, testClosePattern, testAbortPattern) {
		return nil
	}

	var findings []finding.Finding
	for i, line := range lines {
		if !testEventSourcePattern.MatchString(line) {
			continue
		}
		findings = append(findings, finding.Finding{
			Rule:       t.Meta().ID,
			Severity:   t.Meta().Severity,
			File:       path,
			Line:       i + 1,
			Code:       line,
			Message:    "EventSource/SSE connection created in test without close/abort cleanup",
			Suggestion: "Close or abort SSE connections in after()/afterEach()/cleanup to avoid hanging test runs",
		})
	}
	return findings
}

// TestMissingCleanup detects before/beforeEach resource setup without after/afterEach teardown hooks.
type TestMissingCleanup struct{}

func (t *TestMissingCleanup) Meta() Rule {
	return Rule{
		ID:          "test-isolation.test-missing-cleanup",
		Category:    "test-isolation",
		Name:        "Missing test cleanup hook",
		Description: "before/beforeEach creates resources but file has no after/afterEach cleanup hook",
		Severity:    finding.SeverityInfo,
		FileTypes:   testIsolationFileTypes,
		ScanType:    ScanTypeRegex,
	}
}

func (t *TestMissingCleanup) CheckLines(path string, lines []string) []finding.Finding {
	if !isTestIsolationFile(path) {
		return nil
	}
	if hasAfterHook(lines) {
		return nil
	}

	var findings []finding.Finding
	for i, line := range lines {
		if !testBeforeHookPattern.MatchString(line) {
			continue
		}
		if !hookContainsPattern(lines, i, testBeforeHookPattern, testResourcePattern) {
			continue
		}
		findings = append(findings, finding.Finding{
			Rule:       t.Meta().ID,
			Severity:   t.Meta().Severity,
			File:       path,
			Line:       i + 1,
			Code:       line,
			Message:    "before/beforeEach allocates test resources without a matching after/afterEach hook",
			Suggestion: "Add after()/afterEach() cleanup for servers, connections, and timers created during setup",
		})
	}
	return findings
}

func isTestIsolationFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return strings.Contains(base, "test") || strings.Contains(base, "spec")
}

func hasAfterHook(lines []string) bool {
	for _, line := range lines {
		if testAfterHookPattern.MatchString(line) {
			return true
		}
	}
	return false
}

func hasCleanupCallInHooks(lines []string, patterns ...*regexp.Regexp) bool {
	for i, line := range lines {
		if !testAfterHookPattern.MatchString(line) && !testCleanupPattern.MatchString(line) {
			continue
		}
		if hookContainsPattern(lines, i, nil, patterns...) {
			return true
		}
	}
	return false
}

func hookContainsPattern(lines []string, start int, startPattern *regexp.Regexp, patterns ...*regexp.Regexp) bool {
	if start < 0 || start >= len(lines) {
		return false
	}

	startLine := lines[start]
	startIndex := 0
	if startPattern != nil {
		if idx := startPattern.FindStringIndex(startLine); idx != nil {
			startIndex = idx[0]
		}
	}

	parenDepth := 0
	braceDepth := 0
	for i := start; i < len(lines); i++ {
		line := lines[i]
		segment := line
		if i == start && startIndex < len(line) {
			segment = line[startIndex:]
		}

		for _, pattern := range patterns {
			if pattern.MatchString(segment) {
				return true
			}
		}

		parenDepth += strings.Count(segment, "(") - strings.Count(segment, ")")
		braceDepth += strings.Count(segment, "{") - strings.Count(segment, "}")

		if parenDepth <= 0 && braceDepth <= 0 {
			break
		}
	}
	return false
}
