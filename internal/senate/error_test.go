package senate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- Verdict parsing error paths ---

func TestParseVerdict_Empty(t *testing.T) {
	_, err := ParseVerdict([]byte(""))
	if err == nil {
		t.Fatal("expected error for empty verdict")
	}
	if !strings.Contains(err.Error(), "empty verdict") {
		t.Fatalf("expected 'empty verdict' in error, got %v", err)
	}
}

func TestParseVerdict_WhitespaceOnly(t *testing.T) {
	_, err := ParseVerdict([]byte("   \n  "))
	if err == nil {
		t.Fatal("expected error for whitespace-only verdict")
	}
}

func TestParseVerdict_InvalidJSON(t *testing.T) {
	_, err := ParseVerdict([]byte("{broken json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "decode verdict json") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestParseVerdict_NoFencedJSON(t *testing.T) {
	_, err := ParseVerdict([]byte("Just some markdown without JSON"))
	if err == nil {
		t.Fatal("expected error for missing JSON")
	}
	if !strings.Contains(err.Error(), "expected JSON object or fenced") {
		t.Fatalf("expected fenced error, got %v", err)
	}
}

func TestParseVerdict_MissingID(t *testing.T) {
	_, err := ParseVerdict([]byte(`{"id":"","status":"rejected"}`))
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if !strings.Contains(err.Error(), "verdict id") {
		t.Fatalf("expected 'verdict id' in error, got %v", err)
	}
}

func TestParseVerdict_InvalidStatus(t *testing.T) {
	_, err := ParseVerdict([]byte(`{"id":"v1","status":"pending"}`))
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("expected 'invalid status' in error, got %v", err)
	}
}

func TestParseVerdict_ApprovedNoAmendments(t *testing.T) {
	_, err := ParseVerdict([]byte(`{"id":"v1","status":"approved","amendments":[]}`))
	if err == nil {
		t.Fatal("expected error for approved with no amendments")
	}
	if !strings.Contains(err.Error(), "must include amendments") {
		t.Fatalf("expected 'must include amendments' in error, got %v", err)
	}
}

func TestParseVerdict_RejectedNoAmendments(t *testing.T) {
	v, err := ParseVerdict([]byte(`{"id":"v1","status":"rejected"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != StatusRejected {
		t.Fatalf("expected rejected status, got %q", v.Status)
	}
}

func TestParseVerdict_KeepAsIs(t *testing.T) {
	v, err := ParseVerdict([]byte(`{"id":"v1","status":"keep_as_is"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != StatusKeepAsIs {
		t.Fatalf("expected keep_as_is status, got %q", v.Status)
	}
}

// --- Amendment validation ---

func TestAmendmentValidate_EmptyRuleID(t *testing.T) {
	a := Amendment{RuleID: "", Action: ActionDisableRule}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "rule_id") {
		t.Fatalf("expected rule_id error, got %v", err)
	}
}

func TestAmendmentValidate_InvalidAction(t *testing.T) {
	a := Amendment{RuleID: "r", Action: "explode"}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid action") {
		t.Fatalf("expected invalid action error, got %v", err)
	}
}

func TestAmendmentValidate_SetSeverity_InvalidSeverity(t *testing.T) {
	a := Amendment{RuleID: "r", Action: ActionSetSeverity, Severity: "critical"}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "set_severity requires severity") {
		t.Fatalf("expected severity error, got %v", err)
	}
}

func TestAmendmentValidate_SetSeverity_Valid(t *testing.T) {
	for _, sev := range []string{"error", "warning", "info"} {
		a := Amendment{RuleID: "r", Action: ActionSetSeverity, Severity: sev}
		if err := a.Validate(); err != nil {
			t.Fatalf("unexpected error for severity %q: %v", sev, err)
		}
	}
}

func TestAmendmentValidate_AddException_EmptyException(t *testing.T) {
	a := Amendment{RuleID: "r", Action: ActionAddException, Exception: ""}
	err := a.Validate()
	if err == nil || !strings.Contains(err.Error(), "add_exception requires exception") {
		t.Fatalf("expected exception error, got %v", err)
	}
}

func TestAmendmentValidate_AddException_Valid(t *testing.T) {
	a := Amendment{RuleID: "r", Action: ActionAddException, Exception: "test pattern"}
	if err := a.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAmendmentValidate_DisableEnableRule(t *testing.T) {
	for _, action := range []string{ActionDisableRule, ActionEnableRule} {
		a := Amendment{RuleID: "r", Action: action}
		if err := a.Validate(); err != nil {
			t.Fatalf("unexpected error for action %q: %v", action, err)
		}
	}
}

// --- AppliedAmendment validation ---

func TestAppliedAmendment_ValidateMissingVerdictID(t *testing.T) {
	a := AppliedAmendment{
		VerdictID: "",
		AppliedAt: time.Now(),
		Amendment: Amendment{RuleID: "r", Action: ActionDisableRule},
	}
	if err := a.validate(); err == nil || !strings.Contains(err.Error(), "verdict_id") {
		t.Fatalf("expected verdict_id error, got %v", err)
	}
}

func TestAppliedAmendment_ValidateZeroAppliedAt(t *testing.T) {
	a := AppliedAmendment{
		VerdictID: "v1",
		Amendment: Amendment{RuleID: "r", Action: ActionDisableRule},
	}
	if err := a.validate(); err == nil || !strings.Contains(err.Error(), "applied_at") {
		t.Fatalf("expected applied_at error, got %v", err)
	}
}

func TestAppliedAmendment_ValidateInvalidAmendment(t *testing.T) {
	a := AppliedAmendment{
		VerdictID: "v1",
		AppliedAt: time.Now(),
		Amendment: Amendment{RuleID: "", Action: ActionDisableRule},
	}
	if err := a.validate(); err == nil || !strings.Contains(err.Error(), "amendment validation") {
		t.Fatalf("expected amendment validation error, got %v", err)
	}
}

// --- ParseVerdictFile ---

func TestParseVerdictFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "verdict.json")
	data := `{"id":"v1","status":"rejected"}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	v, err := ParseVerdictFile(path)
	if err != nil {
		t.Fatalf("parse verdict file: %v", err)
	}
	if v.ID != "v1" {
		t.Fatalf("expected id 'v1', got %q", v.ID)
	}
}

func TestParseVerdictFile_NonexistentFile(t *testing.T) {
	_, err := ParseVerdictFile("/nonexistent/verdict.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "read verdict file") {
		t.Fatalf("expected 'read verdict file' in error, got %v", err)
	}
}

// --- AmendmentStore error paths ---

func TestNewAmendmentStore_EmptyPath(t *testing.T) {
	store := NewAmendmentStore("")
	if store.path != DefaultAmendmentsPath {
		t.Fatalf("expected default path %q, got %q", DefaultAmendmentsPath, store.path)
	}
}

func TestAmendmentStore_LoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "amendments.json")
	if err := os.WriteFile(path, []byte("   "), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewAmendmentStore(path)
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load empty file: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0, got %d", len(got))
	}
}

func TestAmendmentStore_LoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "amendments.json")
	if err := os.WriteFile(path, []byte("{broken"), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewAmendmentStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decode amendments") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestAmendmentStore_LoadInvalidEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "amendments.json")
	// Valid JSON but invalid applied amendment (missing verdict_id)
	data := `[{"verdict_id":"","applied_at":"2026-02-20T00:00:00Z","amendment":{"rule_id":"r","action":"disable_rule"}}]`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewAmendmentStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid applied amendment") {
		t.Fatalf("expected 'invalid applied amendment' in error, got %v", err)
	}
}

func TestAmendmentStore_SaveInvalidEntry(t *testing.T) {
	store := NewAmendmentStore(filepath.Join(t.TempDir(), "amendments.json"))
	err := store.Save([]AppliedAmendment{{VerdictID: ""}})
	if err == nil {
		t.Fatal("expected validation error on save")
	}
}

func TestAmendmentStore_SaveSubdir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "dir", "amendments.json")
	store := NewAmendmentStore(path)
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	err := store.Save([]AppliedAmendment{{
		VerdictID: "v1",
		AppliedAt: now,
		Amendment: Amendment{RuleID: "r", Action: ActionDisableRule},
	}})
	if err != nil {
		t.Fatalf("save to subdir: %v", err)
	}
}

// --- ApplyVerdict error paths ---

func TestApplyVerdict_Rejected(t *testing.T) {
	store := NewAmendmentStore(filepath.Join(t.TempDir(), "amendments.json"))
	v := Verdict{ID: "v1", Status: StatusRejected}
	added, err := store.ApplyVerdict(v, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(added) != 0 {
		t.Fatalf("expected 0 added for rejected verdict, got %d", len(added))
	}
}

func TestApplyVerdict_InvalidVerdict(t *testing.T) {
	store := NewAmendmentStore(filepath.Join(t.TempDir(), "amendments.json"))
	v := Verdict{ID: "", Status: "bad"}
	_, err := store.ApplyVerdict(v, time.Now())
	if err == nil {
		t.Fatal("expected error for invalid verdict")
	}
}

func TestApplyVerdict_Idempotent(t *testing.T) {
	store := NewAmendmentStore(filepath.Join(t.TempDir(), "amendments.json"))
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	v := Verdict{
		ID:     "v1",
		Status: StatusApproved,
		Amendments: []Amendment{
			{RuleID: "r", Action: ActionDisableRule},
		},
	}
	added1, err := store.ApplyVerdict(v, now)
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if len(added1) != 1 {
		t.Fatalf("expected 1 added, got %d", len(added1))
	}
	// Apply again - should be idempotent
	added2, err := store.ApplyVerdict(v, now)
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if len(added2) != 0 {
		t.Fatalf("expected 0 added on duplicate apply, got %d", len(added2))
	}
}

func TestApplyVerdict_ZeroTime(t *testing.T) {
	store := NewAmendmentStore(filepath.Join(t.TempDir(), "amendments.json"))
	v := Verdict{
		ID:     "v1",
		Status: StatusApproved,
		Amendments: []Amendment{
			{RuleID: "r", Action: ActionEnableRule},
		},
	}
	before := time.Now().UTC()
	added, err := store.ApplyVerdict(v, time.Time{})
	if err != nil {
		t.Fatalf("apply with zero time: %v", err)
	}
	after := time.Now().UTC()
	if len(added) != 1 {
		t.Fatalf("expected 1 added, got %d", len(added))
	}
	if added[0].AppliedAt.Before(before) || added[0].AppliedAt.After(after) {
		t.Fatalf("expected AppliedAt near now, got %v", added[0].AppliedAt)
	}
}

func TestApplyVerdict_LoadError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "amendments.json")
	// Write corrupt data so Load fails
	if err := os.WriteFile(path, []byte("{broken"), 0644); err != nil {
		t.Fatal(err)
	}
	store := NewAmendmentStore(path)
	v := Verdict{
		ID:     "v1",
		Status: StatusApproved,
		Amendments: []Amendment{
			{RuleID: "r", Action: ActionDisableRule},
		},
	}
	_, err := store.ApplyVerdict(v, time.Now())
	if err == nil {
		t.Fatal("expected error when load fails during apply")
	}
}

// --- AppendAudit ---

func TestAppendAudit_Empty(t *testing.T) {
	err := AppendAudit(filepath.Join(t.TempDir(), "audit.jsonl"), nil)
	if err != nil {
		t.Fatalf("expected nil for empty applied, got %v", err)
	}
}

func TestAppendAudit_Subdir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "audit.jsonl")
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	err := AppendAudit(path, []AppliedAmendment{
		{VerdictID: "v1", AppliedAt: now, Amendment: Amendment{RuleID: "r", Action: ActionDisableRule}},
	})
	if err != nil {
		t.Fatalf("append audit to subdir: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected audit content")
	}
}

func TestAppendAudit_MultipleEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	now := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	entries := []AppliedAmendment{
		{VerdictID: "v1", AppliedAt: now, Amendment: Amendment{RuleID: "r1", Action: ActionDisableRule}},
		{VerdictID: "v2", AppliedAt: now, Amendment: Amendment{RuleID: "r2", Action: ActionEnableRule}},
	}
	if err := AppendAudit(path, entries); err != nil {
		t.Fatalf("first append: %v", err)
	}
	// Append again to test append behavior
	if err := AppendAudit(path, entries[:1]); err != nil {
		t.Fatalf("second append: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 audit lines, got %d", len(lines))
	}
}
