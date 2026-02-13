package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/perttulands/truthsayer/internal/config"
	"github.com/perttulands/truthsayer/internal/rules"
)

func runRules(args []string) int {
	configPath, remaining := parseConfigFlag(args)

	// Check for --enabled flag
	enabled := false
	for _, arg := range remaining {
		if arg == "--enabled" {
			enabled = true
		}
	}

	reg := rules.DefaultRegistry()

	// Apply config if --enabled is used (to respect disabled rules)
	if enabled {
		cfg, err := config.Load(".", configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
			return 2
		}
		for _, id := range cfg.Rules.Disable {
			reg.Disable(id)
		}
		// Apply severity overrides so listed severities reflect config
		for id, sev := range cfg.Rules.Severity {
			reg.SetSeverity(id, sev)
		}
	}

	var ruleList []rules.Rule
	if enabled {
		ruleList = reg.EnabledRules()
	} else {
		ruleList = reg.AllRules()
	}

	sort.Slice(ruleList, func(i, j int) bool {
		return ruleList[i].ID < ruleList[j].ID
	})

	if len(ruleList) == 0 {
		fmt.Fprintln(os.Stderr, "No rules registered.")
		return 0
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSEVERITY\tFILES\tDESCRIPTION")
	fmt.Fprintln(w, "--\t--------\t-----\t-----------")
	for _, r := range ruleList {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			r.ID,
			strings.ToUpper(string(r.Severity)),
			strings.Join(r.FileTypes, ","),
			r.Description,
		)
	}
	w.Flush()

	label := "available"
	if enabled {
		label = "enabled"
	}
	fmt.Fprintf(os.Stdout, "\n%d rules %s\n", len(ruleList), label)
	return 0
}
