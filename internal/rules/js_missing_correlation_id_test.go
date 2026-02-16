package rules

import (
	"testing"
)

func TestJSMissingCorrelationID(t *testing.T) {
	checker := &JSMissingCorrelationID{}

	t.Run("express app with routes but no correlation ID", func(t *testing.T) {
		src := `
const express = require('express');
const app = express();

app.get('/users', (req, res) => {
  res.json({ users: [] });
});

app.post('/users', (req, res) => {
  res.status(201).json({ created: true });
});

app.listen(3000);
`
		findings := runJSCheckerOnSource(t, checker, "server.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if findings[0].Rule != "trace-gaps.js-missing-correlation-id" {
			t.Errorf("wrong rule: %s", findings[0].Rule)
		}
	})

	t.Run("express app with request-id header", func(t *testing.T) {
		src := `
const express = require('express');
const app = express();

app.use((req, res, next) => {
  req.id = req.headers['x-request-id'] || crypto.randomUUID();
  next();
});

app.get('/users', (req, res) => {
  res.json({ users: [] });
});
`
		findings := runJSCheckerOnSource(t, checker, "server.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("express app with correlation-id package", func(t *testing.T) {
		src := `
const express = require('express');
const correlationId = require('express-correlation-id');
const app = express();

app.use(correlationId());
app.get('/api', handler);
`
		findings := runJSCheckerOnSource(t, checker, "server.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("express app with uuid", func(t *testing.T) {
		src := `
const express = require('express');
const { v4: uuid } = require('uuid');
const app = express();

app.use((req, res, next) => {
  req.requestId = uuid();
  next();
});

app.get('/api', handler);
`
		findings := runJSCheckerOnSource(t, checker, "server.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("non-express file", func(t *testing.T) {
		src := `
const http = require('http');
http.createServer((req, res) => {
  res.end('ok');
}).listen(3000);
`
		findings := runJSCheckerOnSource(t, checker, "server.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for non-express file, got %d", len(findings))
		}
	})

	t.Run("test file skipped", func(t *testing.T) {
		src := `
const express = require('express');
const app = express();
app.get('/test', (req, res) => { res.send('ok'); });
`
		findings := runJSCheckerOnSource(t, checker, "server.test.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for test file, got %d", len(findings))
		}
	})

	t.Run("Koa app with traceId", func(t *testing.T) {
		src := `
const Koa = require('koa');
const app = new Koa();

app.use(async (ctx, next) => {
  ctx.state.traceId = crypto.randomUUID();
  await next();
});

app.get('/api', handler);
`
		findings := runJSCheckerOnSource(t, checker, "server.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings with traceId, got %d", len(findings))
		}
	})
}
