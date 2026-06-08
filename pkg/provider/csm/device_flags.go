package csm

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// RegisterDeviceUpdateFlags implements provider.DeviceUpdateFlagProvider.
// It adds CSM-specific flags to the generic "update device" command.
func (p *Csm) RegisterDeviceUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Int("nid", 0, "Node ID (CSM provider)")
	cmd.Flags().String("alias", "", "Node alias (CSM provider)")
}

// ApplyDeviceUpdateFlags implements provider.DeviceUpdateFlagProvider.
// It applies any changed CSM-specific flags to the device's metadata.
func (p *Csm) ApplyDeviceUpdateFlags(cmd *cobra.Command, device *devicetypes.CaniDeviceType) error {
	if cmd.Flags().Changed("nid") {
		nid, _ := cmd.Flags().GetInt("nid")
		device.SetProviderMeta(p.Slug(), "nid", nid)
	}
	if cmd.Flags().Changed("alias") {
		alias, _ := cmd.Flags().GetString("alias")
		device.SetProviderMeta(p.Slug(), "aliases", []string{alias})
	}
	return nil
}
