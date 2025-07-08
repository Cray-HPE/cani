package ngsm

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ngsm/show"
	"github.com/spf13/cobra"
)

// Show is called when the user runs `cani list`
func (p *Ngsm) Show(cmd *cobra.Command, args []string, devices []*devicetypes.CaniDeviceType) (err error) {
	return show.Show(cmd, args, devices)
}
