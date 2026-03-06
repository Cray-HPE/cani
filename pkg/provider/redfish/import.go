package redfish

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
	"github.com/spf13/cobra"
)

// Import syncs the local CANI inventory from an external system.
// This is the "Extract" step in ETL.
func (p *Redfish) Import(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	return import_.Import(cmd, args, inventory)
}
