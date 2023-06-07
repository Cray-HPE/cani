package cabinet

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	cabinetNumber int
	vlanId        int
	auto          bool
)

func init() {
	// Add variants to root commands
	root.AddCmd.AddCommand(AddCabinetCmd)
	root.ListCmd.AddCommand(ListCabinetCmd)
	root.RemoveCmd.AddCommand(RemoveCabinetCmd)

	// Add a flag to show supported types
	AddCabinetCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

	// Cabinets
	AddCabinetCmd.Flags().IntVar(&cabinetNumber, "cabinet", 1001, "Cabinet number.")
	AddCabinetCmd.MarkFlagRequired("cabinet")
	AddCabinetCmd.Flags().IntVar(&vlanId, "vlan-id", -1, "Vlan ID for the cabinet.")
	AddCabinetCmd.MarkFlagRequired("vlan-id")
	AddCabinetCmd.MarkFlagsRequiredTogether("cabinet", "vlan-id")
	AddCabinetCmd.Flags().BoolVar(&auto, "auto", false, "Automatically recommend and assign required flags.")
}
