package example

import (
	"context"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/transform"
)

// Transform converts queued data into CANI types (devices, racks, cables).
// Delegates to the transform package which processes all records in a single pass.
func (p *Example) Transform(ctx context.Context, existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return transform.Transform(existing)
}
