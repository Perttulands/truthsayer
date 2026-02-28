package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// UnusedVariable detects likely unused variable declarations by file-local reference counting.
type UnusedVariable struct{}

func (u *UnusedVariable) Meta() Rule {
	return Rule{
		ID:          "code-quality.unused-variable",
		Category:    "code-quality",
		Name:        "Likely unused variable",
		Description: "Variable is declared but appears only at its declaration site",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go", ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".py", ".pyi"},
		ScanType:    ScanTypeRegex,
	}
}

var (
	jsVarDeclPattern  = regexp.MustCompile(`^\s*(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\b`)
	goVarPattern      = regexp.MustCompile(`^\s*var\s+([A-Za-z_][A-Za-z0-9_]*)\b`)
	goShortVarPattern = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*(?:\s*,\s*[A-Za-z_][A-Za-z0-9_]*)*)\s*:=`)
	pyVarPattern      = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*(?:\s*,\s*[A-Za-z_][A-Za-z0-9_]*)*)\s*=`)
)

type declaration struct {
	line int
	code string
}

func (u *UnusedVariable) CheckLines(path string, lines []string) []finding.Finding {
	ext := filepath.Ext(path)
	decls := collectDeclarations(ext, lines)
	if len(decls) == 0 {
		return nil
	}

	counts := make(map[string]int, len(decls))
	for name := range decls {
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
		for _, line := range lines {
			code := stripInlineComment(line)
			counts[name] += len(re.FindAllStringIndex(code, -1))
		}
	}

	var findings []finding.Finding
	for name, decl := range decls {
		if counts[name] > 1 {
			continue
		}
		findings = append(findings, finding.Finding{
			Rule:       u.Meta().ID,
			Severity:   u.Meta().Severity,
			File:       path,
			Line:       decl.line,
			Code:       decl.code,
			Message:    "Variable " + name + " appears to be unused",
			Suggestion: "Remove the variable or use it; if intentionally unused, rename it to _",
		})
	}
	return findings
}

func collectDeclarations(ext string, lines []string) map[string]declaration {
	decls := make(map[string]declaration)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}

		switch ext {
		case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
			m := jsVarDeclPattern.FindStringSubmatch(line)
			if len(m) == 2 {
				name := m[1]
				if isCandidateVar(name) {
					decls[name] = declaration{line: i + 1, code: line}
				}
			}
		case ".go":
			if m := goVarPattern.FindStringSubmatch(line); len(m) == 2 {
				name := m[1]
				// Skip package-level var declarations — they're visible across
				// all files in the package and can't be checked by single-file analysis.
				if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
					continue
				}
				if isCandidateVar(name) {
					decls[name] = declaration{line: i + 1, code: line}
				}
				continue
			}
			if m := goShortVarPattern.FindStringSubmatch(line); len(m) == 2 {
				for _, name := range splitNames(m[1]) {
					if isCandidateVar(name) {
						decls[name] = declaration{line: i + 1, code: line}
					}
				}
			}
		case ".py", ".pyi":
			if m := pyVarPattern.FindStringSubmatch(line); len(m) == 2 {
				for _, name := range splitNames(m[1]) {
					if isCandidateVar(name) {
						decls[name] = declaration{line: i + 1, code: line}
					}
				}
			}
		}
	}
	return decls
}

func splitNames(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func isCandidateVar(name string) bool {
	if name == "_" {
		return false
	}
	if strings.HasPrefix(name, "_") {
		return false
	}
	return true
}

// UnreachableCode detects likely unreachable statements immediately after terminal statements.
type UnreachableCode struct{}

func (u *UnreachableCode) Meta() Rule {
	return Rule{
		ID:          "code-quality.unreachable-code",
		Category:    "code-quality",
		Name:        "Likely unreachable code",
		Description: "Statement appears after return/throw/raise/break/continue in the same block",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".go", ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".py", ".pyi"},
		ScanType:    ScanTypeRegex,
	}
}

var (
	braceTerminatorPattern = regexp.MustCompile(`^(?:return|throw|continue|break)\b|^panic\s*\(`)
	pyTerminatorPattern    = regexp.MustCompile(`^(?:return|raise|continue|break)\b`)
)

