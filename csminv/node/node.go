package node

import (
	"os"

	"github.com/spf13/cobra"
)

// AddPhysicalCmd represents the add node physical command
var AddNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Add node to the inventory.",
	Long:  `Add node to the inventory.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}
