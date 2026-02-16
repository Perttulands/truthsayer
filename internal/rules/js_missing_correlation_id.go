package rules

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/perttulands/truthsayer/internal/finding"
)

// JSMissingCorrelationID detects Express/Koa middleware chains that handle
// requests but never set or propagate a correlation/request ID.
type JSMissingCorrelationID struct{}

func (j *JSMissingCorrelationID) Meta() Rule {
	return Rule{
		ID:          "trace-gaps.js-missing-correlation-id",
		Category:    "trace-gaps",
		Name:        "Missing correlation ID in middleware",
		Description: "Express/Koa app handles requests without setting correlation or request ID",
		Severity:    finding.SeverityInfo,
		FileTypes:   []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		ScanType:    ScanTypeAST,
	}
}

// correlationIDPatterns are strings whose presence in source indicates correlation ID usage.
var correlationIDPatterns = []string{
	"correlation-id", "correlationId", "correlation_id",
	"request-id", "requestId", "request_id",
	"x-request-id", "X-Request-Id", "X-Request-ID",
	"x-correlation-id", "X-Correlation-Id", "X-Correlation-ID",
	"trace-id", "traceId", "trace_id",
	"express-request-id", "cls-hooked", "express-correlation-id",
	"uuid", "nanoid", "crypto.randomUUID",
}

func (j *JSMissingCorrelationID) CheckJSAST(tree *sitter.Tree, source []byte, path string) []finding.Finding {
	if jsIsTestFile(path) {
		return nil
	}

	info := jsFindExpressApp(tree.RootNode(), source)
	if info == nil || !info.hasRoute {
		return nil
	}

	srcText := string(source)
	for _, pattern := range correlationIDPatterns {
		if strings.Contains(srcText, pattern) {
			return nil
		}
	}

	line := jsLineNumber(info.firstRouteNode)
	return []finding.Finding{{
		Rule:       j.Meta().ID,
		Severity:   j.Meta().Severity,
		File:       path,
		Line:       line,
		Code:       jsSourceLine(source, line),
		Message:    "HTTP server defines routes but no correlation/request ID is set",
		Suggestion: "Add middleware to generate and propagate a request ID (e.g., req.id = crypto.randomUUID())",
	}}
}
