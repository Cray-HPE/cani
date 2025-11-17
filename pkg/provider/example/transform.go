package example

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/transform"
)

// Transform converts queued data into CANI types (devices, racks, cables).
// Delegates to the transform package which processes all records in a single pass.
func (p *Example) Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	return transform.Transform(existing)
}
