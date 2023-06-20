/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package blade

import (
	"fmt"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// RemoveBladeCmd represents the blade remove command
var RemoveBladeCmd = &cobra.Command{
	Use:   "blade",
	Short: "Remove blades from the inventory.",
	Long:  `Remove blades from the inventory.`,
	Args:  cobra.ArbitraryArgs,
	RunE:  removeBlade,
}

// removeBlade removes a blade from the inventory.
func removeBlade(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		// Convert the argument to a UUID
		u, err := uuid.Parse(arg)
		if err != nil {
			return fmt.Errorf("Need a UUID to remove: %s", err.Error())
		}

		d, err := domain.New(root.Conf.Session.DomainOptions)
		if err != nil {
			return err
		}

		// Remove the blade from the inventory
		err = d.RemoveBlade(u, recursion)
		if err != nil {
			return err
		}
	}
	return nil
}
