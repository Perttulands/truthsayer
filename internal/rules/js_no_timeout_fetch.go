package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSNoTimeoutFetch detects fetch() calls without AbortController/signal — can hang indefinitely.
type JSNoTimeoutFetch struct{}

func (j *JSNoTimeoutFetch) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.no-timeout-fetch",
		Category:    "bad-defaults",
		Name:        "No timeout on fetch",
		Description: "fetch() without AbortController/signal can hang indefinitely",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSNoTimeoutFetch) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	for _, node := range jsFindNodesByType(tree.RootNode(), "call_expression") {
		fn := node.ChildByFieldName("function")
		if fn == nil {
			continue
		}
		fnText := jsNodeText(fn, source)
		if fnText != "fetch" {
			continue
		}

		args := node.ChildByFieldName("arguments")
		if args == nil {
			continue
		}

		if hasSignalArg(args, source) {
			continue
		}

		line := jsLineNumber(node)
		findings = append(findings, finding.Finding{
			Rule:       j.Meta().ID,
			Severity:   j.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       jsSourceLine(source, line),
			Message:    "fetch() without AbortController/signal can hang indefinitely",
			Suggestion: "Pass { signal: AbortSignal.timeout(5000) } or use an AbortController",
		})
	}
	return findings
}

// hasSignalArg checks if a fetch call's arguments contain a signal property.
func hasSignalArg(args *sitter.Node, source []byte) bool {
	// Look for an object argument with a "signal" property
	for i := 0; i < int(args.NamedChildCount()); i++ {
		arg := args.NamedChild(i)
		if arg.Type() == "object" {
			if objectHasProperty(arg, source, "signal") {
				return true
			}
		}
		// Also check spread: fetch(url, opts) where opts might have signal
		// We can't statically know, so check for identifier/member that suggests config
		if arg.Type() == "spread_element" {
			return true // conservative: spread might include signal
		}
	}
	return false
}

// objectHasProperty checks if an object literal has a property with the given name.
func objectHasProperty(obj *sitter.Node, source []byte, propName string) bool {
	for i := 0; i < int(obj.NamedChildCount()); i++ {
		child := obj.NamedChild(i)
		if child.Type() == "pair" {
			key := child.ChildByFieldName("key")
			if key != nil && jsNodeText(key, source) == propName {
				return true
			}
		}
		if child.Type() == "shorthand_property_identifier" {
			if jsNodeText(child, source) == propName {
				return true
			}
		}
		// Spread in object: { ...opts } — might include signal
		if child.Type() == "spread_element" {
			return true // conservative: spread might include signal
		}
	}
	return false
}
