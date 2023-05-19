package sw

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

	// Get the list of hardware types that are switchs
	mgmtSwitchTypes := library.GetDeviceTypesByHardwareType(hardwaretypes.HardwareTypeManagementSwitch)
	hsnSwitchTypes := library.GetDeviceTypesByHardwareType(hardwaretypes.HardwareTypeHighSpeedSwitch)

	if cmd.Flags().Changed("list-supported-types") {
		for _, hw := range mgmtSwitchTypes {
			cmd.Printf("- %s\n", hw.Slug)
		}
		for _, hw := range hsnSwitchTypes {
			cmd.Printf("- %s\n", hw.Slug)
		}
		os.Exit(0)
	}

	if len(args) == 0 {
		types := []string{}
		for _, hw := range mgmtSwitchTypes {
			types = append(types, hw.Slug)
		}
		for _, hw := range hsnSwitchTypes {
			types = append(types, hw.Slug)
		}
		return fmt.Errorf("No hardware type provided: Choose from: %s", strings.Join(types, "\", \""))
	}

	// Check that each arg is a valid switch type
	for _, arg := range args {
		matchFound := false
		for _, device := range mgmtSwitchTypes {
			if arg == device.Slug {
				matchFound = true
				break
			}
		}
		for _, device := range hsnSwitchTypes {
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
