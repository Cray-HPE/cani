package example

import (
	"context"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/spf13/cobra"
)

// Import syncs the local CANI inventory from an external system.
// This is the "Extract" step in ETL.
func (p *Example) Import(ctx context.Context, cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return import_.Import(cmd, args, inventory)
}
