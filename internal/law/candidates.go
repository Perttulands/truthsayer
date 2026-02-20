package law

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/perttulands/truthsayer/internal/precedent"
)

// DefaultCandidatesPath is where consistent-rule candidates are persisted.
const DefaultCandidatesPath = ".truthsayer-law-candidates.json"

// Candidate represents a consistent ruling pattern suitable for law updates.
type Candidate struct {
	RuleID      string             `json:"rule_id"`
	PatternHash string             `json:"pattern_hash"`
	Decision    precedent.Decision `json:"decision"`
	Count       int                `json:"count"`
	Threshold   int                `json:"threshold"`
	FirstSeen   time.Time          `json:"first_seen"`
	LastSeen    time.Time          `json:"last_seen"`
}

// Validate checks required candidate fields.
func (c Candidate) Validate() error {
	if strings.TrimSpace(c.RuleID) == "" {
		return errors.New("law: rule_id is required")
	}
	if strings.TrimSpace(c.PatternHash) == "" {
		return errors.New("law: pattern_hash is required")
	}
	if c.Decision != precedent.DecisionAllow && c.Decision != precedent.DecisionDeny {
		return fmt.Errorf("law: invalid decision %q", c.Decision)
	}
	if c.Count <= 0 {
		return errors.New("law: count must be > 0")
	}
	if c.Threshold <= 0 {
		return errors.New("law: threshold must be > 0")
	}
	if c.FirstSeen.IsZero() || c.LastSeen.IsZero() {
		return errors.New("law: first_seen and last_seen are required")
	}
	return nil
}

type candidateKey struct {
	ruleID      string
	patternHash string
	decision    precedent.Decision
}

type candidateAgg struct {
	count     int
	firstSeen time.Time
	lastSeen  time.Time
}

// DetectCandidates finds consistent ruling patterns at or above threshold.
func DetectCandidates(precedents []precedent.Precedent, threshold int) []Candidate {
	if threshold <= 0 {
		threshold = 10
	}
	aggs := make(map[candidateKey]candidateAgg)
	for _, p := range precedents {
		if strings.TrimSpace(p.RuleID) == "" || strings.TrimSpace(p.PatternHash) == "" {
			continue
		}
		if p.Decision != precedent.DecisionAllow && p.Decision != precedent.DecisionDeny {
			continue
		}
		at := p.CreatedAt
		if at.IsZero() {
			continue
		}
		key := candidateKey{ruleID: p.RuleID, patternHash: p.PatternHash, decision: p.Decision}
		agg := aggs[key]
		agg.count++
		if agg.firstSeen.IsZero() || at.Before(agg.firstSeen) {
			agg.firstSeen = at
		}
		if agg.lastSeen.IsZero() || at.After(agg.lastSeen) {
			agg.lastSeen = at
		}
		aggs[key] = agg
	}

	candidates := make([]Candidate, 0, len(aggs))
	for key, agg := range aggs {
		if agg.count < threshold {
			continue
		}
		candidates = append(candidates, Candidate{
			RuleID:      key.ruleID,
			PatternHash: key.patternHash,
			Decision:    key.decision,
			Count:       agg.count,
			Threshold:   threshold,
			FirstSeen:   agg.firstSeen,
			LastSeen:    agg.lastSeen,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		if a.RuleID != b.RuleID {
			return a.RuleID < b.RuleID
		}
		if a.Decision != b.Decision {
			return a.Decision < b.Decision
		}
		return a.PatternHash < b.PatternHash
	})
	return candidates
}

// CandidateStore persists law candidates.
type CandidateStore struct {
	path string
}

// NewCandidateStore creates a candidate store for path.
func NewCandidateStore(path string) *CandidateStore {
	if strings.TrimSpace(path) == "" {
		path = DefaultCandidatesPath
	}
	return &CandidateStore{path: path}
}

// Load reads candidates from disk.
func (s *CandidateStore) Load() ([]Candidate, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Candidate{}, nil
		}
		return nil, fmt.Errorf("law: read %s: %w", s.path, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return []Candidate{}, nil
	}
	var candidates []Candidate
	if err := json.Unmarshal(data, &candidates); err != nil {
		return nil, fmt.Errorf("law: decode %s: %w", s.path, err)
	}
	for i, c := range candidates {
		if err := c.Validate(); err != nil {
			return nil, fmt.Errorf("law: invalid candidate %d: %w", i, err)
		}
	}
	return candidates, nil
}

// Save writes candidates to disk.
func (s *CandidateStore) Save(candidates []Candidate) error {
	for i, c := range candidates {
		if err := c.Validate(); err != nil {
			return fmt.Errorf("law: candidate %d: %w", i, err)
		}
	}
	out, err := json.MarshalIndent(candidates, "", "  ")
	if err != nil {
		return fmt.Errorf("law: encode %s: %w", s.path, err)
	}
	out = append(out, '\n')
	dir := filepath.Dir(s.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("law: create dir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(s.path, out, 0o644); err != nil {
		return fmt.Errorf("law: write %s: %w", s.path, err)
	}
	return nil
}

