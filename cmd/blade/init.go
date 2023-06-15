package blade

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	auto      bool
	cabinet   int
	chassis   int
	blade     int
	recursion bool
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
	AddBladeCmd.MarkFlagRequired("cabinet")
	AddBladeCmd.Flags().IntVar(&chassis, "chassis", 7, "Parent chassis")
	AddBladeCmd.MarkFlagRequired("chassis")
	AddBladeCmd.Flags().IntVar(&blade, "blade", 1, "Blade")
	AddBladeCmd.MarkFlagRequired("blade")
	AddBladeCmd.MarkFlagsRequiredTogether("cabinet", "chassis", "blade")

	AddBladeCmd.Flags().BoolVar(&auto, "auto", false, "Automatically recommend values for parent hardware")
	AddBladeCmd.MarkFlagsRequiredTogether("list-supported-types")

	RemoveBladeCmd.Flags().BoolVarP(&recursion, "recursive", "R", false, "Recursively delete child hardware")

}
