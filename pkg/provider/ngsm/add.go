package ngsm

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ngsm/add"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func (p *Ngsm) Add(cmd *cobra.Command, args []string, deviceType devicetypes.DeviceType) (devicesToAdd map[uuid.UUID]*devicetypes.CaniDeviceType, err error) {
	return add.Add(cmd, args, deviceType)
}
