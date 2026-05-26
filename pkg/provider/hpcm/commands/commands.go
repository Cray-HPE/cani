package commands

import "github.com/spf13/cobra"

// NodeJsonFile holds the file path for HPCM node JSON import.
var NodeJsonFile string

// CmConfigFile holds the file path for HPCM cm.config import.
var CmConfigFile string

func NewImportCommand(base *cobra.Command) (*cobra.Command, error) {
	cmd := &cobra.Command{}
	cmd.Flags().StringVarP(&NodeJsonFile, "node-json-file", "f", "", "HPCM node JSON file to import (reads stdin if omitted)")
	cmd.Flags().StringVar(&CmConfigFile, "cm-config", "", "HPCM cm.config file to import")
	return cmd, nil
}

func NewExportCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add export-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewShowCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add show-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewAddCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add add-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewRemoveCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add remove-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewUpdateCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add update-specific flags or subcommands
	return &cobra.Command{}, nil
}
