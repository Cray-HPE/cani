package pdu

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	hwType      string
	supportedHw []string
)

func init() {
	// Add variants to root commands
	root.AddCmd.AddCommand(AddPduCmd)
	root.ListCmd.AddCommand(ListPduCmd)
	root.RemoveCmd.AddCommand(RemovePduCmd)

	// Add a flag to show supported types
	AddPduCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

}
