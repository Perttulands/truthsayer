package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/perttulands/truthsayer/internal/config"
	"github.com/perttulands/truthsayer/internal/engine"
	"github.com/perttulands/truthsayer/internal/finding"
	"github.com/perttulands/truthsayer/internal/rules"
	"github.com/perttulands/truthsayer/internal/senate"
)

// validSeverities is the set of valid severity values for config overrides.
var validSeverities = map[string]bool{
	string(finding.SeverityError):   true,
	string(finding.SeverityWarning): true,
	string(finding.SeverityInfo):    true,
}

// buildEngine creates a configured engine from the given config path.
// scanDir is used to find .truthsayer.toml when configPath is empty.
func buildEngine(scanDir, configPath string) (*engine.Engine, error) {
	cfg, err := config.Load(scanDir, configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	reg := rules.DefaultRegistry()

	for _, id := range cfg.Rules.Disable {
		reg.Disable(id)
	}

	for id, sev := range cfg.Rules.Severity {
		if !validSeverities[sev] {
			return nil, fmt.Errorf("config: invalid severity %q for rule %q (valid: error, warning, info)", sev, id)
		}
		if !reg.SetSeverity(id, sev) {
			fmt.Fprintf(os.Stderr, "warning: unknown rule %q in severity config\n", id)
		}
	}

	// Apply approved Senate amendments from repo-local amendment store.
	applied, err := senate.NewAmendmentStore(filepath.Join(scanDir, senate.DefaultAmendmentsPath)).Load()
	if err != nil {
		return nil, fmt.Errorf("load senate amendments: %w", err)
	}
	for _, rec := range applied {
		am := rec.Amendment
		switch am.Action {
		case senate.ActionSetSeverity:
			if !validSeverities[am.Severity] {
				return nil, fmt.Errorf("senate amendment: invalid severity %q for rule %q", am.Severity, am.RuleID)
			}
			if !reg.SetSeverity(am.RuleID, am.Severity) {
				fmt.Fprintf(os.Stderr, "warning: unknown senate amendment rule %q for set_severity\n", am.RuleID)
			}
		case senate.ActionDisableRule:
			reg.Disable(am.RuleID)
		case senate.ActionEnableRule:
			reg.Enable(am.RuleID)
		case senate.ActionAddException:
			// Rule-specific exceptions can be implemented incrementally per rule.
			fmt.Fprintf(os.Stderr, "info: senate add_exception recorded for %s (%s)\n", am.RuleID, am.Exception)
		}
	}

	eng := engine.New(reg)

	eng.SetLanguages(&cfg.Scan.Languages)

	// Apply scan exclusions from config
	if len(cfg.Scan.ExcludeDirs) > 0 {
		merged := make(map[string]bool)
		for k, v := range engine.DefaultExcludeDirs() {
			merged[k] = v
		}
		for _, d := range cfg.Scan.ExcludeDirs {
			merged[d] = true
		}
		eng.SetExcludeDirs(merged)
	}

	if len(cfg.Scan.ExcludePatterns) > 0 {
		eng.SetExcludePatterns(cfg.Scan.ExcludePatterns)
	}

	return eng, nil
}

// parseConfigFlag extracts --config flag from args, returning the config path and remaining args.
func parseConfigFlag(args []string) (configPath string, remaining []string) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--config" && i+1 < len(args) {
			configPath = args[i+1]
			i++
		} else {
			remaining = append(remaining, args[i])
		}
	}
	return configPath, remaining
}