func (u *UnreachableCode) CheckLines(path string, lines []string) []finding.Finding {
	ext := filepath.Ext(path)
	if ext == ".py" || ext == ".pyi" {
		return u.checkPython(lines, path)
	}
	return u.checkBraceLang(lines, path)
}

func (u *UnreachableCode) checkPython(lines []string, path string) []finding.Finding {
	terminatedIndent := -1
	var findings []finding.Finding

	for i, line := range lines {
		trimmed := strings.TrimSpace(stripInlineComment(line))
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}

		indent := leadingIndent(line)
		if terminatedIndent >= 0 {
			if indent < terminatedIndent {
				terminatedIndent = -1
			} else if indent == terminatedIndent {
				if !strings.HasPrefix(trimmed, "elif ") &&
					!strings.HasPrefix(trimmed, "else:") &&
					!strings.HasPrefix(trimmed, "except") &&
					!strings.HasPrefix(trimmed, "finally") {
					findings = append(findings, finding.Finding{
						Rule:       u.Meta().ID,
						Severity:   u.Meta().Severity,
						File:       path,
						Line:       i + 1,
						Code:       line,
						Message:    "Statement appears unreachable after control flow already returned/raised",
						Suggestion: "Remove unreachable code or move it before the return/raise statement",
					})
				}
			}
		}

		if pyTerminatorPattern.MatchString(trimmed) {
			terminatedIndent = indent
		}
	}

	return findings
}

func (u *UnreachableCode) checkBraceLang(lines []string, path string) []finding.Finding {
	depth := 0
	terminatedDepth := -1
	var findings []finding.Finding

	// For Go files, pre-compute which lines are inside raw string literals
	// (backtick-delimited). Go raw strings cannot contain backticks, so
	// toggling on odd backtick count per line is reliable.
	var inRawString []bool
	if filepath.Ext(path) == ".go" {
		inRawString = goRawStringLines(lines)
	}

	for i, line := range lines {
		// Skip lines inside Go raw string literals
		if inRawString != nil && inRawString[i] {
			continue
		}

		code := strings.TrimSpace(stripInlineComment(line))
		depthBefore := depth

		if code != "" && !isCommentLine(code) {
			if terminatedDepth >= 0 {
				if depthBefore < terminatedDepth {
					terminatedDepth = -1
				} else if strings.HasPrefix(code, "}") ||
					strings.HasPrefix(code, "case ") ||
					strings.HasPrefix(code, "default:") ||
					strings.HasPrefix(code, "else") {
					// New branch resets termination — code after case/default/else is reachable
					terminatedDepth = -1
				} else if depthBefore == terminatedDepth {
					findings = append(findings, finding.Finding{
						Rule:       u.Meta().ID,
						Severity:   u.Meta().Severity,
						File:       path,
						Line:       i + 1,
						Code:       line,
						Message:    "Statement appears unreachable after control flow already exited this block",
						Suggestion: "Remove unreachable code or move it before return/throw/break/continue",
					})
				}
			}

			if braceTerminatorPattern.MatchString(code) &&
				!isContinuationLine(code) &&
				!isContinuationLine(strings.TrimSpace(line)) {
				terminatedDepth = depthBefore
			}
		}

		depth += strings.Count(line, "{")
		depth -= strings.Count(line, "}")
		if depth < 0 {
			depth = 0
		}
	}

	return findings
}

func leadingIndent(line string) int {
	count := 0
	for _, ch := range line {
		switch ch {
		case ' ':
			count++
		case '\t':
			count += 4
		default:
			return count
		}
	}
	return count
}

// goRawStringLines returns a bool slice where true means the line is inside
// a Go raw string literal (backtick-delimited). The line that opens the
// backtick string is NOT marked (it contains real code like `return`).
// Go raw strings cannot contain backticks, so toggling on odd count is exact.
func goRawStringLines(lines []string) []bool {
	result := make([]bool, len(lines))
	inRaw := false
	for i, line := range lines {
		result[i] = inRaw
		if strings.Count(line, "`")%2 == 1 {
			inRaw = !inRaw
		}
	}
	return result
}

// isContinuationLine returns true if a trimmed code line ends with an operator
// that means the expression continues on the next line. A "return x +" or
// "return a ||" is not a complete terminator.
func isContinuationLine(code string) bool {
	for _, suffix := range []string{"+", "-", "||", "&&", ",", "|", "(", ".", "{"} {
		if strings.HasSuffix(code, suffix) {
			return true
		}
	}
	return false
}

