package cdu

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	hwType      string
	supportedHw []string
)

func init() {
	// Add variants to root commands
	root.AddCmd.AddCommand(AddCduCmd)
	root.ListCmd.AddCommand(ListCduCmd)
	root.RemoveCmd.AddCommand(RemoveCduCmd)

	// Add a flag to show supported types
	AddCduCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

}
