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
package session

import (
	"fmt"

	"log"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a session",
		Long:  `Initialize a session`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}

	addProviderSubcommands(cmd)

	return cmd
}

// postSessionInit lives in your session package (or wherever you want)
// it sets ActiveProvider in the config, saves it, etc.
func postSessionInit(cmd *cobra.Command, args []string, p provider.Provider) (err error) {
	cfgFile := cmd.Root().PersistentFlags().Lookup("config").Value.String()

	// 1. EXTRACT call provider's Extract() method
	log.Printf("Extracting data from provider %s", p.Slug())
	if err := p.Extract(cmd, args); err != nil {
		return fmt.Errorf("failed to extract external source for %s: %w", p.Slug(), err)
	}

	// 2. Initialize inventory datastore
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	// 3. Get existing inventory (if any)
	existing, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load existing inventory: %w", err)
	}

	// 4. TRANSFORM call provider's Transform() method
	log.Printf("Transforming devices from provider %s", p.Slug())
	transformed, err := p.Transform(*existing)
	if err != nil {
		return fmt.Errorf("failed to transform extracted data from external source for %s: %w", p.Slug(), err)
	}

	// 5. Merge the transformed devices into the existing inventory
	// Parent must be set to calculate the correct children to be created
	existing.Merge(transformed)

	// 6. LOAD the updated inventory
	if err := datastores.Datastore.Save(existing); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	// 7. Set active provider in config and save
	log.Printf("Setting active provider to %s in config", p.Slug())
	config.Cfg.ActiveProvider = p.Slug()
	if err := config.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

func addProviderSubcommands(initCmd *cobra.Command) {
	for _, p := range provider.GetProviders() {
		if provInit, err := p.NewProviderCmd(initCmd); err == nil && provInit != nil {
			provInit.Use = p.Slug()
			provInit.Short = "Initialize a session with " + p.Slug()

			// wrap the provider's RunE
			orig := provInit.RunE
			provInit.RunE = func(cmd *cobra.Command, args []string) error {
				// // 1) plugin does its thing
				if orig != nil {
					if err := orig(cmd, args); err != nil {
						return err
					}
				}
				// 2) then call your session‐package hook
				return postSessionInit(cmd, args, p)
			}
			initCmd.AddCommand(provInit)
		}
	}
}
