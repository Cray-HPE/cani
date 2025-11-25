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
	"strings"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/internal/core"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	cmd.PersistentFlags().StringSlice("types-dirs", nil, "local directories with additional hardware types")
	cmd.PersistentFlags().StringSlice("types-repos", nil, "git repo URLs with additional hardware types")
	cmd.PersistentFlags().Bool("types-repo-clone", false, "clone types repos that are not yet cached locally")
	cmd.PersistentFlags().Bool("types-repo-pull", false, "pull latest changes from types repos on startup")
	cmd.PersistentFlags().Bool("strict", true, "require a resolved device type (slug) for all devices")

	// Bind debug flag to Viper for config/env/flag precedence
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("types_dirs", cmd.PersistentFlags().Lookup("types-dirs"))
	_ = viper.BindPFlag("types_repos", cmd.PersistentFlags().Lookup("types-repos"))
	_ = viper.BindPFlag("types_repo_clone", cmd.PersistentFlags().Lookup("types-repo-clone"))
	_ = viper.BindPFlag("types_repo_pull", cmd.PersistentFlags().Lookup("types-repo-pull"))
	_ = viper.BindPFlag("strict", cmd.PersistentFlags().Lookup("strict"))

	return cmd
}

// initViper initializes Viper for config/env/flag binding
// This sets up the precedence: CLI flags > env vars > config file > defaults
func initViper() {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, "."+core.App)

	viper.SetConfigName(core.App)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Enable environment variable support with CANI_ prefix
	// e.g., CANI_NAUTOBOT_IMPORT_URL maps to nautobot.import.url
	viper.SetEnvPrefix("CANI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read config file if it exists (errors are ignored, config may not exist yet)
	_ = viper.ReadInConfig()
}

func setupDomain(cmd *cobra.Command, args []string) error {
	// Initialize Viper for config/env/flag binding
	initViper()

	// Load all hardware type libraries from YAML once at startup
	devicetypes.Debug = viper.GetBool("debug")
	typesDirs := viper.GetStringSlice("types_dirs")
	typesRepos := viper.GetStringSlice("types_repos")
	typesRepoClone := viper.GetBool("types_repo_clone")
	typesRepoPull := viper.GetBool("types_repo_pull")
	if err := devicetypes.LoadAll(typesDirs, typesRepos, typesRepoClone, typesRepoPull); err != nil {
		return fmt.Errorf("loading device type libraries: %w", err)
	}

	// 1) load or create the config
	if err := config.Load(cfgFile); err != nil {
		return err
	}

	// Apply flag precedence: CLI > env > config > default
	config.Cfg.Debug = viper.GetBool("debug")
	config.Cfg.Strict = viper.GetBool("strict")

	// 2) ensure every registered provider has a key in Conf.Providers
	for name := range provider.GetProviders() {
		if _, ok := config.Cfg.Providers[name]; !ok {
			config.Cfg.Providers[name] = map[string]any{}
		}
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

// RootCommand returns the root cobra.Command for doc generation and testing.
func RootCommand() *cobra.Command {
	return rootCmd
}

// runRoot is the main entrypoint for the cani command
func runRoot(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.Help()
	}

	return nil
}
