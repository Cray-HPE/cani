package chassis

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	hwType      string
	supportedHw []string
)

func init() {
	// Add variants to root commands
	root.AddCmd.AddCommand(AddChassisCmd)
	root.ListCmd.AddCommand(ListChassisCmd)
	root.RemoveCmd.AddCommand(RemoveChassisCmd)

	// Add a flag to show supported types
	AddChassisCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

}
