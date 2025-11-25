package serve

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCommand creates the "serve" command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the API server.",
		Long:  `Run the API server.`,
		RunE:  serve,
	}
	return cmd
}

func serve(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("%s: not yet implemented", cmd.Name())
}
