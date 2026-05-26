package example

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/spf13/cobra"
)

// Import syncs the local CANI inventory from an external system.
// This is the "Extract" step in ETL.
func (p *Example) Import(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	return import_.Import(cmd, args, inventory)
}
