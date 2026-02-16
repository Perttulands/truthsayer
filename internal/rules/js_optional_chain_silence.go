package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSOptionalChainSilence detects deep optional chaining (>3 levels of ?.)
// that may mask structural failures as undefined.
type JSOptionalChainSilence struct{}

func (j *JSOptionalChainSilence) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.js-optional-chain-silence",
		Category:    "silent-fallback",
		Name:        "Deep optional chaining",
		Description: "Deep optional chaining (>3 levels) masks structural failures as undefined",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

func (j *JSOptionalChainSilence) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	var findings []finding.Finding
	reported := make(map[uint32]bool) // track by start byte to avoid duplicates

	for _, node := range jsFindNodesByType(tree.RootNode(), "member_expression") {
		// Only process outermost member_expression chains — skip if parent is also member_expression
		if node.Parent() != nil && node.Parent().Type() == "member_expression" {
			continue
		}

		depth := countOptionalChainDepth(node)
		if depth > 3 {
			startByte := node.StartByte()
			if reported[startByte] {
				continue
			}
			reported[startByte] = true

			line := jsLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       j.Meta().ID,
				Severity:   j.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       jsSourceLine(source, line),
				Message:    "Deep optional chaining masks structural failures as undefined",
				Suggestion: "Extract intermediate values and validate them, or use a type guard",
			})
		}
	}
	return findings
}

// countOptionalChainDepth counts the number of ?. operators in a member_expression chain.
func countOptionalChainDepth(node *sitter.Node) int {
	count := 0
	current := node
	for current != nil && current.Type() == "member_expression" {
		if hasOptionalChain(current) {
			count++
		}
		current = current.ChildByFieldName("object")
	}
	return count
}

// hasOptionalChain checks if a member_expression node uses ?. (optional chaining).
func hasOptionalChain(node *sitter.Node) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "optional_chain" {
			return true
		}
	}
	return false
}
