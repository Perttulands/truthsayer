package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/perttulands/truthsayer/internal/beads"
	"github.com/perttulands/truthsayer/internal/finding"
)

type problemBeadCreator interface {
	CreateProblemBead(rule string, file string, count int) (string, error)
}

var newProblemBeadCreator = func() problemBeadCreator {
	return beads.NewBeadCreator()
}

type findingGroup struct {
	rule  string
	file  string
	count int
}

type createdBead struct {
	id    string
	rule  string
	file  string
	count int
}

func createErrorBeads(findings []finding.Finding, threshold int, creator problemBeadCreator) ([]createdBead, error) {
	groups := groupErrorFindings(findings)
	created := make([]createdBead, 0, len(groups))

	for _, g := range groups {
		if g.count <= threshold {
			continue
		}
		id, err := creator.CreateProblemBead(g.rule, g.file, g.count)
		if err != nil {
			return created, fmt.Errorf("create bead for %s in %s: %w", g.rule, g.file, err)
		}
		created = append(created, createdBead{
			id:    id,
			rule:  g.rule,
			file:  g.file,
			count: g.count,
		})
	}

	return created, nil
}

func groupErrorFindings(findings []finding.Finding) []findingGroup {
	m := make(map[string]*findingGroup)
	order := make([]string, 0)

	for _, f := range findings {
		if f.Severity != finding.SeverityError {
			continue
		}

		key := f.Rule + "\x00" + f.File
		if _, ok := m[key]; !ok {
			m[key] = &findingGroup{rule: f.Rule, file: f.File}
			order = append(order, key)
		}
		m[key].count++
	}

	groups := make([]findingGroup, 0, len(m))
	for _, key := range order {
		groups = append(groups, *m[key])
	}

	sort.Slice(groups, func(i, j int) bool {
		if groups[i].rule != groups[j].rule {
			return groups[i].rule < groups[j].rule
		}
		return groups[i].file < groups[j].file
	})

	return groups
}

func printBeadSummary(w io.Writer, created []createdBead) {
	fmt.Fprintf(w, "Beads created: %d\n", len(created))
	for _, b := range created {
		fmt.Fprintf(w, "  %s\n", b.id)
	}
}
