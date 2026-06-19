package csm

import (
	"context"
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

// Import reads raw SLS and SMD data from JSON files and stores it on the
// provider singleton. This is the "Extract" step in ETL.
// When stdin is piped and no API/file flags are set, it performs a CSV import.
func (p *Csm) Import(ctx context.Context, cmd *cli.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	useAPI, err := import_.ShouldUseAPI(cmd)
	if err != nil {
		return err
	}

	hasFile, _ := cmd.Flags().GetString("sls-file")

	// If stdin is piped and we're not doing API or file import, use CSV import.
	if !useAPI && hasFile == "" && import_.IsStdinPiped() {
		modified, total, csvErr := import_.ImportCSV(inventory)
		if csvErr != nil {
			return fmt.Errorf("CSV import failed: %w", csvErr)
		}
		log.Printf("Success: Wrote %d records of a total %d records from the CSV data", modified, total)
		return nil
	}

	if err := import_.Import(cmd, args); err != nil {
		return err
	}

	// Validate the imported SLS data.
	if p.RawSls != nil {
		if vErr := import_.ValidateSlsDumpstate(p.RawSls); vErr != nil {
			ignoreValidation, _ := cmd.Flags().GetBool("ignore-validation")
			if ignoreValidation {
				log.Printf("WRN Ignoring these failures: %v", vErr)
			} else {
				return fmt.Errorf("%w\nExternal inventory is unstable", vErr)
			}
		}
	}

	return nil
}
