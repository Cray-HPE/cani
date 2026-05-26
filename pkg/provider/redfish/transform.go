package redfish

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/redfish/transform"
)

// Transform converts imported data into CANI's inventory format.
// Returns a TransformResult containing devices, racks, and cables.
func (p *Redfish) Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	return transform.Transform(existing)
}
