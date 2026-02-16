package rules

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// PyGetattrSilentDefault detects getattr(obj, 'attr', None) where the silent
// default may mask a bug rather than handle an optional attribute.
type PyGetattrSilentDefault struct{}

func (p *PyGetattrSilentDefault) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.py-getattr-silent-default",
		Category:    "silent-fallback",
		Name:        "getattr with silent None default",
		Description: "getattr(obj, 'attr', None) silently returns None when the attribute is missing, which may mask a bug",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".py"},
		ScanType:    ScanTypeAST,
	}
}

func (p *PyGetattrSilentDefault) CheckPyAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	var findings []finding.Finding
	for _, node := range pyFindNodesByType(tree.RootNode(), "call") {
		fn := node.ChildByFieldName("function")
		if fn == nil {
			continue
		}
		if pyNodeText(fn, source) != "getattr" {
			continue
		}
		args := node.ChildByFieldName("arguments")
		if args == nil {
			continue
		}
		// getattr with 3 args where 3rd is None
		argList := namedArgs(args)
		if len(argList) != 3 {
			continue
		}
		third := argList[2]
		if third.Type() == "none" || pyNodeText(third, source) == "None" {
			line := pyLineNumber(node)
			findings = append(findings, finding.Finding{
				Rule:       p.Meta().ID,
				Severity:   p.Meta().Severity,
				File:       path,
				Line:       line,
				Code:       pySourceLine(source, line),
				Message:    "getattr() with None default silently masks missing attributes",
				Suggestion: "Use hasattr() check or let AttributeError propagate if the attribute should exist",
			})
		}
	}
	return findings
}

// namedArgs returns the positional argument nodes from an argument_list,
// skipping keyword arguments and **kwargs.
func namedArgs(argList *sitter.Node) []*sitter.Node {
	var args []*sitter.Node
	for i := 0; i < int(argList.NamedChildCount()); i++ {
		child := argList.NamedChild(i)
		switch child.Type() {
		case "keyword_argument", "dictionary_splat":
			continue
		default:
			args = append(args, child)
		}
	}
	return args
}
