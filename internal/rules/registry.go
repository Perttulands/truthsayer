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
	jsAST             []JSASTChecker
	pyAST             []PyASTChecker
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

// RegisterJSAST adds a JS/TS AST-based checker.
func (r *Registry) RegisterJSAST(c JSASTChecker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jsAST = append(r.jsAST, c)
}

// RegisterPyAST adds a Python AST-based checker.
func (r *Registry) RegisterPyAST(c PyASTChecker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pyAST = append(r.pyAST, c)
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
	for _, c := range r.jsAST {
		if c.Meta().ID == id {
			r.severityOverrides[id] = sev
			return true
		}
	}
	for _, c := range r.pyAST {
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

// JSASTCheckers returns all enabled JS/TS AST checkers.
func (r *Registry) JSASTCheckers() []JSASTChecker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []JSASTChecker
	for _, c := range r.jsAST {
		if !r.disabled[c.Meta().ID] {
			out = append(out, c)
		}
	}
	return out
}

// PyASTCheckers returns all enabled Python AST checkers.
func (r *Registry) PyASTCheckers() []PyASTChecker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []PyASTChecker
	for _, c := range r.pyAST {
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
	for _, c := range r.jsAST {
		out = append(out, c.Meta())
	}
	for _, c := range r.pyAST {
		out = append(out, c.Meta())
	}
	return out
}

// EnabledRules returns metadata for all enabled (non-disabled) rules,
// with severity overrides applied.
func (r *Registry) EnabledRules() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Rule
	for _, c := range r.ast {
		m := c.Meta()
		if !r.disabled[m.ID] {
			if sev, ok := r.severityOverrides[m.ID]; ok {
				m.Severity = sev
			}
			out = append(out, m)
		}
	}
	for _, c := range r.regex {
		m := c.Meta()
		if !r.disabled[m.ID] {
			if sev, ok := r.severityOverrides[m.ID]; ok {
				m.Severity = sev
			}
			out = append(out, m)
		}
	}
	for _, c := range r.jsAST {
		m := c.Meta()
		if !r.disabled[m.ID] {
			if sev, ok := r.severityOverrides[m.ID]; ok {
				m.Severity = sev
			}
			out = append(out, m)
		}
	}
	for _, c := range r.pyAST {
		m := c.Meta()
		if !r.disabled[m.ID] {
			if sev, ok := r.severityOverrides[m.ID]; ok {
				m.Severity = sev
			}
			out = append(out, m)
		}
	}
	return out
}

// DefaultRegistry creates a registry with all built-in rules registered.
func DefaultRegistry() *Registry {
	reg := NewRegistry()

	// AST rules (Go)
	reg.RegisterAST(&EmptyErrorCheck{})
	reg.RegisterAST(&BareReturnOnError{})
	reg.RegisterAST(&IgnoredError{})
	reg.RegisterAST(&UnwrappedError{})
	reg.RegisterAST(&HTTP200OnError{})
	reg.RegisterAST(&NilOnError{})
	reg.RegisterAST(&GenericError{})
	reg.RegisterAST(&NoTimeout{})
	reg.RegisterAST(&MagicNumber{})
	reg.RegisterAST(&UnvalidatedEnvGo{})
	reg.RegisterAST(&LongFunctionNoLog{})
	reg.RegisterAST(&ErrorPathNoLog{})
	reg.RegisterAST(&MockImportNonTest{})
	reg.RegisterAST(&NoRequestID{})
	reg.RegisterAST(&DebugGuard{})
	reg.RegisterAST(&UncheckedTypeAssert{})
	reg.RegisterAST(&DeferInLoop{})
	reg.RegisterAST(&GoroutineNoContext{})
	reg.RegisterAST(&SwallowedError{})
	reg.RegisterAST(&ContextTodo{})
	reg.RegisterAST(&ErrorStringCompare{})

	// Regex rules (bash/config/multi-lang)
	reg.RegisterRegex(&MissingPipefail{})
	reg.RegisterRegex(&HiddenFailureBash{})
	reg.RegisterRegex(&HardcodedPath{})
	reg.RegisterRegex(&UnvalidatedEnvBash{})
	reg.RegisterRegex(&SecretInConfig{})
	reg.RegisterRegex(&NoErrTrap{})
	reg.RegisterRegex(&TestFixtureRef{})
	reg.RegisterRegex(&NoStderrCapture{})
	reg.RegisterRegex(&MissingGitignore{})
	reg.RegisterRegex(&TestLeakedServer{})
	reg.RegisterRegex(&TestLeakedSSE{})
	reg.RegisterRegex(&TestMissingCleanup{})

	// JS/TS AST rules
	reg.RegisterJSAST(&JSEmptyCatch{})
	reg.RegisterJSAST(&JSCatchReturnNull{})
	reg.RegisterJSAST(&JSFloatingPromise{})
	reg.RegisterJSAST(&JSCallbackErrIgnored{})
	reg.RegisterJSAST(&JSOptionalChainSilence{})
	reg.RegisterJSAST(&JSRethrowNoWrap{})
	reg.RegisterJSAST(&JSGenericErrorMessage{})
	reg.RegisterJSAST(&JSPromiseReject{})
	reg.RegisterJSAST(&JSConsoleErrorNoThrow{})
	reg.RegisterJSAST(&JSHTTP200OnError{})
	reg.RegisterJSAST(&JSNoErrorHandlerExpress{})
	reg.RegisterJSAST(&JSTestImportInSrc{})
	reg.RegisterJSAST(&JSEnvTestCheck{})
	reg.RegisterJSAST(&JSMissingCorrelationID{})
	reg.RegisterJSAST(&JSNoTimeoutFetch{})
	reg.RegisterJSAST(&JSAnyAssertion{})
	reg.RegisterJSAST(&JSNonNullAssertion{})
	reg.RegisterJSAST(&JSEvalUsage{})
	reg.RegisterJSAST(&JSNoAfterallCleanup{})
	reg.RegisterJSAST(&JSTestOnlyImport{})

	return reg
}
