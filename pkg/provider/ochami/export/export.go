package export

import (
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

func Export(existing devicetypes.Inventory) error {
	// Common patterns:
	//   - Compare local inventory with external system
	//   - Create/update/delete resources in external system
	//   - Report what was created, updated, skipped, or errored

	fmt.Println(existing)
	return nil
}
