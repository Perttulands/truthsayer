package rules

import "testing"

func TestTestLeakedServer_FindsListenWithoutCleanup(t *testing.T) {
	checker := &TestLeakedServer{}
	lines := []string{
		`describe("api", () => {`,
		`  const server = app.listen(0)`,
		`})`,
	}

	findings := checker.CheckLines("api.test.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "test-isolation.test-leaked-server" {
		t.Fatalf("expected rule test-isolation.test-leaked-server, got %s", findings[0].Rule)
	}
}

func TestTestLeakedServer_SkipsWhenAfterEachClosesServer(t *testing.T) {
	checker := &TestLeakedServer{}
	lines := []string{
		`describe("api", () => {`,
		`  let server`,
		`  beforeEach(() => {`,
		`    server = app.listen(0)`,
		`  })`,
		`  afterEach(() => {`,
		`    server.close()`,
		`  })`,
		`})`,
	}

	findings := checker.CheckLines("api.test.js", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestTestLeakedServer_SkipsNonTestFiles(t *testing.T) {
	checker := &TestLeakedServer{}
	lines := []string{
		`const server = app.listen(0)`,
	}

	findings := checker.CheckLines("server.js", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestTestLeakedSSE_FindsEventSourceWithoutCleanup(t *testing.T) {
	checker := &TestLeakedSSE{}
	lines := []string{
		`describe("stream", () => {`,
		`  const source = new EventSource("/events")`,
		`})`,
	}

	findings := checker.CheckLines("stream.spec.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "test-isolation.test-leaked-sse" {
		t.Fatalf("expected rule test-isolation.test-leaked-sse, got %s", findings[0].Rule)
	}
}

func TestTestLeakedSSE_SkipsWhenCleanupAborts(t *testing.T) {
	checker := &TestLeakedSSE{}
	lines := []string{
		`describe("stream", () => {`,
		`  let controller`,
		`  const source = new EventSource("/events")`,
		`  after(() => {`,
		`    controller.abort()`,
		`  })`,
		`})`,
	}

	findings := checker.CheckLines("stream.spec.js", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestTestMissingCleanup_FindsBeforeEachResourceSetupWithoutAfterHook(t *testing.T) {
	checker := &TestMissingCleanup{}
	lines := []string{
		`beforeEach(() => {`,
		`  server = app.listen(0)`,
		`  ticker = setInterval(() => {}, 1000)`,
		`})`,
		`it("works", () => {})`,
	}

	findings := checker.CheckLines("routes.test.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "test-isolation.test-missing-cleanup" {
		t.Fatalf("expected rule test-isolation.test-missing-cleanup, got %s", findings[0].Rule)
	}
}

func TestTestMissingCleanup_NoFindingWhenAfterEachExists(t *testing.T) {
	checker := &TestMissingCleanup{}
	lines := []string{
		`beforeEach(() => {`,
		`  server = app.listen(0)`,
		`})`,
		`afterEach(() => {`,
		`  server.close()`,
		`})`,
	}

	findings := checker.CheckLines("routes.test.js", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestTestIsolationRuleMetadata(t *testing.T) {
	server := (&TestLeakedServer{}).Meta()
	if server.Category != "test-isolation" {
		t.Fatalf("expected category test-isolation, got %s", server.Category)
	}
	if server.Severity != "warning" {
		t.Fatalf("expected warning severity, got %s", server.Severity)
	}

	sse := (&TestLeakedSSE{}).Meta()
	if sse.Category != "test-isolation" {
		t.Fatalf("expected category test-isolation, got %s", sse.Category)
	}
	if sse.Severity != "warning" {
		t.Fatalf("expected warning severity, got %s", sse.Severity)
	}

	missing := (&TestMissingCleanup{}).Meta()
	if missing.Category != "test-isolation" {
		t.Fatalf("expected category test-isolation, got %s", missing.Category)
	}
	if missing.Severity != "info" {
		t.Fatalf("expected info severity, got %s", missing.Severity)
	}
}
