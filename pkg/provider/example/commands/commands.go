package commands

import "github.com/spf13/cobra"

// FileFlag holds the file path for YAML import
var FileFlag string

// CsvFlag holds the file path for CSV import
var CsvFlag string

// NoColorFlag disables colorized output (inherited from parent)
var NoColorFlag bool

func NewImportCommand(base *cobra.Command) (*cobra.Command, error) {
	cmd := &cobra.Command{}
	cmd.Flags().StringVarP(&FileFlag, "file", "f", "", "YAML inventory file to import")
	cmd.Flags().StringVarP(&CsvFlag, "csv", "c", "", "CSV file to import (columns: PartNumber, Description, Quantity, [ConfigGroup])")
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
