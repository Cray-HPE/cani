package serve

import (
	"fmt"

	"github.com/Cray-HPE/cani/pkg/datastores"
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
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	return fmt.Errorf("not yet implemented")
}
