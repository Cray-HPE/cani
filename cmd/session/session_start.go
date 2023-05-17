package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/cani/domain"
	"github.com/Cray-HPE/cani/internal/cani/external-inventory-provider/csm"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	dopts          *domain.NewOpts
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
		err := initDomain(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

func getSessionFilename(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	sessionPath := base + taxonomy.SessionExt
	return sessionPath
}

// createSession checks if a session is already active and initializes a new one
func createSession(path string) error {
	// Check if the session file exists
	if _, err := os.Stat(path); err == nil {
		// If the session file exists, a session is already active
		return fmt.Errorf("a session is already active")
	}

	// Create the session file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Close()

	return nil
}

func initDomain(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	_, sesh, err := getDatastoreAndSession(cmd, args)
	if err != nil {
		return err
	}
	// Get the provider
	var provider string
	provider = cmd.Parent().Flag("datastore").Value.String()

	var ds string
	// If the datastore flag is set, use that path for the datastore
	if cmd.Parent().Flag("datastore").Changed {
		ds = cmd.Parent().Flag("datastore").Value.String()
		sesh = getSessionFilename(ds)
		log.Info().Msg(fmt.Sprintf("Using user-defined datastore: %s (%s)", ds, sesh))

		// if the domain data is nil or the session is not active, create a new domain
		if domain.Data == nil || !domain.Data.SessionActive {
			dopts = &domain.NewOpts{
				DatastorePath: ds,
				Provider:      provider,
				EIPCSMOpts:    csm.NewOpts{},
			}
		}
	} else {
		// Set a default database file if the user did not specify one
		ds = filepath.Join(homeDir, taxonomy.DsPath)
		sesh = getSessionFilename(ds)
		log.Info().Msg(fmt.Sprintf("Using default datastore: %s (%s)", ds, sesh))

		dopts = &domain.NewOpts{
			DatastorePath: ds,
			Provider:      provider,
			EIPCSMOpts:    csm.NewOpts{},
		}
	}

	// Set the Global domain Data using the options provided
	domain.Data, err = domain.New(dopts)
	if err != nil {
		return err
	}

	// Create a new session if all checks pass
	err = createSession(sesh)
	if err != nil {
		return err
	}

	log.Info().Msg(fmt.Sprintf("Session started: %s", sesh))

	return nil
}

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
