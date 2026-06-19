package serve

import (
	"fmt"

	"github.com/Cray-HPE/cani/internal/cli"
)

// NewCommand creates the "serve" command
func NewCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "serve",
		Short: "Run the API server.",
		Long:  `Run the API server.`,
		RunE:  serve,
	}
	return cmd
}

func serve(cmd *cli.Command, args []string) error {
	return fmt.Errorf("%s: not yet implemented", cmd.Name())
}
