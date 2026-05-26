package export

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

type node struct {
	Name       string              `json:"name"`
	UUID       string              `json:"uuid,omitempty"`
	Type       string              `json:"type"`
	Managed    bool                `json:"managed"`
	Platform   *platformSettings   `json:"platform,omitempty"`
	Management *managementSettings `json:"management,omitempty"`
}

type platformSettings struct {
	Name string `json:"name,omitempty"`
}

type managementSettings struct {
	CardType      string `json:"cardType,omitempty"`
	CardIpAddress string `json:"cardIpAddress,omitempty"`
}

// Export writes the inventory devices as an HPCM-compatible JSON array of
// nodes to stdout. Only devices of type "node" or "switch" are included.
func Export(existing devicetypes.Inventory) error {
	var nodes []node
	for _, dev := range existing.Devices {
		if dev == nil {
			continue
		}
		n := node{
			Name:    dev.Name,
			UUID:    dev.ID.String(),
			Managed: true,
		}
		switch dev.Type {
		case devicetypes.TypeNode:
			n.Type = "compute"
		case devicetypes.TypeSwitch:
			n.Type = "switch"
		default:
			continue
		}
		if dev.Model != "" {
			n.Platform = &platformSettings{Name: dev.Model}
		}
		n.Management = &managementSettings{CardType: "iLO6"}
		nodes = append(nodes, n)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(nodes); err != nil {
		return fmt.Errorf("encoding HPCM nodes: %w", err)
	}
	return nil
}
