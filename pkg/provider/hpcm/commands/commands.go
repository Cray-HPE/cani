package commands

import "github.com/Cray-HPE/cani/internal/cli"

// NodeJsonFile holds the file path for HPCM node JSON import.
var NodeJsonFile string

// CmConfigFile holds the file path for HPCM cm.config import.
var CmConfigFile string

func NewImportCommand(base *cli.Command) (*cli.Command, error) {
	cmd := &cli.Command{}
	cmd.Flags().StringVarP(&NodeJsonFile, "node-json-file", "f", "", "HPCM node JSON file to import (reads stdin if omitted)")
	cmd.Flags().StringVar(&CmConfigFile, "cm-config", "", "HPCM cm.config file to import")
	return cmd, nil
}

func NewExportCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add export-specific flags or subcommands
	return &cli.Command{}, nil
}

func NewShowCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add show-specific flags or subcommands
	return &cli.Command{}, nil
}

func NewAddCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add add-specific flags or subcommands
	return &cli.Command{}, nil
}

func NewRemoveCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add remove-specific flags or subcommands
	return &cli.Command{}, nil
}

func NewUpdateCommand(base *cli.Command) (*cli.Command, error) {
	// TODO: Add update-specific flags or subcommands
	return &cli.Command{}, nil
}
