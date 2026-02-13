package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/perttulands/truthsayer/internal/rules"
)

func runRules(args []string) int {
	reg := rules.DefaultRegistry()

	allRules := reg.AllRules()
	sort.Slice(allRules, func(i, j int) bool {
		return allRules[i].ID < allRules[j].ID
	})

	if len(allRules) == 0 {
		fmt.Fprintln(os.Stderr, "No rules registered.")
		return 0
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSEVERITY\tFILES\tDESCRIPTION")
	fmt.Fprintln(w, "--\t--------\t-----\t-----------")
	for _, r := range allRules {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			r.ID,
			strings.ToUpper(string(r.Severity)),
			strings.Join(r.FileTypes, ","),
			r.Description,
		)
	}
	w.Flush()

	fmt.Fprintf(os.Stdout, "\n%d rules available\n", len(allRules))
	return 0
}
