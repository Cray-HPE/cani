package nautobot

import (
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/commands"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/logcolor"
	"github.com/spf13/cobra"
)

var nautobotProvider *Nautobot
var clog = logcolor.New("[nautobot] ", false)

func init() {
	nautobotProvider = New()
	provider.Register("nautobot", nautobotProvider)
}

// NewProviderCmd creates provider-specific CLI commands.
func (p *Nautobot) NewProviderCmd(base *cobra.Command) (*cobra.Command, error) {
	switch base.Name() {
	case "import":
		return commands.NewImportCommand(base)
	case "export":
		return commands.NewExportCommand(base)
	default:
		return nil, nil
	}
}
