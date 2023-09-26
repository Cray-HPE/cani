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
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/Cray-HPE/cani/cmd/config"
	"github.com/Cray-HPE/cani/internal/domain"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// SetupDomain checks that the datastore exists
// func SetupDomain(cmd *cobra.Command, args []string) (*domain.Domain, *config.Config, *hardwaretypes.Library, []hardwaretypes.DeviceType, []hardwaretypes.DeviceType, error) {
func SetupDomain(cmd *cobra.Command, args []string) error {
	if cmd.Name() != "init" {
		// Load the session/domain from the config file
		cfgFile = cmd.Root().PersistentFlags().Lookup("config").Value.String()
		if _, err := os.Stat(cfgFile); !os.IsNotExist(err) {
			Conf, err = config.LoadConfig(cfgFile)
			if err != nil {
				return err
			}
		}

		if Debug {
			log.Debug().Msgf("Loaded config file %s", cfgFile)
		}

		// Find an active session
		activeDomains := []*domain.Domain{}
		activeProviders := []string{}
		for p, d := range Conf.Session.Domains {
			log.Debug().Msgf("Checking provider '%s'...", p)
			if d.Active {
				log.Debug().Msgf("Provider '%s' is active", p)
				activeDomains = append(activeDomains, d)
				activeProviders = append(activeProviders, p)
			} else {
				log.Debug().Msgf("Provider '%s' is inactive", p)
			}
		}

		if !cmd.Flags().Changed("list-supported-types") {
			// Error if no sessions are active
			if len(activeProviders) == 0 {
				// These commands are special because they validate hardware in the args
				// so SetupDomain is called manually
				// The timing of events works out such that simply returning the error
				// will exit without the message
				if cmd.Name() == "cabinet" || cmd.Name() == "blade" || cmd.Name() == "node" {
					log.Error().Msgf("No active session.  Run 'session init' to begin.")
					os.Exit(1)
				}
				return errors.Join(
					fmt.Errorf("no active session"),
					fmt.Errorf("run 'session init' to begin"),
				)
			}
		}

		// Check that only one session is active
		var err, joined error
		if len(activeProviders) > 1 {
			for _, p := range activeProviders {
				err = fmt.Errorf("currently active: %v", p)
				joined = errors.Join(err)
			}
			// These commands are special because they validate hardware in the args
			// so SetupDomain is called manually
			// The timing of events works out such that simply returning the error
			// will exit without the message
			if cmd.Name() == "cabinet" || cmd.Name() == "blade" || cmd.Name() == "node" || cmd.Name() == "list" {
				log.Error().Msgf("only one session may be active at a time")
				os.Exit(1)
			}
			return errors.Join(
				fmt.Errorf("only one session may be active at a time"),
				joined,
			)
		}
		activeDomain := activeDomains[0]
		// activeProvider := activeProviders[0]

		if !cmd.Flags().Changed("list-supported-types") {
			// Check that a datastore path is defined
			if activeDomain.DatastorePath == "" {
				// These commands are special because they validate hardware in the args
				// so SetupDomain is called manually
				// The timing of events works out such that simply returning the error
				// will exit without the message
				if cmd.Name() == "cabinet" || cmd.Name() == "blade" || cmd.Name() == "node" || cmd.Name() == "list" {
					log.Error().Msgf("Need a datastore path.  Run 'session init' to begin")
					os.Exit(1)
				}
				return errors.Join(
					fmt.Errorf("need a datastore path"),
					fmt.Errorf("run 'session init' to begin"),
				)
			}
		}

		if !cmd.Flags().Changed("list-supported-types") {
			// Error if the datastore does not exist
			if _, err := os.Stat(activeDomain.DatastorePath); os.IsNotExist(err) {
				ds := activeDomain.DatastorePath
				// These commands are special because they validate hardware in the args
				// so SetupDomain is called manually
				// The timing of events works out such that simply returning the error
				// will exit without the message
				if cmd.Name() == "cabinet" || cmd.Name() == "blade" || cmd.Name() == "node" || cmd.Name() == "list" {
					log.Error().Msgf("Datastore '%s' does not exist.  Run 'session init' to begin", ds)
					os.Exit(1)
				}
				return errors.Join(
					fmt.Errorf("datastore '%s' does not exist", ds),
					fmt.Errorf("run 'session init' to begin"),
				)
			}
		}

		D = activeDomain

		err = D.SetupDomain(cmd, args)
		if err != nil {
			return err
		}

		HwLibrary, err = hardwaretypes.NewEmbeddedLibrary(D.CustomHardwareTypesDir)
		if err != nil {
			return err
		}

		// Get the list of supported hardware types
		CabinetTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.Cabinet)
		BladeTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.NodeBlade)
		NodeTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.NodeBlade)
		MgmtSwitchTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.ManagementSwitch)
		HsnSwitchTypes = HwLibrary.GetDeviceTypesByHardwareType(hardwaretypes.HighSpeedSwitch)
	}
	return nil
}
