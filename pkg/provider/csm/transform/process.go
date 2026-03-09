package transform

import (
	"fmt"
	"log"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
	"github.com/google/uuid"
)

// transformSls runs the multi-pass transform over SLS hardware.
func transformSls(
	sls *import_.SlsDumpstate,
	smdMap map[string]import_.SmdComponent,
	existing *devicetypes.Inventory,
) (*devicetypes.TransformResult, error) {
	result := emptyResult()

	classified := classifyAll(sls, smdMap)
	byXname := make(map[string]uuid.UUID)

	processCabinets(classified, result, byXname, existing)
	processChassis(classified, result, byXname, existing)
	processBlades(classified, result, byXname, existing)
	processNodeCards(classified, result, byXname, existing)
	processNodes(classified, result, byXname, existing)
	processMgmtSwitches(classified, result, byXname, existing)
	processNetworkSwitches(classified, result, byXname, existing)
	processCabinetPDUs(classified, result, byXname, existing)

	// Assign rack U positions so the rack view renders correctly.
	assignRackPositions(result)

	// Set import_source on all devices.
	for _, dev := range result.Devices {
		dev.SetImportSource("csm", "sls")
	}

	log.Printf("Transform: %d devices, %d racks",
		len(result.Devices), len(result.Racks))
	return result, nil
}

// classifyAll classifies and sorts all SLS hardware by xname.
func classifyAll(
	sls *import_.SlsDumpstate,
	smdMap map[string]import_.SmdComponent,
) []CsmClassification {
	var cls []CsmClassification
	for _, hw := range sls.Hardware {
		cl := classifyHardware(hw, smdMap)
		if cl.Warning != "" {
			log.Printf("Warning: %s", cl.Warning)
		}
		if !cl.Skip {
			cls = append(cls, cl)
		}
	}
	sort.Slice(cls, func(i, j int) bool {
		return cls[i].Xname.Raw < cls[j].Xname.Raw
	})
	return cls
}

// processCabinets creates a rack and device for each cabinet.
func processCabinets(
	cls []CsmClassification,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) {
	for _, cl := range cls {
		if cl.CaniType != devicetypes.TypeCabinet {
			continue
		}
		devID := resolveExistingID(cl.Xname.Raw, existing)
		byXname[cl.Xname.Raw] = devID

		// Log when the cabinet is new to the datastore.
		if !existsInInventory(cl.Xname.Raw, existing) {
			locPath := fmt.Sprintf("System:0->Cabinet:%d", cl.Xname.Cabinet)
			log.Printf("Cabinet %s does not exist in datastore at %s",
				cl.Xname.Raw, locPath)
		}

		meta := CsmMetadata{Xname: cl.Xname.Raw, Class: cl.Hardware.Class}
		meta.HMNVlan = extractCabinetHMNVlan(cl.Hardware.ExtraProperties)

		result.Devices[devID] = &devicetypes.CaniDeviceType{
			ID:               devID,
			Name:             cl.Xname.Raw,
			Type:             devicetypes.TypeCabinet,
			Status:           "active",
			ProviderMetadata: map[string]any{"csm": toProviderMetadata(meta)},
		}
		slug := resolveSlug(cl)
		if slug != "" {
			_ = devicetypes.ApplyDeviceType(result.Devices[devID], slug)
			log.Printf("Cabinet %s device type slug is %s",
				cl.Xname.Raw, slug)
		}

		rackID := resolveExistingRackID(cl.Xname.Raw, existing)
		result.Racks[rackID] = &devicetypes.CaniRackType{
			ID:               rackID,
			Name:             cl.Xname.Raw,
			UHeight:          cabinetUHeight(cl.Hardware.Class),
			Status:           "active",
			ProviderMetadata: map[string]any{"csm": toProviderMetadata(meta)},
		}
		result.Devices[devID].Parent = rackID
	}
}

// cabinetUHeight returns the height based on class.
func cabinetUHeight(class string) int {
	switch class {
	case "Mountain":
		return 48
	case "Hill":
		return 44
	default:
		return 42
	}
}

// processChassis creates a device for each chassis.
func processChassis(
	cls []CsmClassification,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) {
	for _, cl := range cls {
		if cl.CaniType != devicetypes.TypeChassis {
			continue
		}
		devID := resolveExistingID(cl.Xname.Raw, existing)
		byXname[cl.Xname.Raw] = devID
		parentID := resolveParent(
			cl.Xname.Parent(), devicetypes.TypeCabinet,
			cl.Hardware.Class, result, byXname, existing,
		)
		dev := newDevice(devID, cl, byXname)
		dev.Parent = parentID
		result.Devices[devID] = dev
	}
}

// processBlades creates a device for each ComputeModule.
func processBlades(
	cls []CsmClassification,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) {
	for _, cl := range cls {
		if cl.CaniType != devicetypes.TypeBlade {
			continue
		}
		devID := resolveExistingID(cl.Xname.Raw, existing)
		byXname[cl.Xname.Raw] = devID
		parentID := resolveParent(
			cl.Xname.Parent(), devicetypes.TypeChassis,
			cl.Hardware.Class, result, byXname, existing,
		)
		dev := newDevice(devID, cl, byXname)
		dev.Parent = parentID
		result.Devices[devID] = dev
	}
}

// processNodeCards creates a device for each NodeBMC.
func processNodeCards(
	cls []CsmClassification,
	result *devicetypes.TransformResult,
	byXname map[string]uuid.UUID,
	existing *devicetypes.Inventory,
) {
	for _, cl := range cls {
		if cl.CaniType != devicetypes.TypeNodeCard {
			continue
		}
		devID := resolveExistingID(cl.Xname.Raw, existing)
		byXname[cl.Xname.Raw] = devID
		parentID := resolveParent(
			cl.Xname.Parent(), devicetypes.TypeBlade,
			cl.Hardware.Class, result, byXname, existing,
		)
		dev := newDevice(devID, cl, byXname)
		dev.Parent = parentID
		result.Devices[devID] = dev
	}
}

// extractCabinetHMNVlan returns the HMN VLan from SLS cabinet
// ExtraProperties.Networks.cn.HMN.VLan, or 0 if not found.
func extractCabinetHMNVlan(ep map[string]any) int {
	if ep == nil {
		return 0
	}
	cab, err := import_.DecodeExtraProperties[import_.SlsCabinetExtraProperties](ep)
	if err != nil {
		return 0
	}
	cn, ok := cab.Networks["cn"]
	if !ok {
		return 0
	}
	hmn, ok := cn["HMN"]
	if !ok {
		return 0
	}
	return hmn.VLan
}

// existsInInventory reports whether a device with the given CSM xname
// already exists in the inventory.
func existsInInventory(xname string, existing *devicetypes.Inventory) bool {
	if existing == nil {
		return false
	}
	checks := []devicetypes.ProviderKeyCheck{
		{Key: "xname", Value: xname},
	}
	return existing.FindDeviceByProviderKeys("csm", checks) != nil
}
