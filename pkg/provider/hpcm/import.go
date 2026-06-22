package hpcm

import (
	"context"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/hpcm/import"
	"github.com/spf13/cobra"
)

// Import reads raw HPCM node data from a file or stdin and stores it on the
// provider. This is the "Extract" step in ETL — no transformation is done here.
func (p *Hpcm) Import(ctx context.Context, cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return import_.Import(cmd, args)
}
