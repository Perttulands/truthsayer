package senate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultAmendmentsPath stores approved Senate amendments.
const DefaultAmendmentsPath = ".truthsayer-amendments.json"

// DefaultAmendmentsAuditPath stores append-only amendment application audit lines.
const DefaultAmendmentsAuditPath = ".truthsayer-amendments.audit.jsonl"

// AppliedAmendment records one applied Senate amendment.
type AppliedAmendment struct {
	VerdictID string    `json:"verdict_id"`
	AppliedAt time.Time `json:"applied_at"`
	Amendment Amendment `json:"amendment"`
}

func (a AppliedAmendment) validate() error {
	if strings.TrimSpace(a.VerdictID) == "" {
		return errors.New("senate: verdict_id is required")
	}
	if a.AppliedAt.IsZero() {
		return errors.New("senate: applied_at is required")
	}
	if err := a.Amendment.Validate(); err != nil {
		return err
	}
	return nil
}

// AmendmentStore persists applied amendments.
type AmendmentStore struct {
	path string
}

// NewAmendmentStore creates an amendment store for path.
func NewAmendmentStore(path string) *AmendmentStore {
	if strings.TrimSpace(path) == "" {
		path = DefaultAmendmentsPath
	}
	return &AmendmentStore{path: path}
}

// Load returns applied amendments.
func (s *AmendmentStore) Load() ([]AppliedAmendment, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []AppliedAmendment{}, nil
		}
		return nil, fmt.Errorf("senate: read amendments %s: %w", s.path, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return []AppliedAmendment{}, nil
	}
	var out []AppliedAmendment
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("senate: decode amendments %s: %w", s.path, err)
	}
	for i, a := range out {
		if err := a.validate(); err != nil {
			return nil, fmt.Errorf("senate: invalid applied amendment %d: %w", i, err)
		}
	}
	return out, nil
}

// Save writes applied amendments.
func (s *AmendmentStore) Save(amendments []AppliedAmendment) error {
	for i, a := range amendments {
		if err := a.validate(); err != nil {
			return fmt.Errorf("senate: applied amendment %d: %w", i, err)
		}
	}
	out, err := json.MarshalIndent(amendments, "", "  ")
	if err != nil {
		return fmt.Errorf("senate: encode amendments %s: %w", s.path, err)
	}
	out = append(out, '\n')
	dir := filepath.Dir(s.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("senate: create dir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(s.path, out, 0o644); err != nil {
		return fmt.Errorf("senate: write amendments %s: %w", s.path, err)
	}
	return nil
}

func appliedKey(a AppliedAmendment) string {
	return strings.Join([]string{
		a.VerdictID,
		a.Amendment.RuleID,
		a.Amendment.Action,
		a.Amendment.Severity,
		a.Amendment.Exception,
	}, "\x00")
}

// ApplyVerdict persists approved verdict amendments (idempotent by verdict+amendment key).
func (s *AmendmentStore) ApplyVerdict(v Verdict, appliedAt time.Time) ([]AppliedAmendment, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if v.Status != StatusApproved {
		return []AppliedAmendment{}, nil
	}
	if appliedAt.IsZero() {
		appliedAt = time.Now().UTC()
	}

	existing, err := s.Load()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(existing))
	for _, a := range existing {
		seen[appliedKey(a)] = struct{}{}
	}

	added := make([]AppliedAmendment, 0, len(v.Amendments))
	for _, amendment := range v.Amendments {
		rec := AppliedAmendment{
			VerdictID: v.ID,
			AppliedAt: appliedAt.UTC(),
			Amendment: amendment,
		}
		key := appliedKey(rec)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		existing = append(existing, rec)
		added = append(added, rec)
	}
	if err := s.Save(existing); err != nil {
		return nil, err
	}
	return added, nil
}

// AppendAudit appends applied amendment records as JSON lines.
func AppendAudit(path string, applied []AppliedAmendment) error {
	if len(applied) == 0 {
		return nil
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("senate: create dir %s: %w", dir, err)
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("senate: open audit %s: %w", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, a := range applied {
		if err := enc.Encode(a); err != nil {
			return fmt.Errorf("senate: append audit %s: %w", path, err)
		}
	}
	return nil
}
