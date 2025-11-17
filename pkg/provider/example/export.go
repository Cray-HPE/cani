package example

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/export"
	"github.com/spf13/cobra"
)

// Export syncs the local CANI inventory to an external system.
// This is the "Load" step in ETL.
func (p *Example) Export(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	// Common patterns:
	//   - Compare local inventory with external system
	//   - Create/update/delete resources in external system
	//   - Report what was created, updated, skipped, or errored
	return export.Export(*inventory)
}
