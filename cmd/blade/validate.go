package blade

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/spf13/cobra"
)

// validHardware checks that the hardware type is valid by comparing it against the list of hardware types
func validHardware(cmd *cobra.Command, args []string) error {
	library, err := hardwaretypes.NewEmbeddedLibrary()
	if err != nil {
		return err
	}

	// Get the list of hardware types that are blades
	deviceTypes := library.GetDeviceTypesByHardwareType(hardwaretypes.HardwareTypeNodeBlade)
	if cmd.Flags().Changed("list-supported-types") {
		for _, hw := range deviceTypes {
			cmd.Printf("- %s\n", hw.Slug)
		}
		os.Exit(0)
	}

	if len(args) == 0 {
		bladeTypes := []string{}
		for _, hw := range deviceTypes {
			bladeTypes = append(bladeTypes, hw.Slug)
		}
		return fmt.Errorf("No hardware type provided: Choose from: %s", strings.Join(bladeTypes, "\", \""))
	}

	// Check that each arg is a valid blade type
	for _, arg := range args {
		matchFound := false
		for _, device := range deviceTypes {
			if arg == device.Slug {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return errors.New("Invalid hardware type: " + arg)
		}
	}

	return nil
}
