package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// rsIsTestFile returns true if the file is in a tests/test directory or ends in _test.rs.
func rsIsTestFile(path string) bool {
	if strings.HasSuffix(path, "_test.rs") {
		return true
	}
	// Normalize to forward slashes for consistent matching.
	norm := filepath.ToSlash(path)
	// Check both embedded (/tests/) and leading (tests/) positions.
	if strings.Contains(norm, "/tests/") || strings.HasPrefix(norm, "tests/") {
		return true
	}
	if strings.Contains(norm, "/test/") || strings.HasPrefix(norm, "test/") {
		return true
	}
	return false
}

// rsHasCfgTest returns true if any line contains #[cfg(test)].
func rsHasCfgTest(lines []string) bool {
	for _, line := range lines {
		if strings.Contains(line, "#[cfg(test)]") {
			return true
		}
	}
	return false
}

// --- ERROR severity rules ---

// RustUnwrapInProduction detects .unwrap() in non-test .rs files.
type RustUnwrapInProduction struct{}

func (r *RustUnwrapInProduction) Meta() Rule {
	return Rule{
		ID:          "rust.unwrap-in-production",
		Category:    "rust",
		Name:        "unwrap() in production code",
		Description: ".unwrap() can panic at runtime — use .expect() with context or propagate with ?",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsUnwrapPattern = regexp.MustCompile(`\.unwrap\(\)`)

func (r *RustUnwrapInProduction) CheckLines(path string, lines []string) []finding.Finding {
	if rsIsTestFile(path) || rsHasCfgTest(lines) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsUnwrapPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       r.Meta().ID,
				Severity:   r.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    ".unwrap() can panic in production — use .expect(\"context\") or propagate with ?",
				Suggestion: "Replace .unwrap() with .expect(\"descriptive message\") or use the ? operator",
			})
		}
	}
	return findings
}

// RustIgnoredResult detects `let _ =` patterns that discard results.
type RustIgnoredResult struct{}

func (r *RustIgnoredResult) Meta() Rule {
	return Rule{
		ID:          "rust.ignored-result",
		Category:    "rust",
		Name:        "Ignored Result or Error",
		Description: "let _ = discards a Result or Error value without handling it",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsIgnoredResultPattern = regexp.MustCompile(`^[[:space:]]*let _ =`)

func (r *RustIgnoredResult) CheckLines(path string, lines []string) []finding.Finding {
	if rsIsTestFile(path) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsIgnoredResultPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       r.Meta().ID,
				Severity:   r.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "let _ = discards a potentially important Result or Error",
				Suggestion: "Handle the result explicitly or use _ = with a comment explaining why it's safe to ignore",
			})
		}
	}
	return findings
}

// RustPanicInLib detects panic!() in non-test, non-main .rs files.
type RustPanicInLib struct{}

func (r *RustPanicInLib) Meta() Rule {
	return Rule{
		ID:          "rust.panic-in-lib",
		Category:    "rust",
		Name:        "panic!() in library code",
		Description: "panic!() in library code crashes the caller — return Result instead",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsPanicPattern = regexp.MustCompile(`panic!\(`)

func (r *RustPanicInLib) CheckLines(path string, lines []string) []finding.Finding {
	if rsIsTestFile(path) {
		return nil
	}
	base := filepath.Base(path)
	if base == "main.rs" {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsPanicPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       r.Meta().ID,
				Severity:   r.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "panic!() in library code will crash the calling application",
				Suggestion: "Return a Result<T, E> and let the caller decide how to handle the error",
			})
		}
	}
	return findings
}

// RustUnsafeNoComment detects unsafe blocks without a preceding // SAFETY: comment.
type RustUnsafeNoComment struct{}

func (r *RustUnsafeNoComment) Meta() Rule {
	return Rule{
		ID:          "rust.unsafe-no-comment",
		Category:    "rust",
		Name:        "unsafe block without SAFETY comment",
		Description: "unsafe blocks should have a // SAFETY: comment on the preceding line explaining why it's safe",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsUnsafeBlockPattern = regexp.MustCompile(`unsafe\s*\{`)

func (r *RustUnsafeNoComment) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsUnsafeBlockPattern.MatchString(line) {
			hasSafety := false
			if i > 0 {
				prevTrimmed := strings.TrimSpace(lines[i-1])
				if strings.Contains(prevTrimmed, "// SAFETY:") {
					hasSafety = true
				}
			}
			if !hasSafety {
				findings = append(findings, finding.Finding{
					Rule:       r.Meta().ID,
					Severity:   r.Meta().Severity,
					File:       path,
					Line:       i + 1,
					Code:       line,
					Message:    "unsafe block without // SAFETY: comment on the preceding line",
					Suggestion: "Add a '// SAFETY: <explanation>' comment on the line above the unsafe block",
				})
			}
		}
	}
	return findings
}

// --- WARNING severity rules ---

// RustTodoInProduction detects todo!() and unimplemented!() in non-test code.
type RustTodoInProduction struct{}

