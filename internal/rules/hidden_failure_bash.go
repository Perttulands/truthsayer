package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// HiddenFailureBash detects || true, 2>/dev/null, || : patterns in bash scripts.
type HiddenFailureBash struct{}

func (h *HiddenFailureBash) Meta() Rule {
	return Rule{
		ID:          "silent-fallback.hidden-failure-bash",
		Category:    "silent-fallback",
		Name:        "Hidden failure in bash",
		Description: "Command failure silently suppressed with || true, 2>/dev/null, or || : (trap/cleanup-context || true is exempt)",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".sh", ".bash"},
		ScanType:    ScanTypeRegex,
	}
}

var hiddenFailurePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\|\|\s*true`),
	regexp.MustCompile(`\|\|\s*:`),
	regexp.MustCompile(`2>\s*/dev/null`),
}

// Commands where || true is harmless and expected.
var hiddenFailureHarmlessCommands = regexp.MustCompile(
	`^\s*(shopt|mkdir|rmdir)\b`,
)

// simpleOrTrue matches lines where || true follows a simple command (no pipes, &&, etc.).
var simpleOrTrue = regexp.MustCompile(`^\s*\S+(\s+\S+)*\s*\|\|\s*true\s*$`)

var hiddenFailureMessages = []string{
	"'|| true' silently swallows command failure",
	"'|| :' silently swallows command failure",
	"'2>/dev/null' hides error output",
}

var hiddenFailureSuggestions = []string{
	"Handle the error explicitly or log it: cmd || { echo 'cmd failed' >&2; exit 1; }",
	"Handle the error explicitly or log it: cmd || { echo 'cmd failed' >&2; exit 1; }",
	"Capture stderr to a variable or log file instead of discarding it",
}

const cleanupContextAnnotation = "truthsayer:cleanup-context"

var (
	bashFunctionPattern     = regexp.MustCompile(`^\s*(?:function\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*(?:\(\s*\))?\s*\{`)
	trapSingleQuotePattern  = regexp.MustCompile(`\btrap\s+'([^']*)'`)
	trapDoubleQuotePattern  = regexp.MustCompile(`\btrap\s+"([^"]*)"`)
	trapBareCommandPattern  = regexp.MustCompile(`^\s*trap\s+([A-Za-z_][A-Za-z0-9_]*)\b`)
)

type bashFunctionBlock struct {
	name      string
	startLine int
	endLine   int
	annotated bool
}

// hasReasonComment checks if the line has an inline # REASON: comment justifying the suppression,
// or if the immediately preceding line is a # REASON: comment.
func hasReasonComment(line string, lines []string, lineIdx int) bool {
	if strings.Contains(line, "# REASON:") {
		return true
	}
	if lineIdx > 0 && strings.Contains(lines[lineIdx-1], "# REASON:") {
		return true
	}
	return false
}

func (h *HiddenFailureBash) CheckLines(path string, lines []string) []finding.Finding {
	functions := findBashFunctions(lines)
	exemptOrTrueLines := findExemptOrTrueLines(lines, functions)

	base := filepath.Base(path)
	isInstallScript := base == "install.sh"

	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		for j, pat := range hiddenFailurePatterns {
			if pat.MatchString(line) {
				// Senate verdict quick-1771535739:
				// `|| true` inside trap handlers and one-hop cleanup contexts is intentional.
				if j == 0 && exemptOrTrueLines[i+1] {
					continue
				}

				severity := finding.SeverityError
				msg := hiddenFailureMessages[j]
				suggestion := hiddenFailureSuggestions[j]

				if hasReasonComment(line, lines, i) {
					severity = finding.SeverityInfo
					msg = msg + " (justified with REASON comment)"
					suggestion = ""
				} else {
					msg = msg + " — add '# REASON: ...' to justify or fix the suppression"
				}

				// Downgrade harmless commands (shopt, mkdir, rmdir)
				// or simple "cmd || true" at end of line.
				if j == 0 && severity == finding.SeverityError {
					if hiddenFailureHarmlessCommands.MatchString(line) || simpleOrTrue.MatchString(line) {
						severity = finding.SeverityWarning
					}
				}

				// install.sh files get all findings downgraded to Info.
				if isInstallScript {
					severity = finding.SeverityInfo
				}

				findings = append(findings, finding.Finding{
					Rule:       h.Meta().ID,
					Severity:   severity,
					File:       path,
					Line:       i + 1,
					Code:       line,
					Message:    msg,
					Suggestion: suggestion,
				})
			}
		}
	}
	return findings
}

