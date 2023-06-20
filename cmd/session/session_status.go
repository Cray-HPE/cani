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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SessionStatusCmd represents the session status command
var SessionStatusCmd = &cobra.Command{
	Use:          "status",
	Short:        "View session status.",
	Long:         `View session status.`,
	SilenceUsage: true, // Errors are more important than the usage
	RunE:         showSession,
}

// showSession shows the status of the session
func showSession(cmd *cobra.Command, args []string) error {
	ds := root.Conf.Session.DomainOptions.DatastorePath
	provider := root.Conf.Session.DomainOptions.Provider
	conf := root.RootCmd.Flag("config").Value.String()

	// If the session is active, check that the datastore exists
	log.Info().Msgf("See %s for session details", conf)
	if root.Conf.Session.Active {
		_, err := os.Stat(ds)
		if err != nil {
			return fmt.Errorf("Session is ACTIVE with provider '%s' but datastore '%s' does not exist", provider, ds)
		}
		log.Info().Msgf("Session is ACTIVE")
	} else {
		log.Info().Msgf("Session is INACTIVE")
	}

	return nil
}
