package transform

import (
	"log"
	"slices"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
	"github.com/google/uuid"
)

// providerGetter returns the Csm singleton with raw SLS/SMD data.
// Set by the parent package to break import cycles.
var providerGetter func() interface {
	GetSls() *import_.SlsDumpstate
	GetSmd() *import_.SmdComponentList
}

// SetProviderGetter allows the parent package to provide singleton access.
func SetProviderGetter(getter func() interface {
	GetSls() *import_.SlsDumpstate
	GetSmd() *import_.SmdComponentList
}) {
	providerGetter = getter
}

// IdentityKeys names the keys in the CSM bucket that identify one physical
// device: the xname alone. Transform declares them on its result.
var IdentityKeys = []string{"xname"}

// Transform converts imported SLS/SMD data into CANI's inventory format.
func Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	p := providerGetter()
	sls := p.GetSls()
	if sls == nil || len(sls.Hardware) == 0 {
		log.Println("No SLS data to transform")
		return emptyResult(), nil
	}
	smdMap := buildSmdMap(p.GetSmd())
	result, err := transformSls(sls, smdMap, &existing)
	if err != nil {
		return nil, err
	}
	result.IdentityKeys = slices.Clone(IdentityKeys)
	return result, nil
}

func emptyResult() *devicetypes.TransformResult {
	return &devicetypes.TransformResult{
		Locations: make(map[uuid.UUID]*devicetypes.CaniLocationType),
		Racks:     make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices:   make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		Modules:   make(map[uuid.UUID]*devicetypes.CaniModuleType),
		Frus:      make(map[uuid.UUID]*devicetypes.CaniFruType),
	}
}

// buildSmdMap creates a lookup from xname to SmdComponent.
func buildSmdMap(smd *import_.SmdComponentList) map[string]import_.SmdComponent {
	m := make(map[string]import_.SmdComponent)
	if smd == nil {
		return m
	}
	for _, c := range smd.Components {
		m[c.ID] = c
	}
	return m
}
