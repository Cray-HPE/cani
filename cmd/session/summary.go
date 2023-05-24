package session

import (
	"fmt"
	"os"
	"text/tabwriter"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/plugin"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// SessionSummaryCmd represents the session stop command
var SessionSummaryCmd = &cobra.Command{
	Use:          "summary",
	Short:        "Show the summary of a stopped session",
	Long:         `Show the summary of a stopped session`,
	SilenceUsage: true, // Errors are more important than the usage
	RunE:         showSummary,
}

func showSummary(cmd *cobra.Command, args []string) error {
	// Instanstiate the domain
	d, err := plugin.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	// Get the entire inventory
	inv, err := d.List()
	if err != nil {
		return err
	}

	// print a header
	fmt.Println("Summary:")
	fmt.Println("--------")
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tTYPE\tSTATUS")

	staged := make(map[uuid.UUID]inventory.Hardware, 0)

	// Create the colors you want to use
	black := color.New(color.FgBlack).FprintfFunc()
	red := color.New(color.FgRed).FprintfFunc()
	green := color.New(color.FgGreen).FprintfFunc()
	yellow := color.New(color.FgYellow).FprintfFunc()
	blue := color.New(color.FgBlue).FprintfFunc()
	// magenta := color.New(color.FgMagenta).FprintfFunc()
	// cyan := color.New(color.FgCyan).FprintfFunc()

	const format = "%s\t%s\t(%s)\n"
	// for each new hardware, print some details
	for i, hw := range inv.Hardware {
		// Only show staged hardware
		// TODO: Better logic as staged hardware could have been added in a different session
		if hw.Status == inventory.HardwareStatusStaged {
			switch hw.Type {
			case hardwaretypes.HardwareTypeCabinet:
				red(tw, format, i.String(), hw.Type, hw.Status)
			case hardwaretypes.HardwareTypeChassis:
				yellow(tw, format, i.String(), hw.Type, hw.Status)
			case hardwaretypes.HardwareTypeNodeBlade:
				green(tw, format, i.String(), hw.Type, hw.Status)
			case hardwaretypes.HardwareTypeNode:
				blue(tw, format, i.String(), hw.Type, hw.Status)
			default:
				black(tw, format, i.String(), hw.Type, hw.Status)
			}
			staged[i] = hw
		}
	}

	tw.Flush()

	// print the next steps
	fmt.Printf("\n%d hardware item(s) are staged:\n", len(staged))
	fmt.Println("\nNext steps:")
	fmt.Println("-----------")
	fmt.Println("1. Power on the new nodes using your existing methods.")
	fmt.Println("2. Check the status of the nodes using your existing methods.")
	fmt.Println("3. Proceed with system configuration as needed.")

	return nil
}
