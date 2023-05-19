package cabinet

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	hwType      string
	supportedHw []string
)

func init() {
	// Add variants to root commands
	root.AddCmd.AddCommand(AddCabinetCmd)
	root.ListCmd.AddCommand(ListCabinetCmd)
	root.RemoveCmd.AddCommand(RemoveCabinetCmd)

	// Add a flag to show supported types
	AddCabinetCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

}
