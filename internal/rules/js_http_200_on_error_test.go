package rules

import (
	"testing"
)

func TestJSHTTP200OnError_Status200InCatch(t *testing.T) {
	src := `
app.get('/api/data', async (req, res) => {
  try {
    const data = await fetchData();
    res.json(data);
  } catch (err) {
    res.status(200).json({ data: null });
  }
});
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "error-context.js-http-200-on-error" {
		t.Errorf("unexpected rule ID: %s", findings[0].Rule)
	}
}

func TestJSHTTP200OnError_ResJsonInCatch(t *testing.T) {
	src := `
app.post('/api/submit', async (req, res) => {
  try {
    await save(req.body);
    res.json({ ok: true });
  } catch (err) {
    res.json({ ok: false });
  }
});
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSHTTP200OnError_Status201InCatch(t *testing.T) {
	src := `
app.post('/api/create', async (req, res) => {
  try {
    const item = await create(req.body);
    res.status(201).json(item);
  } catch (err) {
    res.status(201).json({ created: false });
  }
});
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestJSHTTP200OnError_Negative_ErrorStatus(t *testing.T) {
	src := `
app.get('/api/data', async (req, res) => {
  try {
    const data = await fetchData();
    res.json(data);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for error status, got %d", len(findings))
	}
}

func TestJSHTTP200OnError_Negative_Status400(t *testing.T) {
	src := `
app.post('/api/submit', async (req, res) => {
  try {
    await save(req.body);
    res.json({ ok: true });
  } catch (err) {
    res.status(400).json({ error: 'Bad request' });
  }
});
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for 400 status, got %d", len(findings))
	}
}

func TestJSHTTP200OnError_Negative_NoTryCatch(t *testing.T) {
	src := `
app.get('/api/data', (req, res) => {
  const data = getData();
  res.status(200).json(data);
});
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings outside catch, got %d", len(findings))
	}
}

func TestJSHTTP200OnError_Negative_TestFile(t *testing.T) {
	src := `
try {
  callAPI();
} catch (err) {
  res.status(200).json({ ok: false });
}
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.test.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSHTTP200OnError_MultipleCatches(t *testing.T) {
	src := `
app.get('/a', async (req, res) => {
  try {
    await doA();
  } catch (err) {
    res.status(200).json({ error: null });
  }
});

app.get('/b', async (req, res) => {
  try {
    await doB();
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.get('/c', async (req, res) => {
  try {
    await doC();
  } catch (err) {
    res.json({ ok: false });
  }
});
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.js", src)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestJSHTTP200OnError_CtxResponse(t *testing.T) {
	src := `
router.get('/data', async (ctx) => {
  try {
    const data = await fetch('/api');
  } catch (err) {
    ctx.json({ error: null });
  }
});
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for ctx.json, got %d", len(findings))
	}
}

func TestJSHTTP200OnError_ResponseIdentifier(t *testing.T) {
	src := `
app.get('/data', async (req, response) => {
  try {
    const data = await fetch('/api');
  } catch (err) {
    response.json({ ok: false });
  }
});
`
	checker := &JSHTTP200OnError{}
	findings := runJSCheckerOnSource(t, checker, "routes.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for response.json, got %d", len(findings))
	}
}
