package session

import (
	"github.com/Cray-HPE/cani/internal/cani/domain"
	"github.com/Cray-HPE/cani/internal/cani/external-inventory-provider/csm"
	"github.com/spf13/cobra"
)

// SessionStartCmd represents the session start command
var SessionStartCmd = &cobra.Command{
	Use:               "start",
	Short:             "Start a session.",
	Long:              `Start a session.`,
	PersistentPreRunE: preRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := preRun(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

func preRun(cmd *cobra.Command, args []string) error {
	dopts := &domain.NewOpts{
		DatastorePath: "./session.db",
		Provider:      "csm",
		EIPCSMOpts:    csm.NewOpts{},
	}
	var err error
	domain.Data, err = domain.New(dopts)
	if err != nil {
		return err
	}
	// Activate the session
	domain.Data.SessionActive = true
	return nil
}
