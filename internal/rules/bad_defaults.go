package rules

import (
	"regexp"
	"strings"

	"github.com/perttulands/truthsayer/internal/finding"
)

// MissingPipefail detects bash scripts without `set -euo pipefail`.
type MissingPipefail struct{}

func (m *MissingPipefail) Meta() Rule {
	return Rule{
		ID:          "bad-defaults.missing-pipefail",
		Category:    "bad-defaults",
		Name:        "Missing pipefail",
		Description: "Bash script without set -euo pipefail",
		Severity:    finding.SeverityError,
		FileTypes:   []string{".sh", ".bash"},
		ScanType:    ScanTypeRegex,
	}
}

var pipefailPattern = regexp.MustCompile(`set\s+-[a-zA-Z]*e[a-zA-Z]*u[a-zA-Z]*o\s+pipefail`)

func (m *MissingPipefail) CheckLines(path string, lines []string) []finding.Finding {
	// Check if any line has set -euo pipefail (or equivalent)
	for _, line := range lines {
		if pipefailPattern.MatchString(line) {
			return nil
		}
	}
	// Check for shebang — only flag bash scripts
	if len(lines) == 0 {
		return nil
	}
	if !strings.HasPrefix(lines[0], "#!/") || !strings.Contains(lines[0], "bash") {
		return nil
	}
	return []finding.Finding{
		{
			Rule:       m.Meta().ID,
			Severity:   m.Meta().Severity,
			File:       path,
			Line:       1,
			Code:       lines[0],
			Message:    "Bash script missing set -euo pipefail",
			Suggestion: "Add 'set -euo pipefail' after the shebang line",
		},
	}
}
