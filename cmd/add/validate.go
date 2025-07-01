package add

import (
	"fmt"
	"os"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// validDeviceType checks that the hardware type is valid by comparing it against the list of hardware types
func validDeviceType(cmd *cobra.Command, args []string) (err error) {
	var devices map[string]devicetypes.DeviceType
	var validDevices = []string{}
	cmd.SetOut(os.Stdout)
	switch cmd.Name() {

	case "rack":
		devices = devicetypes.Racks()

	case "blade":
		devices = devicetypes.Blades()

	}

	for _, hw := range devices {
		validDevices = append(validDevices, hw.Slug)
	}

	if cmd.Flags().Changed("list-supported-types") {
		return listSupportedTypes(cmd, args)
	}

	if len(args) == 0 {
		sort.Strings(validDevices)
		cmd.SetOut(os.Stderr)
		cmd.Println("No device type provided. Choose from the following:")
		for _, device := range devices {
			cmd.Printf("- %s: (%s)\n", device.Model, device.Slug)
		}
		return fmt.Errorf("no device type provided")
	}

	// Check that each arg is a valid cabinet type
	for _, arg := range args {
		matchFound := false
		for _, device := range validDevices {
			if arg == device {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return fmt.Errorf("invalid hardware type: " + arg)
		}
	}

	return nil
}

func listSupportedTypes(cmd *cobra.Command, args []string) error {
	var devices map[string]devicetypes.DeviceType
	var validDevices = []string{}
	switch cmd.Name() {

	case "rack":
		devices = devicetypes.Racks()

	case "blade":
		devices = devicetypes.Blades()

	}

	for _, device := range devices {
		validDevices = append(validDevices, device.Slug)
	}
	sort.Strings(validDevices)

	cmd.SetOut(os.Stdout)
	cmd.Println("Supported hardware types:")
	for _, device := range devices {
		cmd.Printf("- %s: (%s)\n", device.Model, device.Slug)
	}

	return nil
}
