package export

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

type serviceRoot struct {
	OdataID        string  `json:"@odata.id"`
	OdataType      string  `json:"@odata.type"`
	ID             string  `json:"Id"`
	UUID           string  `json:"UUID"`
	Product        string  `json:"Product"`
	Vendor         string  `json:"Vendor"`
	RedfishVersion string  `json:"RedfishVersion"`
	Oem            oemData `json:"Oem"`
}

type oemData struct {
	Hpe *hpeOem `json:"Hpe,omitempty"`
}

type hpeOem struct {
	Manager []hpeManager `json:"Manager,omitempty"`
}

type hpeManager struct {
	FQDN        string `json:"FQDN,omitempty"`
	HostName    string `json:"HostName,omitempty"`
	ManagerType string `json:"ManagerType,omitempty"`
}

// Export writes the inventory devices as a Redfish-compatible JSON array of
// ServiceRoot objects to stdout. Only "node" type devices are exported
// (Redfish is BMC-discovery, switches are excluded).
//
// When dryRun is true, Export reports how many ServiceRoots would be written
// (to the log) and emits no payload, matching the dry-run semantics honored by
// other providers' exporters.
func Export(existing devicetypes.Inventory, dryRun bool) error {
	var roots []serviceRoot
	for _, dev := range existing.Devices {
		if dev == nil {
			continue
		}
		if dev.Type != devicetypes.TypeNode {
			continue
		}
		r := serviceRoot{
			OdataID:        "/redfish/v1",
			OdataType:      "#ServiceRoot.v1_13_0.ServiceRoot",
			ID:             "RootService",
			UUID:           dev.ID.String(),
			Product:        dev.Model,
			Vendor:         dev.GetVendor(),
			RedfishVersion: "1.13.0",
			Oem: oemData{
				Hpe: &hpeOem{
					Manager: []hpeManager{{
						FQDN:        dev.Name + "-bmc.local",
						HostName:    dev.Name + "-bmc",
						ManagerType: "iLO 6",
					}},
				},
			},
		}
		roots = append(roots, r)
	}

	if dryRun {
		log.Printf("[dry-run] would export %d Redfish ServiceRoot(s); no payload written", len(roots))
		return nil
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(roots); err != nil {
		return fmt.Errorf("encoding Redfish ServiceRoots: %w", err)
	}
	return nil
}
