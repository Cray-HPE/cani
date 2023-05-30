package cabinet

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	hwType      string
	supportedHw []string
	cabinet     int
)

func init() {
	// Add variants to root commands
	root.AddCmd.AddCommand(AddCabinetCmd)
	root.ListCmd.AddCommand(ListCabinetCmd)
	root.RemoveCmd.AddCommand(RemoveCabinetCmd)

	// Add a flag to show supported types
	AddCabinetCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

	// Blades have several parents, so we need to add flags for each
	AddCabinetCmd.Flags().IntVar(&cabinet, "cabinet", 1001, "Cabinet ordinal.")

}
