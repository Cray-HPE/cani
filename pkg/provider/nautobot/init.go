package nautobot

import (
	"context"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/commands"
	imprt "github.com/Cray-HPE/cani/pkg/provider/nautobot/import"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/logcolor"
)

var nautobotProvider *Nautobot
var clog = logcolor.New("[nautobot] ", false)

func init() {
	nautobotProvider = New()
	provider.Register("nautobot", nautobotProvider)

	// Register the provider getter with the import package to break import cycle
	imprt.SetProviderGetter(func() interface {
		ClearRawData()
		SetRawData(imprt.RawData)
		GetClient() *nautobotapi.ClientWithResponses
		GetContext() context.Context
	} {
		return nautobotProvider
	})
}

// NewProviderCmd creates provider-specific CLI commands.
func (p *Nautobot) NewProviderCmd(base *cli.Command) (*cli.Command, error) {
	switch base.Name() {
	case "import":
		cmd, err := commands.NewImportCommand(base)
		if err != nil {
			return nil, err
		}
		cmd.RunE = p.importNautobot
		return cmd, nil
	case "export":
		return commands.NewExportCommand(base)
	default:
		return nil, nil
	}
}
