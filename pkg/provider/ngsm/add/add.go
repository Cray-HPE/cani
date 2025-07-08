package add

import (
	"log"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func Add(cmd *cobra.Command, args []string, deviceType devicetypes.DeviceType) (devicesToAdd map[uuid.UUID]*devicetypes.CaniDeviceType, err error) {
	devicesToAdd = make(map[uuid.UUID]*devicetypes.CaniDeviceType, 0)
	qty, err := cmd.Flags().GetInt("qty")
	if err != nil {
		return nil, err
	}

	parent, err := cmd.Flags().GetString("parent")
	if err != nil {
		return nil, err
	}

	for q := 1; q <= qty; q++ {
		// create a new device of the requested type
		newDevice := deviceType.ToCaniDeviceType()

		// add parent if requested
		if cmd.Flags().Changed("parent") {
			u, err := uuid.Parse(parent)
			if err != nil {
				return nil, err
			}
			newDevice.Parent = u
		}

		// add the device to the map
		devicesToAdd[newDevice.ID] = newDevice
		log.Printf("%d/%d: %v", q, qty, newDevice.DeviceTypeSlug)
	}

	return devicesToAdd, nil
}
