package session

import (
	"fmt"
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStopCmd represents the session stop command
var SessionImportCmd = &cobra.Command{
	Use:               "import",
	Short:             "TODO THIS IS JUST A SHIM COMMAND",
	Long:              `TODO THIS IS JUST A SHIM COMMAND`,
	SilenceUsage:      true,            // Errors are more important than the usage
	PersistentPreRunE: DatastoreExists, // A session must be active to write to a datastore
	RunE:              importSession,
	// PersistentPostRunE: writeSession,
}

// stopSession stops a session if one exists
func importSession(cmd *cobra.Command, args []string) error {
	ds := root.Conf.Session.DomainOptions.DatastorePath
	providerName := root.Conf.Session.DomainOptions.Provider
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	if root.Conf.Session.Active {
		// Check that the datastore exists before proceeding since we cannot continue without it
		_, err := os.Stat(ds)
		if err != nil {
			return fmt.Errorf("Session is STOPPED with provider '%s' but datastore '%s' does not exist", providerName, ds)
		}
		log.Info().Msgf("Session is STOPPED")
	} else {
		log.Info().Msgf("Session with provider '%s' and datastore '%s' is already STOPPED", providerName, ds)
	}

	log.Info().Msgf("Committing changes to session")

	// Commit the external inventory
	if err := d.Import(cmd.Context()); err != nil {
		return err
	}

	return nil
}
