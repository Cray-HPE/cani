package sw

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	hwType      string
	supportedHw []string
)

func init() {
	// Add variants to root commands
	root.AddCmd.AddCommand(AddSwitchCmd)
	root.ListCmd.AddCommand(ListSwitchCmd)
	root.RemoveCmd.AddCommand(RemoveSwitchCmd)

	// Add a flag to show supported types
	AddSwitchCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

}
