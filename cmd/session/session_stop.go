package session

import (
	"fmt"
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStopCmd represents the session stop command
var SessionStopCmd = &cobra.Command{
	Use:                "stop",
	Short:              "Stop a session.",
	Long:               `Stop a session.`,
	SilenceUsage:       true, // Errors are more important than the usage
	RunE:               stopSession,
	PersistentPostRunE: writeSession,
}

// stopSession stops a session if one exists
func stopSession(cmd *cobra.Command, args []string) error {
	ds := root.Conf.Session.DomainOptions.DatastorePath
	provider := root.Conf.Session.DomainOptions.Provider
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	if root.Conf.Session.Active {
		// Check that the datastore exists before proceeding since we cannot continue without it
		_, err := os.Stat(ds)
		if err != nil {
			return fmt.Errorf("Session is STOPPED with provider '%s' but datastore '%s' does not exist", provider, ds)
		}
		log.Info().Msgf("Session is STOPPED")
	} else {
		log.Info().Msgf("Session with provider '%s' and datastore '%s' is already STOPPED", provider, ds)
	}

	if !commit {
		// Prompt user to commit changes if the commit flag is not set
		commit, err = promptForCommit(ds)
		if err != nil {
			return err
		}
	}
	if commit {
		log.Info().Msgf("Committing changes to session")
		inv, err := d.List()
		if err != nil {
			return err
		}

		// Commit the external inventory
		err = d.Commit(inv)
		if err != nil {
			return err
		}
	}

	// "Deactivate" the session if the function has made it this far
	root.Conf.Session.Active = false

	return nil
}

func promptForCommit(path string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Would you like to reconcile and commit %s", path),
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
