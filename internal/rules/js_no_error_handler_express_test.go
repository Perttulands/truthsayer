package rules

import (
	"testing"
)

func TestJSNoErrorHandlerExpress(t *testing.T) {
	checker := &JSNoErrorHandlerExpress{}

	t.Run("express app with routes but no error handler", func(t *testing.T) {
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
		if findings[0].Rule != "trace-gaps.js-no-error-handler-express" {
			t.Errorf("wrong rule: %s", findings[0].Rule)
		}
	})

	t.Run("express app with error middleware", func(t *testing.T) {
		src := `
const express = require('express');
const app = express();

app.get('/users', (req, res) => {
  res.json({ users: [] });
});

app.use((err, req, res, next) => {
  console.error(err.stack);
  res.status(500).send('Something broke!');
});

app.listen(3000);
`
		findings := runJSCheckerOnSource(t, checker, "server.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("express app with error handler using error name", func(t *testing.T) {
		src := `
const express = require('express');
const app = express();

app.get('/', handler);

app.use(function(error, req, res, next) {
  res.status(500).json({ message: error.message });
});
`
		findings := runJSCheckerOnSource(t, checker, "app.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("non-express file", func(t *testing.T) {
		src := `
const http = require('http');
const server = http.createServer((req, res) => {
  res.end('hello');
});
`
		findings := runJSCheckerOnSource(t, checker, "server.js", src)
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings, got %d", len(findings))
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

	t.Run("app.use with only middleware no route", func(t *testing.T) {
		src := `
const express = require('express');
const app = express();

app.use(cors());
app.use(bodyParser.json());

app.get('/api', handler);
`
		findings := runJSCheckerOnSource(t, checker, "server.js", src)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
	})
}
