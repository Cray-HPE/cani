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
package cabinet

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// RemoveCabinetCmd represents the cabinet remove command
var RemoveCabinetCmd = &cobra.Command{
	Use:   "cabinet",
	Short: "Remove cabinets from the inventory.",
	Long:  `Remove cabinets from the inventory.`,
	Args:  cobra.ArbitraryArgs,
	RunE:  removeCabinet,
}

// removeCabinet removes a cabinet from the inventory.
func removeCabinet(cmd *cobra.Command, args []string) error {
	log.Info().Msgf("Not yet implemented")
	// for _, arg := range args {
	// 	// Convert the argument to a UUID
	// 	u, err := uuid.Parse(arg)
	// 	if err != nil {
	// 		return fmt.Errorf("Need a UUID to remove: %s", err.Error())
	// 	}
	// 	// Remove item from the inventory
	// 	err = root.Domain.RemoveCabinet(u)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}
