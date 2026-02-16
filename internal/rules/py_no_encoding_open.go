package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyNoEncodingOpen detects open() calls without explicit encoding= parameter.
type PyNoEncodingOpen struct{}

func (p *PyNoEncodingOpen) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.py-no-encoding-open",
		Category:    "bad-defaults",
		Name:        "open() without encoding",
		Description: "open() without explicit encoding= uses platform-dependent default encoding",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyNoEncodingOpen) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "call") {
		fn := node.ChildByFieldName("function")
		if fn == nil {
			continue
		}
		// Match bare `open(...)` calls (not `io.open` or similar qualified calls)
		if fn.Type() != "identifier" || pyNodeText(fn, source) != "open" {
			continue
		}
		// Skip binary modes — encoding is irrelevant for binary
		if isBinaryMode(node, source) {
			continue
		}
		if hasKeywordArg(node, source, "encoding") {
			continue
		}
		line := pyLineNumber(node)
		findings = append(findings, finding.Finding{
			Rule:       p.Meta().ID,
			Severity:   p.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       pySourceLine(source, line),
			Message:    "open() without encoding= uses platform-dependent default encoding",
			Suggestion: "Add encoding='utf-8' or the appropriate encoding for your data",
		})
	}
	return findings
}

// isBinaryMode checks if an open() call uses a binary mode like "rb", "wb".
func isBinaryMode(callNode *sitter.Node, source []byte) bool {
	args := callNode.ChildByFieldName("arguments")
	if args == nil {
		return false
	}
	// Check positional mode argument (2nd positional arg)
	positionalIdx := 0
	for i := 0; i < int(args.NamedChildCount()); i++ {
		child := args.NamedChild(i)
		if child.Type() == "keyword_argument" {
			name := child.ChildByFieldName("name")
			if name != nil && pyNodeText(name, source) == "mode" {
				value := child.ChildByFieldName("value")
				return value != nil && isBinaryModeStr(pyNodeText(value, source))
			}
			continue
		}
		if positionalIdx == 1 {
			// Second positional arg is the mode
			return isBinaryModeStr(pyNodeText(child, source))
		}
		positionalIdx++
	}
	return false
}

// isBinaryModeStr checks if a mode string contains 'b' indicating binary mode.
func isBinaryModeStr(s string) bool {
	// Strip quotes from string literal
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') {
		s = s[1 : len(s)-1]
	}
	for _, c := range s {
		if c == 'b' {
			return true
		}
	}
	return false
}
