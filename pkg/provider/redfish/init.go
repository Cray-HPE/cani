package redfish

import (
	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/provider/redfish/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
	"github.com/Cray-HPE/cani/pkg/provider/redfish/transform"
)

// instance is the singleton provider instance
var instance *Redfish

func init() {
	instance = New()
	provider.Register("redfish", instance)

	// Wire import sub-package → singleton (ClearRoots, SetRoots).
	import_.SetProviderGetter(func() interface {
		ClearRoots()
		SetRoots(roots []import_.ServiceRoot)
	} {
		return instance
	})

	// Wire transform sub-package → singleton (GetRoots).
	transform.SetProviderGetter(func() interface {
		GetRoots() []import_.ServiceRoot
	} {
		return instance
	})
}

// NewProviderCmd returns provider-specific CLI commands.
// This is called for each base command (import, add, show, etc.) to allow
// the provider to customize or extend the command.
func (p *Redfish) NewProviderCmd(base *cli.Command) (*cli.Command, error) {
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
