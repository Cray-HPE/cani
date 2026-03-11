package ochami

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ochami/export"
	"github.com/spf13/cobra"
)

// Export syncs the local CANI inventory to an external system.
// This is the "Load" step in ETL.
func (p *Ochami) Export(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	return export.Export(*inventory, cmd.OutOrStdout())
}
