package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PySubprocessNoCheck detects subprocess.run()/subprocess.call() without check=True.
type PySubprocessNoCheck struct{}

func (p *PySubprocessNoCheck) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.py-subprocess-no-check",
		Category:    "silent-fallback",
		Name:        "subprocess without check=True",
		Description: "subprocess.run() or subprocess.call() without check=True silently ignores non-zero exit codes",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

// subprocessFuncs are the subprocess functions that should use check=True.
var subprocessFuncs = map[string]bool{
	"run":  true,
	"call": true,
}

func (p *PySubprocessNoCheck) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "call") {
		fn := node.ChildByFieldName("function")
		if fn == nil || fn.Type() != "attribute" {
			continue
		}
		obj := fn.ChildByFieldName("object")
		attr := fn.ChildByFieldName("attribute")
		if obj == nil || attr == nil {
			continue
		}
		if pyNodeText(obj, source) != "subprocess" || !subprocessFuncs[pyNodeText(attr, source)] {
			continue
		}
		if hasCheckTrue(node, source) {
			continue
		}
		line := pyLineNumber(node)
		findings = append(findings, finding.Finding{
			Rule:       p.Meta().ID,
			Severity:   p.Meta().Severity,
			File:       path,
			Line:       line,
			Code:       pySourceLine(source, line),
			Message:    "subprocess." + pyNodeText(attr, source) + "() without check=True — non-zero exit code silently ignored",
			Suggestion: "Add check=True to raise CalledProcessError on failure",
		})
	}
	return findings
}

// hasCheckTrue checks if a call node has a keyword argument check=True.
func hasCheckTrue(callNode *sitter.Node, source []byte) bool {
	args := callNode.ChildByFieldName("arguments")
	if args == nil {
		return false
	}
	for i := 0; i < int(args.NamedChildCount()); i++ {
		child := args.NamedChild(i)
		if child.Type() != "keyword_argument" {
			continue
		}
		name := child.ChildByFieldName("name")
		value := child.ChildByFieldName("value")
		if name != nil && value != nil &&
			pyNodeText(name, source) == "check" &&
			pyNodeText(value, source) == "True" {
			return true
		}
	}
	return false
}
