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
package cdu

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ListCduCmd represents the cdu list command
var ListCduCmd = &cobra.Command{
	Use:   "cdu",
	Short: "List cdus in the inventory.",
	Long:  `List cdus in the inventory.`,
	Args:  cobra.ArbitraryArgs,
	RunE:  listCdu,
}

// listCdu lists cdus in the inventory
func listCdu(cmd *cobra.Command, args []string) error {
	log.Info().Msgf("Not yet implemented")
	return nil
}
