package senate

import (
	"strings"
	"testing"
)

func TestParseVerdict_JSON(t *testing.T) {
	v, err := ParseVerdict([]byte(`{
  "id":"quick-1771535739",
  "status":"approved",
  "amendments":[{"rule_id":"silent-fallback.hidden-failure-bash","action":"set_severity","severity":"warning"}]
}`))
	if err != nil {
		t.Fatalf("parse verdict failed: %v", err)
	}
	if v.ID != "quick-1771535739" {
		t.Fatalf("unexpected verdict id: %q", v.ID)
	}
}

func TestParseVerdict_FencedJSON(t *testing.T) {
	doc := `
# Senate Verdict

` + "```json" + `
{"id":"v-1","status":"approved","amendments":[{"rule_id":"r","action":"disable_rule"}]}
` + "```" + `
`
	_, err := ParseVerdict([]byte(doc))
	if err != nil {
		t.Fatalf("parse fenced verdict failed: %v", err)
	}
}

func TestParseVerdict_InvalidAction(t *testing.T) {
	_, err := ParseVerdict([]byte(`{"id":"v","status":"approved","amendments":[{"rule_id":"r","action":"nope"}]}`))
	if err == nil {
		t.Fatal("expected invalid action error")
	}
	if !strings.Contains(err.Error(), "invalid action") {
		t.Fatalf("unexpected error: %v", err)
	}
}

