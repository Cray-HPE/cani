package commands

import "github.com/spf13/cobra"

// RootFlag holds the path to a Redfish ServiceRoot JSON file.
var RootFlag string

func NewImportCommand(base *cobra.Command) (*cobra.Command, error) {
	cmd := &cobra.Command{}
	cmd.Flags().StringVarP(&RootFlag, "root", "r", "",
		"Path to Redfish ServiceRoot JSON file (single object or array; reads stdin if omitted)")
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
