package redfish

import (
	"context"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/redfish/export"
)

// Export syncs the local CANI inventory to an external system.
// This is the "Load" step in ETL.
func (p *Redfish) Export(ctx context.Context, cmd *cli.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Common patterns:
	//   - Compare local inventory with external system
	//   - Create/update/delete resources in external system
	//   - Report what was created, updated, skipped, or errored
	return export.Export(*inventory)
}
