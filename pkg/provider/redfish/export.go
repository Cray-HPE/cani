package redfish

import (
	"context"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/redfish/commands"
	"github.com/Cray-HPE/cani/pkg/provider/redfish/export"
)

// Export syncs the local CANI inventory to an external system.
// This is the "Load" step in ETL.  The --dry-run flag is honored consistently
// with the other providers' exporters: when set, no payload is written.
func (p *Redfish) Export(ctx context.Context, cmd *cli.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return export.Export(*inventory, commands.DryRunFlag)
}
