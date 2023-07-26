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
	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddCduCmd represents the cdu add command
var AddCduCmd = &cobra.Command{
	Use:               "cdu",
	Short:             "Add cdus to the inventory.",
	Long:              `Add cdus to the inventory.`,
	PersistentPreRunE: root.DatastoreExists, // A session must be active to write to a datastore
	Args:              validHardware,        // Hardware can only be valid if defined in the hardware library
	RunE:              addCdu,               // Add a cdu when this sub-command is called
}

// addCdu adds a cdu to the inventory
func addCdu(cmd *cobra.Command, args []string) error {
	// Create a domain object to interact with the datastore
	_, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}
	log.Info().Msgf("Not yet implemented")
	// Remove the cdu from the inventory using domain methods
	// TODO:
	// err = d.AddCdu()
	// if err != nil {
	// 	return err
	// }
	// log.Info().Msgf("Added cdu %s", args[0])
	return nil
}
