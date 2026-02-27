package rules

import "testing"

func TestSecretInConfig_PasswordMatch(t *testing.T) {
	checker := &SecretInConfig{}
	lines := []string{
		`password = "supersecretvalue123"`,
	}
	findings := checker.CheckLines("config.toml", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Errorf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestSecretInConfig_ApiKey(t *testing.T) {
	checker := &SecretInConfig{}
	lines := []string{
		`api_key = "sk-1234567890abcdef"`,
	}
	findings := checker.CheckLines("config.yaml", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for api_key, got %d", len(findings))
	}
}

func TestSecretInConfig_NoSecret(t *testing.T) {
	checker := &SecretInConfig{}
	lines := []string{
		`host = "localhost"`,
		`port = 8080`,
	}
	findings := checker.CheckLines("config.toml", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestSecretInConfig_CommentSkipped(t *testing.T) {
	checker := &SecretInConfig{}
	lines := []string{
		`# password = "supersecret123"`,
	}
	findings := checker.CheckLines("config.toml", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for comment, got %d", len(findings))
	}
}

func TestSecretInConfig_ShortValue(t *testing.T) {
	checker := &SecretInConfig{}
	lines := []string{
		`password = "short"`,
	}
	findings := checker.CheckLines("config.toml", lines)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for short value, got %d", len(findings))
	}
}
