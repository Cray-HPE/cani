package example

import (
	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/provider/example/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/Cray-HPE/cani/pkg/provider/example/transform"
)

// instance is the singleton provider instance
var instance *Example

func init() {
	instance = New()
	provider.Register("example", instance)

	// Register the provider getter with the import package to break import cycle
	import_.SetProviderGetter(func() interface {
		ClearRecords()
		SetRecords(records []import_.CsvRecord)
	} {
		return instance
	})

	// Register the DCIM provider getter with the import package
	import_.SetDcimProviderGetter(func() interface {
		SetDcimRecords(data *import_.DcimCSV)
		ClearDcimRecords()
		IsDcimImport() bool
	} {
		return instance
	})

	// Register the provider getter with the transform package to break import cycle
	transform.SetProviderGetter(func() interface {
		GetRecords() []import_.CsvRecord
	} {
		return instance
	})

	// Register the DCIM provider getter with the transform package
	transform.SetDcimProviderGetter(func() interface {
		GetDcimRecords() *import_.DcimCSV
		IsDcimImport() bool
	} {
		return instance
	})
}

// GetInstance returns the singleton Example provider instance.
// Used by sub-packages to store/access raw records between ETL phases.
func GetInstance() *Example {
	return instance
}

// NewProviderCmd returns provider-specific CLI commands.
// This is called for each base command (import, add, show, etc.) to allow
// the provider to customize or extend the command.
func (p *Example) NewProviderCmd(base *cli.Command) (*cli.Command, error) {
	// Switch on the base command name to provide customizations
	switch base.Name() {
	case "import":
		return commands.NewImportCommand(base)

	case "export":
		return commands.NewExportCommand(base)

	case "show":
		return commands.NewShowCommand(base)

	case "add":
		return commands.NewAddCommand(base)

	case "remove":
		return commands.NewRemoveCommand(base)

	case "update":
		return commands.NewUpdateCommand(base)

	default:
		// No customization for this command
		return base, nil
	}
}
