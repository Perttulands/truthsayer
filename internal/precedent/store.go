package precedent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultPath is the default file used for precedent storage.
const DefaultPath = "precedents.json"

// Decision captures the historical judgment for a rule violation.
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
)

// Precedent represents a past decision that can inform future judgments.
type Precedent struct {
	RuleID        string    `json:"rule_id"`
	ViolationHash string    `json:"violation_hash"`
	Decision      Decision  `json:"decision"`
	Rationale     string    `json:"rationale"`
	CreatedAt     time.Time `json:"created_at"`
}

// Validate checks required fields and supported decision values.
func (p Precedent) Validate() error {
	if strings.TrimSpace(p.RuleID) == "" {
		return errors.New("precedent: rule_id is required")
	}
	if strings.TrimSpace(p.ViolationHash) == "" {
		return errors.New("precedent: violation_hash is required")
	}
	if strings.TrimSpace(string(p.Decision)) == "" {
		return errors.New("precedent: decision is required")
	}
	if p.Decision != DecisionAllow && p.Decision != DecisionDeny {
		return fmt.Errorf("precedent: invalid decision %q", p.Decision)
	}
	if strings.TrimSpace(p.Rationale) == "" {
		return errors.New("precedent: rationale is required")
	}
	if p.CreatedAt.IsZero() {
		return errors.New("precedent: created_at is required")
	}
	return nil
}

// Store provides file-based persistence for precedents.
type Store struct {
	path string
}

// NewStore creates a new file-backed precedent store.
// If path is empty, precedents.json in the current working directory is used.
func NewStore(path string) *Store {
	if strings.TrimSpace(path) == "" {
		path = DefaultPath
	}
	return &Store{path: path}
}

// Path returns the JSON path used by this store.
func (s *Store) Path() string {
	return s.path
}

// Load reads precedents from disk. Missing files return an empty list.
func (s *Store) Load() ([]Precedent, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Precedent{}, nil
		}
		return nil, fmt.Errorf("precedent: read %s: %w", s.path, err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return []Precedent{}, nil
	}

	var precedents []Precedent
	if err := json.Unmarshal(data, &precedents); err != nil {
		return nil, fmt.Errorf("precedent: decode %s: %w", s.path, err)
	}

	for i, p := range precedents {
		if err := p.Validate(); err != nil {
			return nil, fmt.Errorf("precedent: invalid record %d: %w", i, err)
		}
	}

	return precedents, nil
}

// Save writes precedents to disk as pretty-printed JSON.
func (s *Store) Save(precedents []Precedent) error {
	for i, p := range precedents {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("precedent: record %d: %w", i, err)
		}
	}

	out, err := json.MarshalIndent(precedents, "", "  ")
	if err != nil {
		return fmt.Errorf("precedent: encode %s: %w", s.path, err)
	}
	out = append(out, '\n')

	dir := filepath.Dir(s.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("precedent: create dir %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(s.path, out, 0o644); err != nil {
		return fmt.Errorf("precedent: write %s: %w", s.path, err)
	}
	return nil
}

// Add appends a single precedent to file-backed storage.
// created_at defaults to current UTC time if not set.
func (s *Store) Add(p Precedent) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}

	precedents, err := s.Load()
	if err != nil {
		return err
	}
	precedents = append(precedents, p)
	return s.Save(precedents)
}

// Query returns the most recent matching precedent by rule_id + violation_hash.
func (s *Store) Query(ruleID, violationHash string) (Precedent, bool, error) {
	precedents, err := s.Load()
	if err != nil {
		return Precedent{}, false, err
	}
	p, ok := Query(precedents, ruleID, violationHash)
	return p, ok, nil
}

// Query finds the most recent matching precedent in a list.
func Query(precedents []Precedent, ruleID, violationHash string) (Precedent, bool) {
	for i := len(precedents) - 1; i >= 0; i-- {
		p := precedents[i]
		if p.RuleID == ruleID && p.ViolationHash == violationHash {
			return p, true
		}
	}
	return Precedent{}, false
}

// QueryByRule returns all precedents for a given rule_id.
func QueryByRule(precedents []Precedent, ruleID string) []Precedent {
	filtered := make([]Precedent, 0, len(precedents))
	for _, p := range precedents {
		if p.RuleID == ruleID {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
