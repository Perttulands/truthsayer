package rules

import (
	"testing"
)

func TestJSEnvTestCheck(t *testing.T) {
	checker := &JSEnvTestCheck{}

	t.Run("process.env.NODE_ENV === test", func(t *testing.T) {
		src := `
if (process.env.NODE_ENV === 'test') {
  db = mockDB;
} else {
  db = realDB;
}
`
		findings := runJSCheckerOnSource(t, checker, "src/db.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "mock-leakage.js-env-test-check" {
			t.Errorf("wrong rule: %s", findings[0].Rule)
		}
	})

	t.Run("reverse comparison", func(t *testing.T) {
		src := `
const isTest = 'test' === process.env.NODE_ENV;
`
		findings := runJSCheckerOnSource(t, checker, "src/config.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("double equals", func(t *testing.T) {
		src := `
if (process.env.NODE_ENV == "test") {
  skipAuth();
}
`
		findings := runJSCheckerOnSource(t, checker, "src/auth.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("not-equals check", func(t *testing.T) {
		src := `
if (process.env.NODE_ENV !== 'test') {
  enableTelemetry();
}
`
		findings := runJSCheckerOnSource(t, checker, "src/telemetry.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})

	t.Run("NODE_ENV production check is fine", func(t *testing.T) {
		src := `
if (process.env.NODE_ENV === 'production') {
  enableCaching();
}
`
		findings := runJSCheckerOnSource(t, checker, "src/config.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for production check, got %d", len(findings))
		}
	})

	t.Run("NODE_ENV development check is fine", func(t *testing.T) {
		src := `
if (process.env.NODE_ENV === 'development') {
  enableDevTools();
}
`
		findings := runJSCheckerOnSource(t, checker, "src/config.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for development check, got %d", len(findings))
		}
	})

	t.Run("other env var is fine", func(t *testing.T) {
		src := `
if (process.env.DEBUG === 'true') {
  enableLogging();
}
`
		findings := runJSCheckerOnSource(t, checker, "src/logger.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for other env var, got %d", len(findings))
		}
	})

	t.Run("test file skipped", func(t *testing.T) {
		src := `
if (process.env.NODE_ENV === 'test') {
  setupMocks();
}
`
		findings := runJSCheckerOnSource(t, checker, "setup.test.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for test file, got %d", len(findings))
		}
	})
}