func (r *RustTodoInProduction) Meta() Rule {
	return Rule{
		ID:          "rust.todo-in-production",
		Category:    "rust",
		Name:        "todo!()/unimplemented!() in production code",
		Description: "todo!() and unimplemented!() will panic at runtime if reached",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsTodoPattern = regexp.MustCompile(`todo!\(|unimplemented!\(`)

func (r *RustTodoInProduction) CheckLines(path string, lines []string) []finding.Finding {
	if rsIsTestFile(path) {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsTodoPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       r.Meta().ID,
				Severity:   r.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "todo!()/unimplemented!() will panic if reached in production",
				Suggestion: "Implement the functionality or return a proper error",
			})
		}
	}
	return findings
}

// RustExpectGeneric detects .expect() with generic single-word messages.
type RustExpectGeneric struct{}

func (r *RustExpectGeneric) Meta() Rule {
	return Rule{
		ID:          "rust.expect-generic",
		Category:    "rust",
		Name:        "Generic .expect() message",
		Description: ".expect() with a generic message provides no context when it panics",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsExpectGenericPattern = regexp.MustCompile(`\.expect\("(error|failed|fail|err|none|invalid)"\)`)

func (r *RustExpectGeneric) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsExpectGenericPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       r.Meta().ID,
				Severity:   r.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Generic .expect() message provides no context on failure",
				Suggestion: "Use a descriptive message: .expect(\"failed to open config file\")",
			})
		}
	}
	return findings
}

// RustBoxedError detects -> Box<dyn Error> return types.
type RustBoxedError struct{}

func (r *RustBoxedError) Meta() Rule {
	return Rule{
		ID:          "rust.boxed-error",
		Category:    "rust",
		Name:        "Box<dyn Error> return type",
		Description: "Box<dyn Error> erases error type information — consider a concrete error type",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsBoxedErrorPattern = regexp.MustCompile(`->\s*.*Box<dyn\s+(std::error::)?Error>`)

func (r *RustBoxedError) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsBoxedErrorPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       r.Meta().ID,
				Severity:   r.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "Box<dyn Error> erases error type information, making matching impossible",
				Suggestion: "Define a concrete error enum or use thiserror/anyhow for better error handling",
			})
		}
	}
	return findings
}

// RustProcessExit detects process::exit() calls.
type RustProcessExit struct{}

func (r *RustProcessExit) Meta() Rule {
	return Rule{
		ID:          "rust.process-exit",
		Category:    "rust",
		Name:        "process::exit() call",
		Description: "process::exit() bypasses destructors and cleanup — return from main instead",
		Severity:    finding.SeverityWarning,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsProcessExitPattern = regexp.MustCompile(`(std::process|process)::exit\(`)

func (r *RustProcessExit) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsProcessExitPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       r.Meta().ID,
				Severity:   r.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "process::exit() skips destructors and cleanup code",
				Suggestion: "Return an ExitCode from main() or propagate errors upward instead",
			})
		}
	}
	return findings
}

// --- INFO severity rules ---

// RustAllowSuppress detects #[allow(...)] without an explaining comment on the next line.
type RustAllowSuppress struct{}

func (r *RustAllowSuppress) Meta() Rule {
	return Rule{
		ID:          "rust.allow-suppress",
		Category:    "rust",
		Name:        "#[allow(...)] without explanation",
		Description: "#[allow(dead_code|unused_variables|unused_imports)] should have a comment explaining why",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsAllowSuppressPattern = regexp.MustCompile(`#\[allow\((dead_code|unused_variables|unused_imports)\)\]`)

func (r *RustAllowSuppress) CheckLines(path string, lines []string) []finding.Finding {
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsAllowSuppressPattern.MatchString(line) {
			// Check if the NEXT line is a comment explaining why.
			hasExplanation := false
			if i+1 < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[i+1])
				if strings.HasPrefix(nextTrimmed, "//") {
					hasExplanation = true
				}
			}
			if !hasExplanation {
				findings = append(findings, finding.Finding{
					Rule:       r.Meta().ID,
					Severity:   r.Meta().Severity,
					File:       path,
					Line:       i + 1,
					Code:       line,
					Message:    "#[allow(...)] suppresses a useful lint without explanation",
					Suggestion: "Add a // comment on the next line explaining why this suppression is needed",
				})
			}
		}
	}
	return findings
}

// RustEprintlnInLib detects eprintln!() in library files (src/lib.rs or lib/ directory).
type RustEprintlnInLib struct{}

func (r *RustEprintlnInLib) Meta() Rule {
	return Rule{
		ID:          "rust.eprintln-in-lib",
		Category:    "rust",
		Name:        "eprintln!() in library code",
		Description: "Libraries should return errors, not print to stderr — leave output to the caller",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".rs"},
		ScanType:    ScanTypeRegex,
	}
}

var rsEprintlnPattern = regexp.MustCompile(`eprintln!\(`)

func (r *RustEprintlnInLib) CheckLines(path string, lines []string) []finding.Finding {
	norm := filepath.ToSlash(path)
	base := filepath.Base(path)
	isLib := base == "lib.rs" || strings.Contains(norm, "/lib/")
	if !isLib {
		return nil
	}
	var findings []finding.Finding
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if rsEprintlnPattern.MatchString(line) {
			findings = append(findings, finding.Finding{
				Rule:       r.Meta().ID,
				Severity:   r.Meta().Severity,
				File:       path,
				Line:       i + 1,
				Code:       line,
				Message:    "eprintln!() in library code — libraries should return errors, not print",
				Suggestion: "Return an error or use the log/tracing crate instead of eprintln!()",
			})
		}
	}
	return findings
}
