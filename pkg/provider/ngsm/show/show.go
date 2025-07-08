package show

import (
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

func Show(cmd *cobra.Command, args []string, devices []*devicetypes.CaniDeviceType) (err error) {
	if cmd.PersistentFlags().Changed("format") {
		format, _ := cmd.PersistentFlags().GetString("format")
		if format != "table" {
			return fmt.Errorf("unsupported format '%s'. Supported format is 'table'", format)
		}
	}
	for _, device := range devices {
		fmt.Printf("Device ID: %s, Type: %s, Name: %s\n", device.ID, device.Type, device.Name)
	}
	return nil
}
