package node

import (
	"os"

	"github.com/spf13/cobra"
)

// AddNodeCmd represents the add node command
var AddNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Add nodes to the inventory.",
	Long:  `Add nodes to the inventory.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}
