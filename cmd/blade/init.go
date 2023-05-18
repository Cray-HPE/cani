package blade

import (
	"github.com/Cray-HPE/cani/cmd"
)

func init() {
	cmd.AddCmd.AddCommand(AddBladeCmd)
	cmd.ListCmd.AddCommand(ListBladeCmd)
	cmd.RemoveCmd.AddCommand(RemoveBladeCmd)

}
