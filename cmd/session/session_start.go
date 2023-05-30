package session

import (
	"errors"
	"fmt"
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStartCmd represents the session start command
var SessionStartCmd = &cobra.Command{
	Use:                "start",
	Short:              "Start a session.",
	Long:               `Start a session.`,
	Args:               validProvider,
	ValidArgs:          validArgs,
	SilenceUsage:       true, // Errors are more important than the usage
	RunE:               startSession,
	PersistentPostRunE: writeSession,
}

var (
	providerName string
	validArgs    = []string{"csm"}
)

// startSession starts a session if one does not exist
func startSession(cmd *cobra.Command, args []string) error {
	// TODO This is probably not the right way todo this, but hopefully this will be easy way...
	// Sorry Jacob
	if useSimURLs, _ := cmd.Flags().GetBool("csm-sim-urls"); useSimURLs {
		root.Conf.Session.DomainOptions.CsmOptions.BaseUrlSLS = "https://localhost:8443/apis/sls/v1"
		root.Conf.Session.DomainOptions.CsmOptions.BaseUrlHSM = "https://localhost:8443/apis/smd/hsm/v2"
		root.Conf.Session.DomainOptions.CsmOptions.InsecureSkipVerify = true
	} else {
		root.Conf.Session.DomainOptions.CsmOptions.BaseUrlSLS, _ = cmd.Flags().GetString("csm-url-sls")
		root.Conf.Session.DomainOptions.CsmOptions.BaseUrlHSM, _ = cmd.Flags().GetString("csm-url-hsm")
		root.Conf.Session.DomainOptions.CsmOptions.InsecureSkipVerify, _ = cmd.Flags().GetBool("csm-insecure-https")
	}

	// If a session is already active, there is nothing to do but the user may want to overwrite the existing session
	if root.Conf.Session.Active {
		log.Info().Msgf("Session is already ACTIVE.")
		ds := root.Conf.Session.DomainOptions.DatastorePath
		// Check if the json file exists
		if _, err := os.Stat(ds); err == nil {
			// If the json file exists, prompt user for overwrite
			overwrite, err := promptForOverwrite(ds)
			if err != nil {
				return err
			}
			if !overwrite {
				// User chose not to overwrite the file
				os.Exit(0)
			}
		}
	}

	// Create a domain object to interact with the datastore
	var err error
	root.Conf.Session.Domain, err = domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	// Perform provider plugin specific logic at session start
	switch root.Conf.Session.DomainOptions.Provider {
	case string(inventory.CSMProvider):
		// Need to get the systems Roles/SubRole data from the system
		// TODO CASMINST-6417

		// For now just use the defaults
		root.Conf.Session.DomainOptions.CsmOptions.ValidRoles = csm.DefaultValidRoles
		root.Conf.Session.DomainOptions.CsmOptions.ValidSubRoles = csm.DefaultValidSubRolesRoles
	}

	// Validate the external inventory
	err = root.Conf.Session.Domain.Validate()
	if err != nil {
		return errors.Join(err,
			errors.New("External inventory is unstable.  Fix, and check with 'cani validate' before continuing."))
	}

	// "Activate" the session
	root.Conf.Session.Active = true

	ds := root.Conf.Session.DomainOptions.DatastorePath
	provider := root.Conf.Session.DomainOptions.Provider
	log.Info().Msgf("Session is now ACTIVE with provider %s and datastore %s", provider, ds)
	return nil
}

// writeSession writes the session configuration back to the config file
func writeSession(cmd *cobra.Command, args []string) error {
	// Write the configuration back to the file
	cfgFile := cmd.Root().PersistentFlags().Lookup("config").Value.String()
	err := config.WriteConfig(cfgFile, root.Conf)
	if err != nil {
		return err
	}
	return nil
}

func promptForOverwrite(path string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("File %s already exists. Keep session active but overwrite the datastore", path),
		IsConfirm: true,
	}

	_, err := prompt.Run()

	if err != nil {
		if err == promptui.ErrAbort {
			// User chose not to overwrite the file
			return false, nil
		}
		// An error occurred
		return false, err
	}

	// User chose to overwrite the file
	return true, nil
}
