/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package add

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/core"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// newBladeCommand creates the "add blade" command
func newBladeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "blade",
		Short:   "Add blades to the inventory.",
		Long:    `Add blades to the inventory.`,
		PreRunE: provider.GetActiveProvider,
		Args:    validDeviceType,
		RunE:    add,
	}

	// Add flags

	return cmd
}

// addBlade adds a blade to the inventory
func addBlade(cmd *cobra.Command, args []string) (err error) {
	if err := autoRecommend(cmd, args); err != nil {
		return fmt.Errorf("could not recommend devices: %w", err)
	}
	return nil
}

func autoRecommend(cmd *cobra.Command, args []string) (err error) {
	auto, _ := cmd.PersistentFlags().GetBool("auto")
	accept, _ := cmd.PersistentFlags().GetBool("accept")

	if auto {
		log.Printf("Auto-recommendation mode enabled. This will suggest values for parent hardware.")
		// TODO: Implement auto-recommendation logic
		if accept {
			auto = true
		} else {
			// Prompt the user to confirm the suggestions
			confirmed, err := core.PromptForConfirmation(fmt.Sprintf("Would you like to accept the recommendations and add the %s", devicetypes.Node))
			if err != nil {
				return err
			}

			auto = confirmed
			if !confirmed {
				return fmt.Errorf("operation canceled by user")
			}
		}
	}
	return nil
}
