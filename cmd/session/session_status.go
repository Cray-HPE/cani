package session

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStatusCmd represents the session status command
var SessionStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "View session status.",
	Long:  `View session status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := sessionShow(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

func sessionShow(cmd *cobra.Command, args []string) error {
	_, sesh, err := getDatastoreAndSession(cmd, args)
	if err != nil {
		return err
	}

	if _, err := os.Stat(sesh); err == nil {
		log.Info().Msg(fmt.Sprintf("%s is ACTIVE", sesh))
	} else {
		log.Info().Msg(fmt.Sprintf("%s is INACTIVE", sesh))
	}
	return nil
}

func getDatastoreAndSession(cmd *cobra.Command, args []string) (ds string, sesh string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	if cmd.Parent().Flag("datastore").Changed {
		ds = cmd.Parent().Flag("datastore").Value.String()
		sesh = getSessionFilename(ds)
	} else {
		// Set a default database file if the user did not specify one
		ds = filepath.Join(homeDir, taxonomy.DsPath)
		sesh = getSessionFilename(ds)
	}
	return ds, sesh, nil
}
