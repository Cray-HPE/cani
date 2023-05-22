package session

import (
	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionReconcileCmd represents the session reconcile command
var SessionReconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Reconcile a session.",
	Long:  `Reconcile a session.`,
	RunE:  reconcileSession,
	// PersistentPostRunE: reconcileSession,
}

// stopSession stops a session if one exists
func reconcileSession(cmd *cobra.Command, args []string) error {
	ds := root.Conf.Session.DomainOptions.DatastorePath
	provider := root.Conf.Session.DomainOptions.Provider

	if root.Conf.Session.Active {
		log.Info().Msgf("Session with provider '%s' and datastore '%s' is active", provider, ds)
	} else {
		log.Info().Msgf("Session with provider '%s' and datastore '%s' is inactive", provider, ds)
	}

	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	return d.Reconcile()

	// if root.Conf.Session.Active {
	// 	// "Deactivate" the session
	// 	root.Conf.Session.Active = false

	// 	// Check that the datastore exists before proceeding since we cannot continue without it
	// 	_, err := os.Stat(ds)
	// 	if err != nil {
	// 		return fmt.Errorf("Session is STOPPED with provider '%s' but datastore '%s' does not exist", provider, ds)
	// 	}
	// 	log.Info().Msgf("Session is STOPPED")
	// } else {
	// 	log.Info().Msgf("Session with provider '%s' and datastore '%s' is already STOPPED", provider, ds)
	// }

	return nil
}
