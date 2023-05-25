package node

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	hwType      string
	supportedHw []string
	cabinet     int
	chassis     int
	slot        int
	bmc         int
	node        int
	role        string
	subrole     string
	nid         int
	alias       string
)

func init() {
	// Add variants to root commands
	root.AddCmd.AddCommand(AddNodeCmd)
	root.ListCmd.AddCommand(ListNodeCmd)
	root.RemoveCmd.AddCommand(RemoveNodeCmd)
	root.UpdateCmd.AddCommand(UpdateNodeCmd)

	// Add a flag to show supported types
	AddNodeCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

	AddNodeCmd.Flags().StringVar(&role, "role", "", "Role of the node")
	AddNodeCmd.Flags().StringVar(&subrole, "subrole", "", "Subrole of the node")
	AddNodeCmd.Flags().IntVar(&nid, "nid", 0, "NID of the node")
	AddNodeCmd.Flags().StringVar(&alias, "alias", "", "Alias of the node")

	// Blades have several parents, so we need to add flags for each
	UpdateNodeCmd.Flags().IntVar(&cabinet, "cabinet", 1001, "Parent cabinet")
	UpdateNodeCmd.Flags().IntVar(&chassis, "chassis", 7, "Parent chassis")
	UpdateNodeCmd.Flags().IntVar(&slot, "slot", 1, "Parent slot")
	UpdateNodeCmd.Flags().IntVar(&bmc, "bmc", 1, "Parent BMC")
	UpdateNodeCmd.Flags().IntVar(&node, "node", 1, "Node to update")

	// CSM specific options
	// TODO a thought, it might be neat if the options that CANI shows changes based on the active provider
	UpdateNodeCmd.Flags().StringVar(&role, "role", "", "Role of the node")
	UpdateNodeCmd.Flags().StringVar(&subrole, "subrole", "", "Subrole of the node")
	UpdateNodeCmd.Flags().IntVar(&nid, "nid", 0, "NID of the node")
	UpdateNodeCmd.Flags().StringVar(&alias, "alias", "", "Alias of the node")

	UpdateNodeCmd.MarkFlagsRequiredTogether("cabinet", "chassis", "slot")

}
