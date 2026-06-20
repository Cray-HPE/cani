package commands

import "github.com/Cray-HPE/cani/internal/cli"

// RootFlag holds the path to a Redfish ServiceRoot JSON file.
var RootFlag string

// DryRunFlag, when set, reports what an export would emit without writing the
// Redfish payload, matching the dry-run semantics of other providers.
var DryRunFlag bool

func NewImportCommand(base *cli.Command) (*cli.Command, error) {
	cmd := &cli.Command{}
	cmd.Flags().StringVarP(&RootFlag, "root", "r", "",
		"Path to Redfish ServiceRoot JSON file (single object or array; reads stdin if omitted)")
	return cmd, nil
}

func NewExportCommand(base *cli.Command) (*cli.Command, error) {
	cmd := &cli.Command{}
	cmd.Flags().BoolVar(&DryRunFlag, "dry-run", false,
		"Report what would be exported without writing the Redfish payload")
	return cmd, nil
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
