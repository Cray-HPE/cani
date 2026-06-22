package ochami

import (
	"context"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ochami/export"
)

// Export syncs the local CANI inventory to an external system.
// This is the "Load" step in ETL.
func (p *Ochami) Export(ctx context.Context, cmd *cli.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return export.Export(*inventory, cmd.OutOrStdout())
}