// ErrorSwallowing detects empty catch/except blocks that swallow errors.
type ErrorSwallowing struct{}

func (e *ErrorSwallowing) Meta() Rule {
	return Rule{
		ID:          "code-quality.error-swallowing",
		Category:    "code-quality",
		Name:        "Error swallowing",
		Description: "Empty catch/except blocks hide failures and make incidents harder to debug",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".py", ".pyi"},
		ScanType:    ScanTypeRegex,
	}
}

var (
	jsEmptyCatchInlinePattern  = regexp.MustCompile(`\bcatch\s*(?:\([^)]*\))?\s*\{\s*\}`)
	jsCatchBlockStartPattern   = regexp.MustCompile(`\bcatch\s*(?:\([^)]*\))?\s*\{\s*$`)
	pyEmptyExceptInlinePattern = regexp.MustCompile(`^except\b.*:\s*(?:pass|\.\.\.)\s*$`)
	pyExceptHeaderPattern      = regexp.MustCompile(`^except\b.*:\s*$`)
)

func (e *ErrorSwallowing) CheckLines(path string, lines []string) []finding.Finding {
	ext := filepath.Ext(path)
	if ext == ".py" || ext == ".pyi" {
		return e.checkPython(path, lines)
	}
	if isJSLikeExt(ext) {
		return e.checkJS(path, lines)
	}
	return nil
}

func isJSLikeExt(ext string) bool {
	switch ext {
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}

func (e *ErrorSwallowing) checkJS(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(stripInlineComment(line))
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}

		if jsEmptyCatchInlinePattern.MatchString(trimmed) {
			findings = append(findings, finding.Finding{
				Rule:       e.Meta().ID,
				Severity:   e.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Empty catch block swallows the original error",
				Suggestion: "Log, wrap, or rethrow the error with context",
			})
			continue
		}

		if jsCatchBlockStartPattern.MatchString(trimmed) {
			nextLine, ok := nextSignificantLine(lines, i+1)
			if !ok {
				continue
			}
			if strings.TrimSpace(stripInlineComment(nextLine)) == "}" {
				findings = append(findings, finding.Finding{
					Rule:       e.Meta().ID,
					Severity:   e.Meta().Severity,
					File:       path,
					Line:       i + 1,
					Code:       line,
					Message:    "Empty catch block swallows the original error",
					Suggestion: "Log, wrap, or rethrow the error with context",
				})
			}
		}
	}
	return findings
}

func (e *ErrorSwallowing) checkPython(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(stripInlineComment(line))
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}

		if pyEmptyExceptInlinePattern.MatchString(trimmed) {
			findings = append(findings, finding.Finding{
				Rule:       e.Meta().ID,
				Severity:   e.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Empty except block swallows the original exception",
				Suggestion: "Log the exception or re-raise it",
			})
			continue
		}

		if pyExceptHeaderPattern.MatchString(trimmed) {
			nextLine, ok := nextSignificantLine(lines, i+1)
			if !ok {
				continue
			}
			nextTrimmed := strings.TrimSpace(stripInlineComment(nextLine))
			if nextTrimmed == "pass" || nextTrimmed == "..." {
				findings = append(findings, finding.Finding{
					Rule:       e.Meta().ID,
					Severity:   e.Meta().Severity,
					File:       path,
					Line:       i + 1,
					Code:       line,
					Message:    "Empty except block swallows the original exception",
					Suggestion: "Log the exception or re-raise it",
				})
			}
		}
	}
	return findings
}

func nextSignificantLine(lines []string, start int) (string, bool) {
	for i := start; i < len(lines); i++ {
		trimmed := strings.TrimSpace(stripInlineComment(lines[i]))
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}
		return lines[i], true
	}
	return "", false
}

// HardcodedCredentials detects likely hardcoded secrets across source files.
type HardcodedCredentials struct{}

func (h *HardcodedCredentials) Meta() Rule {
	return Rule{
		ID:          "config-smells.hardcoded-credentials",
		Category:    "config-smells",
		Name:        "Hardcoded credentials",
		Description: "Inline password/token/api key literals should be moved to environment or a secret manager",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".go", ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".py", ".pyi", ".sh", ".bash"},
		ScanType:    ScanTypeRegex,
	}
}

