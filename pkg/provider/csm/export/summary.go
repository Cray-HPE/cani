package export

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// reconcileStats tracks the outcome of a reconcile operation.
type reconcileStats struct {
	PutCount    int
	DeleteCount int
	SkipCount   int
	NewCount    int // hardware items added by user (staged)
}

// capitalizeType returns a title-cased version of the device type
// string (e.g. "node" → "Node", "blade" → "Blade").
func capitalizeType(t string) string {
	if t == "" {
		return t
	}
	return strings.ToUpper(t[:1]) + t[1:]
}

// printSummary writes the export summary to w. The format matches the
// expected shellspec output:
//
//	Summary:
//	--------
//	ID  TYPE  STATUS
//	<list of new items>
//	N new hardware item(s) are in the inventory
func printSummary(w io.Writer, inventory devicetypes.Inventory, stats reconcileStats) {
	fmt.Fprintln(w, "Summary:")
	fmt.Fprintln(w, "--------")
	fmt.Fprintln(w, "ID  TYPE  STATUS")

	// Collect new (staged) hardware that was added by the user.
	type entry struct {
		id     string
		hwtype string
		status string
	}
	var newItems []entry
	for _, dev := range inventory.Devices {
		if dev == nil {
			continue
		}
		if strings.EqualFold(dev.Status, "staged") {
			newItems = append(newItems, entry{
				id:     dev.ID.String(),
				hwtype: capitalizeType(string(dev.GetType())),
				status: dev.Status,
			})
		}
	}

	sort.Slice(newItems, func(i, j int) bool {
		return newItems[i].id < newItems[j].id
	})
	for _, item := range newItems {
		// Pad type to 32 chars and wrap status in parentheses.
		fmt.Fprintf(w, "%-32s(%s)\n", item.hwtype, item.status)
	}
	// Blank separator line between the table and the count.
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%d new hardware item(s) are in the inventory\n", len(newItems))
}
