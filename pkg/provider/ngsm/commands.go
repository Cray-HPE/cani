package ngsm

import (
	"github.com/Cray-HPE/cani/pkg/provider/ngsm/commands"
	"github.com/spf13/cobra"
)

// NewProviderCmd creates commands for the NGSM provider
func (p *Ngsm) NewProviderCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	return commands.NewProviderCmd(caniCmd)
}