var hardcodedCredentialAnyLangPattern = regexp.MustCompile(`(?i)\b(?:password|passwd|api[_-]?key|secret|secret[_-]?key|token|auth[_-]?token)\b\s*[:=]\s*["'` + "`" + `][^"'` + "`" + `]{4,}["'` + "`" + `]`)

func (h *HardcodedCredentials) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(stripInlineComment(line))
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}
		if hardcodedCredentialAnyLangPattern.MatchString(trimmed) {
			findings = append(findings, finding.Finding{
				Rule:       h.Meta().ID,
				Severity:   h.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Hardcoded credential-like value found in source code",
				Suggestion: "Load the value from environment variables or a secrets manager",
			})
		}
	}
	return findings
}

// SQLInjection detects likely SQL string construction with runtime interpolation.
type SQLInjection struct{}

func (s *SQLInjection) Meta() Rule {
	return Rule{
		ID:          "security.sql-injection",
		Category:    "security",
		Name:        "Potential SQL injection",
		Description: "SQL query appears to be built with string concatenation/interpolation",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".go", ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".py", ".pyi"},
		ScanType:    ScanTypeRegex,
	}
}

var (
	sqlConcatPattern        = regexp.MustCompile(`(?i)\b(?:select|insert|update|delete)\b.*(?:\+|\$\{)`)
	sqlFmtSprintfPattern    = regexp.MustCompile(`(?i)fmt\.sprintf\s*\(\s*["'` + "`" + `]\s*(?:select|insert|update|delete)\b`)
	sqlPythonFStringPattern = regexp.MustCompile(`(?i)f["'][^"']*\b(?:select|insert|update|delete)\b[^"']*\{[^}]+\}`)
)

func (s *SQLInjection) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(stripInlineComment(line))
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}
		if sqlConcatPattern.MatchString(trimmed) ||
			sqlFmtSprintfPattern.MatchString(trimmed) ||
			sqlPythonFStringPattern.MatchString(trimmed) {
			findings = append(findings, finding.Finding{
				Rule:       s.Meta().ID,
				Severity:   s.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "SQL query appears to be dynamically interpolated",
				Suggestion: "Use parameterized queries/prepared statements instead of concatenating user input",
			})
		}
	}
	return findings
}

// CommandInjection detects shell command construction with unsanitized interpolation.
type CommandInjection struct{}

func (c *CommandInjection) Meta() Rule {
	return Rule{
		ID:          "security.command-injection",
		Category:    "security",
		Name:        "Potential command injection",
		Description: "Shell command execution appears to include runtime string interpolation",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".go", ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".py", ".pyi", ".sh", ".bash"},
		ScanType:    ScanTypeRegex,
	}
}

var (
	commandConcatPattern = regexp.MustCompile(`(?i)\b(?:os\.system|exec\.command|child_process\.exec|exec\(|spawn\(|subprocess\.(?:run|call|popen))\b.*(?:\+|\$\{|fmt\.sprintf|f["'])`)
	shellTruePattern     = regexp.MustCompile(`(?i)\bsubprocess\.(?:run|call|popen)\b.*\bshell\s*=\s*true\b`)
	execShDashCPattern   = regexp.MustCompile(`(?i)\bexec\.command\s*\(\s*["']sh["']\s*,\s*["']-c["']`)
)

func (c *CommandInjection) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(stripInlineComment(line))
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}
		if commandConcatPattern.MatchString(trimmed) ||
			shellTruePattern.MatchString(trimmed) ||
			execShDashCPattern.MatchString(trimmed) {
			findings = append(findings, finding.Finding{
				Rule:       c.Meta().ID,
				Severity:   c.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Command execution appears to include interpolated input",
				Suggestion: "Use argument arrays and strict allowlists instead of shell concatenation",
			})
		}
	}
	return findings
}

func stripInlineComment(line string) string {
	for _, marker := range []string{"//", "#"} {
		if idx := strings.Index(line, marker); idx >= 0 {
			line = line[:idx]
		}
	}
	return strings.TrimSpace(line)
}

func isCommentLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "//") ||
		strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, "/*") ||
		strings.HasPrefix(trimmed, "*")
}
