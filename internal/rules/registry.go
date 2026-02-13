package rules

import (
	"sync"

	"github.com/perttulands/truthsayer/internal/finding"
)

// Registry holds all registered detection rules.
type Registry struct {
	mu                sync.RWMutex
	ast               []ASTChecker
	regex             []RegexChecker
	disabled          map[string]bool
	severityOverrides map[string]finding.Severity
}

// NewRegistry creates an empty rule registry.
func NewRegistry() *Registry {
	return &Registry{
		disabled:          make(map[string]bool),
		severityOverrides: make(map[string]finding.Severity),
	}
}

// RegisterAST adds an AST-based checker.
func (r *Registry) RegisterAST(c ASTChecker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ast = append(r.ast, c)
}

// RegisterRegex adds a regex-based checker.
func (r *Registry) RegisterRegex(c RegexChecker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.regex = append(r.regex, c)
}

// Disable marks a rule ID as disabled.
func (r *Registry) Disable(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.disabled[id] = true
}

// SetSeverity overrides the severity for a rule by ID.
// Returns false if the rule ID is not registered.
func (r *Registry) SetSeverity(id string, severity string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	sev := finding.Severity(severity)
	// Verify rule exists
	for _, c := range r.ast {
		if c.Meta().ID == id {
			r.severityOverrides[id] = sev
			return true
		}
	}
	for _, c := range r.regex {
		if c.Meta().ID == id {
			r.severityOverrides[id] = sev
			return true
		}
	}
	return false
}

// ApplyOverrides rewrites severity on findings using configured overrides.
func (r *Registry) ApplyOverrides(findings []finding.Finding) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.severityOverrides) == 0 {
		return
	}
	for i := range findings {
		if sev, ok := r.severityOverrides[findings[i].Rule]; ok {
			findings[i].Severity = sev
		}
	}
}

// ASTCheckers returns all enabled AST checkers.
func (r *Registry) ASTCheckers() []ASTChecker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []ASTChecker
	for _, c := range r.ast {
		if !r.disabled[c.Meta().ID] {
			out = append(out, c)
		}
	}
	return out
}

// RegexCheckers returns all enabled regex checkers.
func (r *Registry) RegexCheckers() []RegexChecker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []RegexChecker
	for _, c := range r.regex {
		if !r.disabled[c.Meta().ID] {
			out = append(out, c)
		}
	}
	return out
}

// AllRules returns metadata for all registered rules.
func (r *Registry) AllRules() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Rule
	for _, c := range r.ast {
		out = append(out, c.Meta())
	}
	for _, c := range r.regex {
		out = append(out, c.Meta())
	}
	return out
}

// DefaultRegistry creates a registry with all built-in rules registered.
func DefaultRegistry() *Registry {
	reg := NewRegistry()
	reg.RegisterAST(&EmptyErrorCheck{})
	reg.RegisterRegex(&MissingPipefail{})
	return reg
}
