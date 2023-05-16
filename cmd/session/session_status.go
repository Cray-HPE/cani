package session

import (
	"fmt"

	"github.com/Cray-HPE/cani/internal/cani/domain"
	"github.com/spf13/cobra"
)

// SessionStatusCmd represents the session status command
var SessionStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "View session status.",
	Long:  `View session status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := sessionShow(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

func sessionShow(cmd *cobra.Command, args []string) error {
	// Save the crafted session data to the datastore
	if domain.Data != nil && domain.Data.SessionActive {
		fmt.Println("Session is active")
	} else {
		fmt.Println("No active session")
	}
	// inv, err := domain.Data.List()
	// if err != nil {
	// 	return err
	// }
	// for _, hw := range inv.Hardware {
	// 	// TODO print out the hardware in a nice format
	// 	// TODO print out the hardware type
	// 	fmt.Println(hw.Model, hw.Type)
	// }
	return nil
}
