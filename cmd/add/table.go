package add

import (
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// categoryRank defines the hierarchical display order for hardware categories.
// Lower numbers appear first. Unknown categories sort to the end.
var categoryRank = map[string]int{
	"rack":         0,
	"cabinet":      1,
	"chassis":      2,
	"blade":        3,
	"node":         4,
	"nodecard":     5,
	"switch":       6,
	"mgmt-switch":  7,
	"hsn-switch":   8,
	"cabinet-pdu":  9,
	"cdu":          10,
	"module":       11,
	"nic":          12,
	"gpu":          13,
	"cpu":          14,
	"memory":       15,
	"power-supply": 16,
	"cable":        17,
	"fru":          18,
}

const (
	maxNameLen = 35
	maxSlugLen = 45
	maxPNLen   = 16
)

// printTypeTable renders a sorted, column-aligned table of type entries
// grouped by category. Columns: NAME, SLUG, PART NUMBER, SOURCE.
func printTypeTable(cmd *cobra.Command, entries []devicetypes.TypeEntry) {
	if len(entries) == 0 {
		cmd.Println("No hardware types available.")
		return
	}

	// Group by category
	groups := make(map[string][]devicetypes.TypeEntry)
	for _, e := range entries {
		cat := e.Category
		if cat == "" {
			cat = "other"
		}
		groups[cat] = append(groups[cat], e)
	}

	// Sort categories hierarchically
	cats := make([]string, 0, len(groups))
	for c := range groups {
		cats = append(cats, c)
	}
	sort.Slice(cats, func(i, j int) bool {
		ri, oki := categoryRank[cats[i]]
		rj, okj := categoryRank[cats[j]]
		if !oki {
			ri = len(categoryRank)
		}
		if !okj {
			rj = len(categoryRank)
		}
		if ri != rj {
			return ri < rj
		}
		return cats[i] < cats[j]
	})

	header := formatRow("NAME", "SLUG", "PART NUMBER", "SOURCE")
	separator := formatRow(
		strings.Repeat("-", maxNameLen),
		strings.Repeat("-", maxSlugLen),
		strings.Repeat("-", maxPNLen),
		strings.Repeat("-", 10),
	)

	for _, cat := range cats {
		items := groups[cat]
		sort.Slice(items, func(i, j int) bool {
			return items[i].Slug < items[j].Slug
		})

		cmd.Printf("\n%s (%d):\n", cat, len(items))
		cmd.Println(header)
		cmd.Println(separator)
		for _, e := range items {
			cmd.Println(formatRow(
				truncate(e.Name, maxNameLen),
				truncate(e.Slug, maxSlugLen),
				truncate(e.PartNumber, maxPNLen),
				e.Source,
			))
		}
	}
	cmd.Println()
}

// formatRow pads each field to its column width.
func formatRow(name, slug, pn, source string) string {
	return pad(name, maxNameLen) + "  " +
		pad(slug, maxSlugLen) + "  " +
		pad(pn, maxPNLen) + "  " +
		source
}

// truncate shortens s to max chars, appending "…" if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}

// pad right-pads s with spaces to width n.
func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
