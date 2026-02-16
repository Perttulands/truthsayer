package rules

import "testing"

func TestJSOptionalChainSilence_DetectsDeepChain(t *testing.T) {
	src := `const x = a?.b?.c?.d?.e;`
	findings := runJSCheckerOnSource(t, &JSOptionalChainSilence{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "silent-fallback.js-optional-chain-silence" {
		t.Errorf("expected rule silent-fallback.js-optional-chain-silence, got %s", findings[0].Rule)
	}
}

func TestJSOptionalChainSilence_AllowsThreeLevels(t *testing.T) {
	// Exactly 3 levels is OK (threshold is >3)
	src := `const x = a?.b?.c?.d;`
	findings := runJSCheckerOnSource(t, &JSOptionalChainSilence{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for 3-level chain, got %d", len(findings))
	}
}

func TestJSOptionalChainSilence_AllowsTwoLevels(t *testing.T) {
	src := `const x = a?.b?.c;`
	findings := runJSCheckerOnSource(t, &JSOptionalChainSilence{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for 2-level chain, got %d", len(findings))
	}
}

func TestJSOptionalChainSilence_AllowsRegularChain(t *testing.T) {
	// Regular dot access (no ?.) should never trigger
	src := `const x = a.b.c.d.e.f;`
	findings := runJSCheckerOnSource(t, &JSOptionalChainSilence{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for regular chain, got %d", len(findings))
	}
}

func TestJSOptionalChainSilence_MixedChain(t *testing.T) {
	// Mix of regular and optional — count only ?. operators
	// a.b?.c?.d?.e has 3 optional levels — should not trigger
	src := `const x = a.b?.c?.d?.e;`
	findings := runJSCheckerOnSource(t, &JSOptionalChainSilence{}, "app.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for 3-optional mixed chain, got %d", len(findings))
	}
}

func TestJSOptionalChainSilence_MixedChainDeep(t *testing.T) {
	// a.b?.c?.d?.e?.f has 4 optional levels — should trigger
	src := `const x = a.b?.c?.d?.e?.f;`
	findings := runJSCheckerOnSource(t, &JSOptionalChainSilence{}, "app.js", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for 4-optional mixed chain, got %d", len(findings))
	}
}

func TestJSOptionalChainSilence_SkipsTestFiles(t *testing.T) {
	src := `const x = a?.b?.c?.d?.e;`
	findings := runJSCheckerOnSource(t, &JSOptionalChainSilence{}, "app.test.js", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for test file, got %d", len(findings))
	}
}

func TestJSOptionalChainSilence_NoDuplicates(t *testing.T) {
	// A single deep chain should produce exactly one finding, not one per nesting level
	src := `
const a = x?.y?.z?.w?.v;
const b = p?.q?.r?.s?.t;
`
	findings := runJSCheckerOnSource(t, &JSOptionalChainSilence{}, "app.js", src)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings (one per chain), got %d", len(findings))
	}
}
