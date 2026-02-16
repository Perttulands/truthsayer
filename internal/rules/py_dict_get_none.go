package rules

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyDictGetNone detects dict.get(key) without an explicit default,
// which returns None and may cause downstream failures.
type PyDictGetNone struct{}

func (p *PyDictGetNone) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.py-dict-get-none",
		Category:    "silent-fallback",
		Name:        "dict.get without explicit default",
		Description: "dict.get(key) without a default returns None, which may cause downstream failures",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyDictGetNone) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "call") {
		fn := node.ChildByFieldName("function")
		if fn == nil || fn.Type() != "attribute" {
			continue
		}
		obj := fn.ChildByFieldName("object")
		attr := fn.ChildByFieldName("attribute")
		if attr == nil || pyNodeText(attr, source) != "get" {
			continue
		}
		// Only flag .get() on plain identifiers (dict variables).
		// Skip module.get() patterns like requests.get(), os.environ.get(), etc.
		if obj == nil || obj.Type() != "identifier" {
			continue
		}
		args := node.ChildByFieldName("arguments")
		if args == nil {
			continue
		}
		positional := namedArgs(args)
		// dict.get(key) with only 1 positional arg — no explicit default
		if len(positional) != 1 {
			continue
		}
		// Skip if the argument looks like a URL (heuristic to avoid
		// false positives from requests.get(), session.get(), etc.)
		argText := pyNodeText(positional[0], source)
		if looksLikeURL(argText) {
			continue
		}
		line := pyLineNumber(node)
		findings = append(findings, finding.Finding{
			Rule:       p.Meta().ID,
			Severity:   p.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       pySourceLine(source, line),
			Message:    "dict.get() without explicit default returns None on missing key",
			Suggestion: "Provide an explicit default value, e.g. data.get('key', '')",
		})
	}
	return findings
}

// looksLikeURL checks if a string argument looks like a URL or path,
// indicating the .get() call is an HTTP request rather than dict access.
func looksLikeURL(text string) bool {
	// Strip quotes from string literals
	t := strings.Trim(text, "\"'")
	return strings.HasPrefix(t, "http://") ||
		strings.HasPrefix(t, "https://") ||
		strings.HasPrefix(t, "/")
}
