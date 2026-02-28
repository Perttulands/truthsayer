package debt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultPath is where advisory debt is persisted by default.
const DefaultPath = ".truthsayer-debt.json"

// Entry represents one advisory decision captured as technical debt.
type Entry struct {
	RuleID    string    `json:"rule_id"`
	File      string    `json:"file"`
	Line      int       `json:"line"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Reasoning string    `json:"reasoning"`
	CreatedAt time.Time `json:"created_at"`
}

// Validate checks that required fields are present.
func (e Entry) Validate() error {
	if strings.TrimSpace(e.RuleID) == "" {
		return errors.New("debt: rule_id is required")
	}
	if strings.TrimSpace(e.File) == "" {
		return errors.New("debt: file is required")
	}
	if e.Line <= 0 {
		return errors.New("debt: line must be > 0")
	}
	if strings.TrimSpace(e.Reasoning) == "" {
		return errors.New("debt: reasoning is required")
	}
	if e.CreatedAt.IsZero() {
		return errors.New("debt: created_at is required")
	}
	return nil
}

// Store provides file-backed advisory debt persistence.
type Store struct {
	path string
}

// NewStore creates a debt store for the given path.
func NewStore(path string) *Store {
	if strings.TrimSpace(path) == "" {
		path = DefaultPath
	}
	return &Store{path: path}
}

// Load returns all debt entries.
func (s *Store) Load() ([]Entry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("debt: read %s: %w", s.path, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return []Entry{}, nil
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("debt: decode %s: %w", s.path, err)
	}
	for i, e := range entries {
		if err := e.Validate(); err != nil {
			return nil, fmt.Errorf("debt: invalid entry %d: %w", i, err)
		}
	}
	return entries, nil
}

// Save writes all entries.
func (s *Store) Save(entries []Entry) error {
	for i, e := range entries {
		if err := e.Validate(); err != nil {
			return fmt.Errorf("debt: entry %d: %w", i, err)
		}
	}
	out, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("debt: encode %s: %w", s.path, err)
	}
	out = append(out, '\n')
	dir := filepath.Dir(s.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("debt: create dir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(s.path, out, 0o644); err != nil {
		return fmt.Errorf("debt: write %s: %w", s.path, err)
	}
	return nil
}

// Add appends a new advisory debt entry.
func (s *Store) Add(e Entry) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	entries, err := s.Load()
	if err != nil {
		return fmt.Errorf("load debt entries: %w", err)
	}
	entries = append(entries, e)
	return s.Save(entries)
}
