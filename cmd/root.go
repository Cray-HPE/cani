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
	"fmt"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/internal/core"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	// Create a new root command
	// This is where you would set up the command's name, usage, and description.
	cmd := &cobra.Command{
		Use:               core.App,
		Short:             core.ShortDescription,
		Long:              core.LongDescription,
		PersistentPreRunE: setupDomain, // the domain object is needed for all provider operations, load it early from root so it is available to all subcommands
		RunE:              runRoot,
		Version:           version(),
	}
	// allow user to override config file path
	home, _ := os.UserHomeDir()
	defaultCfg := filepath.Join(home, "."+core.App, core.App+".yml")
	cmd.PersistentFlags().StringVar(&cfgFile, "config", defaultCfg, "config file")
	cmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	cmd.PersistentFlags().String("datastore", "json", "datastore type (json, postgres)")

	return cmd
}

func setupDomain(cmd *cobra.Command, args []string) error {
	// 1) load or create the config
	if err := config.Load(cfgFile); err != nil {
		return err
	}

	// 2) ensure every registered provider has a key in Conf.Providers
	for name := range provider.GetProviders() {
		if _, ok := config.Cfg.Providers[name]; !ok {
			config.Cfg.Providers[name] = map[string]any{}
		}
	}
	// no active provider? skip merges
	if config.Cfg.ActiveProvider == "" {
		return nil
	}

	// write defaults (in case we injected new maps)
	if err := config.Save(cfgFile); err != nil {
		return err
	}

	// 3) hand each plugin its own section of the map
	for name, p := range provider.GetProviders() {
		if cfgSection, ok := config.Cfg.Providers[name]; ok {
			if c, ok := p.(interface {
				Configure(map[string]any) error
			}); ok {
				if err := c.Configure(cfgSection); err != nil {
					return fmt.Errorf("configuring provider %s: %w", name, err)
				}
			}
		}
	}

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runRoot is the main entrypoint for the cani command
func runRoot(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.Help()
	}

	return nil
}
