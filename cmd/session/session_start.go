package session

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cfgFile        string
	ds             string
	Conf           *config.Config
	dopts          *domain.NewOpts
	provider       string
	validProviders = []string{"csm"}
)

// SessionStartCmd represents the session start command
var SessionStartCmd = &cobra.Command{
	Use:       "start",
	Short:     "Start a session.",
	Long:      `Start a session.`,
	Args:      validProvider,
	ValidArgs: []string{"csm"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set the provider and datastore path
		Conf.Session.DomainOptions.DatastorePath = ds
		Conf.Session.DomainOptions.Provider = args[0]
		// "Activate" the session
		Conf.Session.Active = true
		log.Info().Msgf("Session is now ACTIVE with provider %s and datastore %s", provider, ds)
		return nil
	},
}

func init() {
	SessionStartCmd.Flags().StringVarP(&ds, "datastore", "d", ds, "Path to datastore")
	cobra.MarkFlagRequired(SessionStartCmd.Flags(), "datastore")
}

// validProvider checks that the provider is valid and that at least one argument is provided
func validProvider(cmd *cobra.Command, args []string) error {
	// Check that at least one argument is provided
	if len(args) < 1 {
		return errors.New("this command requires at least one argument")
	}

	// Check that all arguments are valid
	for _, arg := range args {
		valid := false
		for _, validArg := range cmd.ValidArgs {
			if arg == validArg {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid argument: %s.  Must be one of: %+v", arg, validProviders)
		}
	}

	return nil
}
