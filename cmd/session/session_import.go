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
package session

import (
	"fmt"
	"os"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStopCmd represents the session stop command
var SessionImportCmd = &cobra.Command{
	Use:               "import",
	Short:             "TODO THIS IS JUST A SHIM COMMAND",
	Long:              `TODO THIS IS JUST A SHIM COMMAND`,
	SilenceUsage:      true,            // Errors are more important than the usage
	PersistentPreRunE: DatastoreExists, // A session must be active to write to a datastore
	RunE:              importSession,
	// PersistentPostRunE: writeSession,
}

// stopSession stops a session if one exists
func importSession(cmd *cobra.Command, args []string) error {
	// Setup profiling
	// f, err := os.Create("cpu_profile")
	// if err != nil {
	// 	panic(err)
	// }

	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	ds := root.Conf.Session.DomainOptions.DatastorePath
	providerName := root.Conf.Session.DomainOptions.Provider
	d, err := domain.New(root.Conf.Session.DomainOptions)
	if err != nil {
		return err
	}

	if root.Conf.Session.Active {
		// Check that the datastore exists before proceeding since we cannot continue without it
		_, err := os.Stat(ds)
		if err != nil {
			return fmt.Errorf("Session is STOPPED with provider '%s' but datastore '%s' does not exist", providerName, ds)
		}
		log.Info().Msgf("Session is STOPPED")
	} else {
		log.Info().Msgf("Session with provider '%s' and datastore '%s' is already STOPPED", providerName, ds)
	}

	log.Info().Msgf("Committing changes to session")

	// Commit the external inventory
	if err := d.Import(cmd.Context()); err != nil {
		return err
	}

	return nil
}
