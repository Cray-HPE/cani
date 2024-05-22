package ngsm

import (
	"github.com/Cray-HPE/cani/pkg/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	log.Trace().Msgf("%+v", "github.com/Cray-HPE/cani/internal/provider/ngsm.init")
}

func NewSessionInitCommand(caniCmd *cobra.Command) (cmd *cobra.Command, err error) {
	cmd = utils.CloneCommand(caniCmd)
	cmd.Flags().String("host", "localhost:8000", "Host or FQDN for APIs")
	cmd.Flags().String("token", "", "API token")
	cmd.Flags().String("cacert", "", "Path to the CA certificate file")
	cmd.Flags().StringArrayP("bom", "b", []string{}, "Path to a BoM file (can specify multiple times)")
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
	cmd.Flags().StringP("bom", "b", "", "Path to a BoM file")
	return cmd, nil
}
