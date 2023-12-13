package hpengi

import (
	"github.com/Cray-HPE/cani/pkg/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func Init() {
	log.Trace().Msgf("%+v", "github.com/Cray-HPE/cani/internal/provider/hpengi.init")
}

func NewSessionInitCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	cmd.Flags().StringP("cid", "c", "", "Path to a Smart Customer Intent Document")
	cmd.Flags().StringP("cm-config", "C", "", "Path to a HPCM config file/cluster manager file/cluster definition file")
	cmd.Flags().StringP("paddle", "P", "", "Path to a Paddle/CCJ/machine-readable SHCD file")
	cmd.Flags().StringP("sls-dumpstate", "s", "", "Path to a SLS input/dumpstate file")
	return cmd, nil
}

// NewProviderCmd returns the appropriate command to the cmd layer
func NewProviderCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	providerCmd = &cobra.Command{}
	// first, choose the right command
	switch caniCmd.Name() {
	case "init":
		providerCmd, err = NewSessionInitCommand(caniCmd)
	case "cabinet":
		switch caniCmd.Parent().Name() {
		case "add":
			providerCmd, err = NewAddCabinetCommand(caniCmd)
		case "update":
			providerCmd, err = NewUpdateCabinetCommand(caniCmd)
		case "list":
			providerCmd, err = NewListCabinetCommand(caniCmd)
		}
	case "blade":
		switch caniCmd.Parent().Name() {
		case "add":
			providerCmd, err = NewAddBladeCommand(caniCmd)
		case "update":
			providerCmd, err = NewUpdateBladeCommand(caniCmd)
		case "list":
			providerCmd, err = NewListBladeCommand(caniCmd)
		}
	case "node":
		// check for add/update variants
		switch caniCmd.Parent().Name() {
		case "add":
			providerCmd, err = NewAddNodeCommand(caniCmd)
		case "update":
			providerCmd, err = NewUpdateNodeCommand(caniCmd)
		case "list":
			providerCmd, err = NewListNodeCommand(caniCmd)
		}
	case "export":
		providerCmd, err = NewExportCommand(caniCmd)
	case "import":
		providerCmd, err = NewImportCommand(caniCmd)
	default:
		log.Debug().Msgf("Command not implemented by provider: %s %s", caniCmd.Parent().Name(), caniCmd.Name())
		// providerCmd = &cobra.Command{}
	}
	if err != nil {
		return providerCmd, err
	}

	return providerCmd, nil
}

func NewAddCabinetCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}

func NewUpdateCabinetCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}

func NewAddNodeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}

func NewUpdateNodeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}

func NewListCabinetCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}

func NewExportCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	cmd.Flags().Bool("hpcm", false, "Export inventory to HPCM format.")

	return cmd, nil
}

func NewAddBladeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}

func NewUpdateBladeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}

func NewListBladeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}

func NewListNodeCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}

func NewImportCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	return cmd, nil
}