func findExemptOrTrueLines(lines []string, functions []bashFunctionBlock) map[int]bool {
	exempt := make(map[int]bool)
	functionNames := make(map[string]struct{}, len(functions))
	for _, fn := range functions {
		functionNames[fn.name] = struct{}{}
	}

	trapInvoked := findTrapInvokedFunctions(lines, functionNames)

	for idx, line := range lines {
		lineNo := idx + 1
		if isTrapLine(line) && hiddenFailurePatterns[0].MatchString(line) {
			exempt[lineNo] = true
		}
	}

	for _, fn := range functions {
		if !fn.annotated && !trapInvoked[fn.name] {
			continue
		}
		for lineNo := fn.startLine; lineNo <= fn.endLine; lineNo++ {
			exempt[lineNo] = true
		}
	}
	return exempt
}

func findBashFunctions(lines []string) []bashFunctionBlock {
	blocks := make([]bashFunctionBlock, 0)
	current := bashFunctionBlock{}
	depth := 0
	inFunc := false

	for i, line := range lines {
		lineNo := i + 1
		if !inFunc {
			matches := bashFunctionPattern.FindStringSubmatch(line)
			if len(matches) < 2 {
				continue
			}
			current = bashFunctionBlock{
				name:      matches[1],
				startLine: lineNo,
				endLine:   lineNo,
				annotated: hasCleanupContextAnnotation(lines, i),
			}
			depth = braceDelta(line)
			inFunc = true
			if depth <= 0 {
				blocks = append(blocks, current)
				inFunc = false
			}
			continue
		}

		depth += braceDelta(line)
		current.endLine = lineNo
		if depth <= 0 {
			blocks = append(blocks, current)
			inFunc = false
		}
	}

	return blocks
}

func hasCleanupContextAnnotation(lines []string, idx int) bool {
	if strings.Contains(lines[idx], cleanupContextAnnotation) {
		return true
	}
	if idx > 0 && strings.Contains(lines[idx-1], cleanupContextAnnotation) {
		return true
	}
	return false
}

func braceDelta(line string) int {
	return strings.Count(line, "{") - strings.Count(line, "}")
}

func findTrapInvokedFunctions(lines []string, functionNames map[string]struct{}) map[string]bool {
	invoked := make(map[string]bool)
	for _, line := range lines {
		if !isTrapLine(line) {
			continue
		}
		for _, candidate := range trapCandidates(line) {
			if _, ok := functionNames[candidate]; ok {
				invoked[candidate] = true
			}
		}
	}
	return invoked
}

func isTrapLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "trap ")
}

func trapCandidates(line string) []string {
	candidates := make([]string, 0)
	for _, matches := range trapSingleQuotePattern.FindAllStringSubmatch(line, -1) {
		if len(matches) >= 2 {
			candidates = append(candidates, firstCommands(matches[1])...)
		}
	}
	for _, matches := range trapDoubleQuotePattern.FindAllStringSubmatch(line, -1) {
		if len(matches) >= 2 {
			candidates = append(candidates, firstCommands(matches[1])...)
		}
	}
	if matches := trapBareCommandPattern.FindStringSubmatch(line); len(matches) >= 2 {
		candidates = append(candidates, matches[1])
	}
	return candidates
}

func firstCommands(value string) []string {
	segments := strings.FieldsFunc(value, func(r rune) bool {
		return r == ';' || r == '\n'
	})
	out := make([]string, 0, len(segments))
	for _, segment := range segments {
		fields := strings.Fields(strings.TrimSpace(segment))
		if len(fields) == 0 {
			continue
		}
		token := strings.Trim(fields[0], "(){}")
		if token == "" {
			continue
		}
		out = append(out, token)
	}
	return out
}
