package session

import (
	"fmt"
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/spf13/cobra"
)

// validProvider checks that the provider is valid and that at least one argument is provided
func validProvider(cmd *cobra.Command, args []string) error {
	// Check that at least one argument is provided
	if len(args) < 1 {
		return fmt.Errorf("Need a provider.  Choose from: %+v", validArgs)
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
			return fmt.Errorf("%s is not a valid provider.  Valid providers: %+v", arg, validArgs)
		}
	}

	return nil
}

// DatastoreExists checks that the datastore exists
func DatastoreExists(cmd *cobra.Command, args []string) error {
	// Check that at least one argument is provided
	if !root.Conf.Session.Active {
		return fmt.Errorf("No active session.  Run 'session start' to begin")
	}

	// Check that at least one argument is provided
	if root.Conf.Session.DomainOptions.DatastorePath == "" {
		return fmt.Errorf("Need a datastore path.  Run 'session start' to begin")
	}

	// if datastore does not exist
	if _, err := os.Stat(root.Conf.Session.DomainOptions.DatastorePath); os.IsNotExist(err) {
		ds := root.Conf.Session.DomainOptions.DatastorePath
		return fmt.Errorf("Datastore '%s' does not exist.  Run 'session start' to begin", ds)
	}

	return nil
}
