package rules

import "testing"

// --- PyHardcodedCredentials ---

func TestPyHardcodedCredentials_Positive(t *testing.T) {
	checker := &PyHardcodedCredentials{}
	lines := []string{
		`password = "supersecret123"`,
		`api_key = "sk-1234567890abcdef"`,
		`secret = "my-secret-value"`,
		`token = "bearer-abc-xyz"`,
	}
	findings := checker.CheckLines("src/config.py", lines)
	if len(findings) != 4 {
		t.Fatalf("expected 4 findings, got %d", len(findings))
	}
	if findings[0].Rule != "config-smells.hardcoded-credentials-py" {
		t.Fatalf("expected rule config-smells.hardcoded-credentials-py, got %s", findings[0].Rule)
	}
}

func TestPyHardcodedCredentials_SingleQuotes(t *testing.T) {
	checker := &PyHardcodedCredentials{}
	lines := []string{`password = 'mysecret'`}
	findings := checker.CheckLines("src/db.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestPyHardcodedCredentials_CaseInsensitive(t *testing.T) {
	checker := &PyHardcodedCredentials{}
	lines := []string{
		`PASSWORD = "value"`,
		`Api_Key = "value"`,
		`SECRET_KEY = "value"`,
	}
	findings := checker.CheckLines("src/settings.py", lines)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
}

func TestPyHardcodedCredentials_EnvVarClean(t *testing.T) {
	checker := &PyHardcodedCredentials{}
	lines := []string{
		`password = os.environ["DB_PASSWORD"]`,
		`api_key = os.environ.get("API_KEY")`,
	}
	findings := checker.CheckLines("src/config.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for env vars, got %d", len(findings))
	}
}

func TestPyHardcodedCredentials_EmptyStringClean(t *testing.T) {
	checker := &PyHardcodedCredentials{}
	lines := []string{`secret = ""`}
	findings := checker.CheckLines("src/config.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for empty string, got %d", len(findings))
	}
}

func TestPyHardcodedCredentials_TestFileSkipped(t *testing.T) {
	checker := &PyHardcodedCredentials{}
	lines := []string{`password = "testpass"`}
	findings := checker.CheckLines("tests/test_auth.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings in test file, got %d", len(findings))
	}
}

func TestPyHardcodedCredentials_CommentSkipped(t *testing.T) {
	checker := &PyHardcodedCredentials{}
	lines := []string{`# password = "old_value"`}
	findings := checker.CheckLines("src/config.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for commented line, got %d", len(findings))
	}
}

func TestPyHardcodedCredentials_AdditionalKeywords(t *testing.T) {
	checker := &PyHardcodedCredentials{}
	lines := []string{
		`passwd = "mypassword"`,
		`apikey = "key123"`,
		`auth_token = "tok"`,
	}
	findings := checker.CheckLines("src/auth.py", lines)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
}

// --- PyRequirementsUnpinned ---

func TestPyRequirementsUnpinned_Positive(t *testing.T) {
	checker := &PyRequirementsUnpinned{}
	lines := []string{
		`# Dependencies`,
		`requests`,
		`flask>=2.0`,
		`django~=4.0`,
		`numpy`,
	}
	findings := checker.CheckLines("requirements.txt", lines)
	if len(findings) != 4 {
		t.Fatalf("expected 4 findings, got %d", len(findings))
	}
	if findings[0].Rule != "config-smells.requirements-unpinned" {
		t.Fatalf("expected rule config-smells.requirements-unpinned, got %s", findings[0].Rule)
	}
}

func TestPyRequirementsUnpinned_PinnedClean(t *testing.T) {
	checker := &PyRequirementsUnpinned{}
	lines := []string{
		`# Dependencies (properly pinned)`,
		`requests==2.28.0`,
		`flask==2.3.2`,
		`django==4.2.1`,
	}
	findings := checker.CheckLines("requirements.txt", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for pinned deps, got %d", len(findings))
	}
}

func TestPyRequirementsUnpinned_SkipsComments(t *testing.T) {
	checker := &PyRequirementsUnpinned{}
	lines := []string{`# this is a comment`, ``, `-r base.txt`}
	findings := checker.CheckLines("requirements.txt", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for comments, got %d", len(findings))
	}
}

func TestPyRequirementsUnpinned_NonRequirementsFileSkipped(t *testing.T) {
	checker := &PyRequirementsUnpinned{}
	lines := []string{`requests`, `flask`}
	findings := checker.CheckLines("packages.txt", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-requirements file, got %d", len(findings))
	}
}

func TestPyRequirementsUnpinned_RequirementsDev(t *testing.T) {
	checker := &PyRequirementsUnpinned{}
	lines := []string{`pytest`, `coverage`}
	findings := checker.CheckLines("requirements-dev.txt", lines)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestPyRequirementsUnpinned_OptionsSkipped(t *testing.T) {
	checker := &PyRequirementsUnpinned{}
	lines := []string{`--index-url https://pypi.org/simple/`, `-r base-requirements.txt`}
	findings := checker.CheckLines("requirements.txt", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for options, got %d", len(findings))
	}
}
