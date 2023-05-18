package blade

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	cabinet     int
	chassis     int
	slot        int
	hwType      string
	supportedHw []string
	recursion   bool
)

func init() {
	// Add blade variants to root commands
	root.AddCmd.AddCommand(AddBladeCmd)
	root.ListCmd.AddCommand(ListBladeCmd)
	root.RemoveCmd.AddCommand(RemoveBladeCmd)

	// Add a flag to show supported types
	AddBladeCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

	// Blades have several parents, so we need to add flags for each
	AddBladeCmd.Flags().IntVar(&cabinet, "cabinet", 1001, "Parent cabinet")
	// cobra.MarkFlagRequired(AddBladeCmd.Flags(), "cabinet")

	AddBladeCmd.Flags().IntVar(&chassis, "chassis", 7, "Parent chassis")
	// cobra.MarkFlagRequired(AddBladeCmd.Flags(), "chassis")

	AddBladeCmd.Flags().IntVar(&slot, "slot", 1, "Parent slot")
	// cobra.MarkFlagRequired(AddBladeCmd.Flags(), "slot")

	RemoveBladeCmd.Flags().BoolVarP(&recursion, "recursive", "R", false, "Parent slot")
	// cobra.MarkFlagRequired(AddBladeCmd.Flags(), "slot")
}
