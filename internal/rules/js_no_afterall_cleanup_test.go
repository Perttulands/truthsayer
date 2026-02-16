package rules

import (
	"testing"
)

func TestJSNoAfterallCleanup(t *testing.T) {
	checker := &JSNoAfterallCleanup{}

	t.Run("beforeAll without afterAll", func(t *testing.T) {
		src := `
describe('Database', () => {
  beforeAll(() => {
    db = connect();
  });

  it('should query', () => {
    expect(db.query('SELECT 1')).toBeDefined();
  });
});
`
		findings := runJSCheckerOnSource(t, checker, "src/db.test.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "test-isolation.no-afterall-cleanup" {
			t.Errorf("expected rule test-isolation.no-afterall-cleanup, got %s", findings[0].Rule)
		}
		if findings[0].Message != "beforeAll without matching afterAll — resources may leak between tests" {
			t.Errorf("unexpected message: %s", findings[0].Message)
		}
	})

	t.Run("beforeEach without afterEach", func(t *testing.T) {
		src := `
describe('Server', () => {
  beforeEach(() => {
    server = startServer();
  });

  it('should respond', () => {
    expect(server.status).toBe(200);
  });
});
`
		findings := runJSCheckerOnSource(t, checker, "src/server.spec.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Message != "beforeEach without matching afterEach — resources may leak between tests" {
			t.Errorf("unexpected message: %s", findings[0].Message)
		}
	})

	t.Run("both beforeAll and beforeEach without cleanup", func(t *testing.T) {
		src := `
describe('API', () => {
  beforeAll(() => {
    db = connect();
  });

  beforeEach(() => {
    req = createRequest();
  });

  it('should work', () => {});
});
`
		findings := runJSCheckerOnSource(t, checker, "src/api.test.js", src)
		if len(findings) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(findings))
		}
	})

	t.Run("beforeAll with afterAll - clean", func(t *testing.T) {
		src := `
describe('Database', () => {
  beforeAll(() => {
    db = connect();
  });

  afterAll(() => {
    db.close();
  });

  it('should query', () => {});
});
`
		findings := runJSCheckerOnSource(t, checker, "src/db.test.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("beforeEach with afterEach - clean", func(t *testing.T) {
		src := `
describe('Server', () => {
  beforeEach(() => {
    server = startServer();
  });

  afterEach(() => {
    server.close();
  });

  it('should respond', () => {});
});
`
		findings := runJSCheckerOnSource(t, checker, "src/server.spec.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("nested describe blocks checked independently", func(t *testing.T) {
		src := `
describe('outer', () => {
  beforeAll(() => { db = connect(); });
  afterAll(() => { db.close(); });

  describe('inner', () => {
    beforeEach(() => { server = start(); });
    it('should work', () => {});
  });
});
`
		findings := runJSCheckerOnSource(t, checker, "src/nested.test.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding (inner beforeEach without afterEach), got %d", len(findings))
		}
	})

	t.Run("not a test file - skip", func(t *testing.T) {
		src := `
beforeAll(() => { setup(); });
`
		findings := runJSCheckerOnSource(t, checker, "src/setup.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for non-test file, got %d", len(findings))
		}
	})

	t.Run("top-level beforeAll without afterAll in test file", func(t *testing.T) {
		src := `
beforeAll(() => {
  db = connect();
});

test('should work', () => {});
`
		findings := runJSCheckerOnSource(t, checker, "src/db.test.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})
}
