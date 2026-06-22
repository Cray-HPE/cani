package csm

import (
	"context"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/csm/transform"
)

// Transform converts imported data into CANI's inventory format.
// Returns a TransformResult containing devices, racks, and cables.
func (p *Csm) Transform(ctx context.Context, existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return transform.Transform(existing)
}
