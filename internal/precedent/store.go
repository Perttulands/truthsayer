package precedent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/perttulands/truthsayer/internal/finding"
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
	PatternHash   string    `json:"pattern_hash,omitempty"`
	Decision      Decision  `json:"decision"`
	Rationale     string    `json:"rationale"`
	Confidence    float64   `json:"confidence,omitempty"`
	SeenCount     int       `json:"seen_count,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	LastSeen      time.Time `json:"last_seen,omitempty"`
}

// Validate checks required fields and supported decision values.
func (p Precedent) Validate() error {
	if strings.TrimSpace(p.RuleID) == "" {
		return errors.New("precedent: rule_id is required")
	}
	if strings.TrimSpace(p.ViolationHash) == "" {
		return errors.New("precedent: violation_hash is required")
	}
	if p.Confidence < 0 || p.Confidence > 1 {
		return fmt.Errorf("precedent: confidence out of range [0,1]: %f", p.Confidence)
	}
	if p.SeenCount < 0 {
		return errors.New("precedent: seen_count cannot be negative")
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
		return fmt.Errorf("precedent: load before add: %w", err)
	}
	precedents = append(precedents, p)
	return s.Save(precedents)
}

// Query returns the most recent matching precedent by rule_id + violation_hash.
func (s *Store) Query(ruleID, violationHash string) (Precedent, bool, error) {
	precedents, err := s.Load()
	if err != nil {
		return Precedent{}, false, fmt.Errorf("precedent: load before query: %w", err)
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

// QueryByPattern returns the most recent precedent matching rule_id + pattern_hash.
func QueryByPattern(precedents []Precedent, ruleID, patternHash string) (Precedent, bool) {
	for i := len(precedents) - 1; i >= 0; i-- {
		p := precedents[i]
		if p.RuleID == ruleID && p.PatternHash == patternHash {
			return p, true
		}
	}
	return Precedent{}, false
}

// MatchOptions controls precedent lookup behavior.
type MatchOptions struct {
	MinConfidence float64
	Limit         int
}

func normalizeMatchOptions(opts MatchOptions) MatchOptions {
	if opts.MinConfidence < 0 {
		opts.MinConfidence = 0
	}
	if opts.MinConfidence > 1 {
		opts.MinConfidence = 1
	}
	return opts
}

func precedenceConfidence(p Precedent) float64 {
	if p.Confidence <= 0 {
		return 0.5
	}
	return p.Confidence
}

// Match returns precedents matching a finding by rule_id + pattern_hash.
// Results are sorted by confidence, then seen_count, then recency.
func Match(precedents []Precedent, f finding.Finding, opts MatchOptions) []Precedent {
	if len(precedents) == 0 {
		return nil
	}
	opts = normalizeMatchOptions(opts)
	patternHash := HashFindingPattern(f)
	if patternHash == "" {
		return nil
	}

	matches := make([]Precedent, 0)
	for _, p := range precedents {
		if p.RuleID != f.Rule {
			continue
		}
		if strings.TrimSpace(p.PatternHash) != patternHash {
			continue
		}
		if precedenceConfidence(p) < opts.MinConfidence {
			continue
		}
		matches = append(matches, p)
	}
	if len(matches) == 0 {
		return nil
	}

	sort.Slice(matches, func(i, j int) bool {
		a, b := matches[i], matches[j]
		if precedenceConfidence(a) != precedenceConfidence(b) {
			return precedenceConfidence(a) > precedenceConfidence(b)
		}
		if a.SeenCount != b.SeenCount {
			return a.SeenCount > b.SeenCount
		}
		return a.CreatedAt.After(b.CreatedAt)
	})

	if opts.Limit > 0 && len(matches) > opts.Limit {
		return matches[:opts.Limit]
	}
	return matches
}

// Match returns precedents matching a finding by rule_id + pattern_hash.
func (s *Store) Match(f finding.Finding, opts MatchOptions) ([]Precedent, error) {
	precedents, err := s.Load()
	if err != nil {
		return nil, fmt.Errorf("precedent: load before match: %w", err)
	}
	return Match(precedents, f, opts), nil
}

const (
	initialConfidence = 0.6
	confidenceStep    = 0.1
	overrideDecay     = 0.5
	minConfidence     = 0.2
)

// AddOrUpdateJudgment appends a precedent with confidence updates applied from prior pattern history.
// Same decision increments confidence/seen_count; overridden decision decays confidence and resets seen_count.
func (s *Store) AddOrUpdateJudgment(p Precedent) (Precedent, error) {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	if p.LastSeen.IsZero() {
		p.LastSeen = p.CreatedAt
	}

	precedents, err := s.Load()
	if err != nil {
		return Precedent{}, fmt.Errorf("precedent: load before upsert: %w", err)
	}

	if prev, ok := QueryByPattern(precedents, p.RuleID, p.PatternHash); ok {
		prevConfidence := precedenceConfidence(prev)
		prevSeen := prev.SeenCount
		if prevSeen <= 0 {
			prevSeen = 1
		}
		if prev.Decision == p.Decision {
			p.Confidence = prevConfidence + confidenceStep
			if p.Confidence > 1 {
				p.Confidence = 1
			}
			p.SeenCount = prevSeen + 1
		} else {
			p.Confidence = prevConfidence * overrideDecay
			if p.Confidence < minConfidence {
				p.Confidence = minConfidence
			}
			p.SeenCount = 1
		}
	} else {
		if p.Confidence <= 0 {
			p.Confidence = initialConfidence
		}
		if p.SeenCount <= 0 {
			p.SeenCount = 1
		}
	}

	precedents = append(precedents, p)
	if err := s.Save(precedents); err != nil {
		return Precedent{}, fmt.Errorf("save precedents: %w", err)
	}
	return p, nil
}
