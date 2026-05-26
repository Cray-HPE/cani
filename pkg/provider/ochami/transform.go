package ochami

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ochami/transform"
)

// Transform converts imported data into CANI's inventory format.
// Returns a TransformResult containing devices, racks, and cables.
func (p *Ochami) Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	return transform.Transform(existing)
}
