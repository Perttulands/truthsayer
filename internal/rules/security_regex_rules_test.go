package rules

import "testing"

// --- UnusedVariable ---

func TestUnusedVariable_GoUnused(t *testing.T) {
	checker := &UnusedVariable{}
	lines := []string{
		"package p",
		"func foo() {",
		"\tresult := compute()",
		"}",
	}
	findings := checker.CheckLines("handler.go", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestUnusedVariable_GoUsedTwice(t *testing.T) {
	checker := &UnusedVariable{}
	lines := []string{
		"package p",
		"func foo() {",
		"\tresult := compute()",
		"\tfmt.Println(result)",
		"}",
	}
	findings := checker.CheckLines("handler.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for used variable, got %d", len(findings))
	}
}

func TestUnusedVariable_GoUnderscore(t *testing.T) {
	checker := &UnusedVariable{}
	lines := []string{
		"package p",
		"func foo() {",
		"\t_ = compute()",
		"}",
	}
	findings := checker.CheckLines("handler.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for _ variable, got %d", len(findings))
	}
}

func TestUnusedVariable_JSUnused(t *testing.T) {
	checker := &UnusedVariable{}
	lines := []string{
		"const result = compute()",
	}
	findings := checker.CheckLines("app.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for unused JS const, got %d", len(findings))
	}
}

func TestUnusedVariable_PyUnused(t *testing.T) {
	checker := &UnusedVariable{}
	lines := []string{
		"result = compute()",
	}
	findings := checker.CheckLines("app.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for unused Python variable, got %d", len(findings))
	}
}

func TestUnusedVariable_GoPackageLevel(t *testing.T) {
	checker := &UnusedVariable{}
	lines := []string{
		"package p",
		"var GlobalConfig = loadConfig()",
	}
	findings := checker.CheckLines("config.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for package-level var, got %d", len(findings))
	}
}

// --- ErrorSwallowing ---

func TestErrorSwallowing_JSEmptyCatch(t *testing.T) {
	checker := &ErrorSwallowing{}
	lines := []string{
		"try { doSomething() } catch (e) {}",
	}
	findings := checker.CheckLines("app.js", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestErrorSwallowing_JSCatchWithBody(t *testing.T) {
	checker := &ErrorSwallowing{}
	lines := []string{
		"try {",
		"  doSomething()",
		"} catch (e) {",
		"  console.error(e)",
		"}",
	}
	findings := checker.CheckLines("app.js", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for catch with body, got %d", len(findings))
	}
}

func TestErrorSwallowing_JSMultilineCatchEmpty(t *testing.T) {
	checker := &ErrorSwallowing{}
	lines := []string{
		"try {",
		"  doSomething()",
		"} catch (e) {",
		"}",
	}
	findings := checker.CheckLines("app.ts", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for multiline empty catch, got %d", len(findings))
	}
}

func TestErrorSwallowing_PyExceptPass(t *testing.T) {
	checker := &ErrorSwallowing{}
	lines := []string{
		"except Exception: pass",
	}
	findings := checker.CheckLines("app.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for except pass, got %d", len(findings))
	}
}

func TestErrorSwallowing_PyExceptWithHandler(t *testing.T) {
	checker := &ErrorSwallowing{}
	lines := []string{
		"except Exception:",
		"    logger.error(e)",
	}
	findings := checker.CheckLines("app.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for except with handler, got %d", len(findings))
	}
}

func TestErrorSwallowing_GoFileIgnored(t *testing.T) {
	checker := &ErrorSwallowing{}
	lines := []string{
		"if err != nil { return nil }",
	}
	findings := checker.CheckLines("app.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for .go file, got %d", len(findings))
	}
}

// --- HardcodedCredentials ---

func TestHardcodedCredentials_PasswordLiteral(t *testing.T) {
	checker := &HardcodedCredentials{}
	lines := []string{
		`password = "mysecretpassword123"`,
	}
	findings := checker.CheckLines("config.go", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestHardcodedCredentials_ApiKey(t *testing.T) {
	checker := &HardcodedCredentials{}
	lines := []string{
		`api_key = "sk-1234567890abcdef"`,
	}
	findings := checker.CheckLines("config.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for api_key, got %d", len(findings))
	}
}

func TestHardcodedCredentials_EnvVarRef(t *testing.T) {
	checker := &HardcodedCredentials{}
	lines := []string{
		`password = os.Getenv("DB_PASSWORD")`,
	}
	findings := checker.CheckLines("config.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for env var reference, got %d", len(findings))
	}
}

func TestHardcodedCredentials_Comment(t *testing.T) {
	checker := &HardcodedCredentials{}
	lines := []string{
		`// password = "mysecretpassword123"`,
	}
	findings := checker.CheckLines("config.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for comment, got %d", len(findings))
	}
}

// --- SQLInjection ---

func TestSQLInjection_ConcatPattern(t *testing.T) {
	checker := &SQLInjection{}
	lines := []string{
		`query := "SELECT * FROM users WHERE id = " + userID`,
	}
	findings := checker.CheckLines("db.go", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestSQLInjection_FmtSprintf(t *testing.T) {
	checker := &SQLInjection{}
	lines := []string{
		`query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", id)`,
	}
	findings := checker.CheckLines("db.go", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for fmt.Sprintf SQL, got %d", len(findings))
	}
}

func TestSQLInjection_PythonFString(t *testing.T) {
	checker := &SQLInjection{}
	lines := []string{
		`query = f"SELECT * FROM users WHERE id = {user_id}"`,
	}
	findings := checker.CheckLines("db.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for Python f-string SQL, got %d", len(findings))
	}
}

func TestSQLInjection_ParameterizedQuery(t *testing.T) {
	checker := &SQLInjection{}
	lines := []string{
		`db.Query("SELECT * FROM users WHERE id = $1", userID)`,
	}
	findings := checker.CheckLines("db.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for parameterized query, got %d", len(findings))
	}
}

func TestSQLInjection_CommentLine(t *testing.T) {
	checker := &SQLInjection{}
	lines := []string{
		`// query := "SELECT * FROM users WHERE id = " + userID`,
	}
	findings := checker.CheckLines("db.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for comment, got %d", len(findings))
	}
}

// --- CommandInjection ---

func TestCommandInjection_ShellTrue(t *testing.T) {
	checker := &CommandInjection{}
	lines := []string{
		`subprocess.run(cmd, shell=True)`,
	}
	findings := checker.CheckLines("deploy.py", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestCommandInjection_ExecShDashC(t *testing.T) {
	checker := &CommandInjection{}
	lines := []string{
		`exec.Command("sh", "-c", cmd)`,
	}
	findings := checker.CheckLines("run.go", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for exec sh -c, got %d", len(findings))
	}
}

func TestCommandInjection_SafeExec(t *testing.T) {
	checker := &CommandInjection{}
	lines := []string{
		`exec.Command("ls", "-la", "/tmp")`,
	}
	findings := checker.CheckLines("run.go", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for safe exec, got %d", len(findings))
	}
}

func TestCommandInjection_Comment(t *testing.T) {
	checker := &CommandInjection{}
	lines := []string{
		`# subprocess.run(cmd, shell=True)`,
	}
	findings := checker.CheckLines("deploy.py", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for comment, got %d", len(findings))
	}
}
