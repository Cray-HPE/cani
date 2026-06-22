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
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/internal/core"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

func newRootCommand() *cli.Command {
	// Create a new root command
	// This is where you would set up the command's name, usage, and description.
	cmd := &cli.Command{
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
	cmd.PersistentFlags().String("datastore-path", "", "override path to the datastore file (for testing)")
	cmd.PersistentFlags().StringSlice("types-dirs", nil, "local directories with additional hardware types")
	cmd.PersistentFlags().StringSlice("types-repos", nil, "git repo URLs with additional hardware types")
	cmd.PersistentFlags().Bool("types-repo-clone", false, "clone types repos that are not yet cached locally")
	cmd.PersistentFlags().Bool("types-repo-pull", false, "pull latest changes from types repos on startup")
	cmd.PersistentFlags().Bool("strict", true, "require a resolved device type (slug) for all devices")

	return cmd
}

// envKeyFor maps a config YAML key to its CANI_-prefixed environment variable
// name (e.g. "types_repo_pull" -> "CANI_TYPES_REPO_PULL").
func envKeyFor(yamlKey string) string {
	return "CANI_" + strings.ToUpper(yamlKey)
}

// resolveBool resolves a global boolean setting with precedence
// CLI flag > env var > config file > flag default.
func resolveBool(cmd *cli.Command, flagName, yamlKey string, configVal bool) bool {
	if cmd.Flags().Changed(flagName) {
		v, _ := cmd.Flags().GetBool(flagName)
		return v
	}
	if s, ok := os.LookupEnv(envKeyFor(yamlKey)); ok {
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}
	}
	if config.HasTopLevelKey(yamlKey) {
		return configVal
	}
	v, _ := cmd.Flags().GetBool(flagName)
	return v
}

// resolveStringSlice resolves a global string-slice setting with precedence
// CLI flag > env var (comma-separated) > config file > flag default.
func resolveStringSlice(cmd *cli.Command, flagName, yamlKey string, configVal []string) []string {
	if cmd.Flags().Changed(flagName) {
		v, _ := cmd.Flags().GetStringSlice(flagName)
		return v
	}
	if s, ok := os.LookupEnv(envKeyFor(yamlKey)); ok {
		return strings.Split(s, ",")
	}
	if config.HasTopLevelKey(yamlKey) {
		return configVal
	}
	v, _ := cmd.Flags().GetStringSlice(flagName)
	return v
}

func setupDomain(cmd *cli.Command, args []string) error {
	// 1) load or create the config first so file values are available for
	// precedence resolution below.
	if err := config.Load(cfgFile); err != nil {
		return err
	}

	// Apply precedence (CLI flag > env var > config file > default) for the
	// global settings that were previously managed by viper.
	config.Cfg.Debug = resolveBool(cmd, "debug", "debug", config.Cfg.Debug)
	config.Cfg.Strict = resolveBool(cmd, "strict", "strict", config.Cfg.Strict)
	typesDirs := resolveStringSlice(cmd, "types-dirs", "types_dirs", config.Cfg.TypesDirs)
	typesRepos := resolveStringSlice(cmd, "types-repos", "types_repos", config.Cfg.TypesRepos)
	typesRepoClone := resolveBool(cmd, "types-repo-clone", "types_repo_clone", config.Cfg.TypesRepoClone)
	typesRepoPull := resolveBool(cmd, "types-repo-pull", "types_repo_pull", config.Cfg.TypesRepoPull)

	// Load all hardware type libraries from YAML once at startup.
	devicetypes.Debug = config.Cfg.Debug
	if err := devicetypes.LoadAll(typesDirs, typesRepos, typesRepoClone, typesRepoPull); err != nil {
		return fmt.Errorf("loading device type libraries: %w", err)
	}

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

	// Override the datastore path if the flag was explicitly set.
	// Applied after Save so the override is not persisted to the config file.
	if dsPath, _ := cmd.Flags().GetString("datastore-path"); dsPath != "" {
		config.Cfg.Datastore = dsPath
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

// RootCommand returns the root cli.Command for doc generation and testing.
func RootCommand() *cli.Command {
	return rootCmd
}

// runRoot is the main entrypoint for the cani command
func runRoot(cmd *cli.Command, args []string) error {
	if len(args) == 0 {
		cmd.Help()
	}

	return nil
}
