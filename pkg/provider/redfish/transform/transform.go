package transform

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/redfish/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// providerGetter returns the Redfish singleton with raw roots.
// Set by the parent package to break import cycles.
var providerGetter func() interface {
	GetRoots() []import_.ServiceRoot
}

// SetProviderGetter allows the parent package to provide singleton access.
func SetProviderGetter(getter func() interface {
	GetRoots() []import_.ServiceRoot
}) {
	providerGetter = getter
}

// Transform converts raw Redfish ServiceRoots into CANI inventory types.
// Each ServiceRoot produces one CaniDeviceType (server) and one
// CaniModuleType (BMC/iLO manager).
// The existing inventory is used to make the import idempotent: if a
// device with matching redfish_uuid, bmc_fqdn, or bmc_hostname already
// exists, its UUID is reused instead of generating a new one.
func Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	p := providerGetter()
	roots := p.GetRoots()
	if len(roots) == 0 {
		log.Println("No raw ServiceRoots to transform")
		return &devicetypes.TransformResult{}, nil
	}
	return transformRoots(roots, &existing)
}

// transformRoots converts raw ServiceRoots into devices and modules.
// existing is used for idempotent deduplication against the current inventory.
func transformRoots(roots []import_.ServiceRoot, existing *devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	result := &devicetypes.TransformResult{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		Modules: make(map[uuid.UUID]*devicetypes.CaniModuleType),
	}

	stepMode := config.Cfg != nil && config.Cfg.StepMode
	noColor := config.Cfg != nil && config.Cfg.NoColor
	opts := visual.ETLOptions{NoColor: noColor}
	tally := visual.StepTally{}

	for i, root := range roots {
		dev := buildDeviceFromRoot(root, existing)

		slug, mq, ms := enrichDeviceFromLibrary(&dev, root)
		result.Devices[dev.ID] = &dev
		tally.Devices++

		if stepMode {
			info := buildRootStepInfo(stepInput{
				Num:        i + 1,
				Total:      len(roots),
				Root:       root,
				Dev:        &dev,
				LibSlug:    slug,
				MatchQuery: mq,
				MatchScore: ms,
			})
			if err := visual.PromptNodeTransformStep(info, tally, opts); err != nil {
				return nil, fmt.Errorf("step interrupted: %w", err)
			}
		}
	}

	log.Printf("Transformed %d ServiceRoot(s) → %d devices",
		len(roots), len(result.Devices))
	return result, nil
}

// buildDeviceFromRoot creates a CaniDeviceType from a Redfish ServiceRoot.
// If a device with matching Redfish metadata already exists in the inventory,
// its UUID is reused to make the import idempotent.
func buildDeviceFromRoot(root import_.ServiceRoot, existing *devicetypes.Inventory) devicetypes.CaniDeviceType {
	id := resolveExistingID(root, existing)

	dev := devicetypes.CaniDeviceType{
		ID:              id,
		Name:            root.Product,
		Manufacturer:    root.Vendor,
		Type:            devicetypes.TypeNode,
		HardwareType:    "server",
		AllowedChildren: []string{"cpu", "dimm", "disk", "gpu", "nic", "power-supply"},
		ObjectMeta:      devicetypes.ObjectMeta{ProviderMetadata: buildProviderMetadata(root)},
	}

	// Set import source from the --root flag value.
	dev.SetImportSource("redfish", commands.RootFlag)

	return dev
}

// resolveExistingID checks whether a device with matching Redfish metadata
// already exists in the inventory. Returns the existing UUID if found,
// otherwise generates a new UUID.
//
// When the ServiceRoot has a BMC FQDN or hostname, a device matches only
// if both the redfish_uuid AND the BMC identity match. This prevents two
// distinct endpoints that share a UUID from collapsing into one device.
func resolveExistingID(root import_.ServiceRoot, existing *devicetypes.Inventory) uuid.UUID {
	if existing == nil {
		return uuid.New()
	}

	fqdn := root.ManagerFQDN()
	host := root.ManagerHostName()
	hasBMCIdentity := fqdn != "" || host != ""

	// When a BMC identity is available, require both UUID and BMC to match.
	if hasBMCIdentity {
		for _, dev := range existing.Devices {
			meta, ok := dev.GetProviderSubMap("redfish")
			if !ok {
				continue
			}
			if !providerValueEquals(meta, "redfish_uuid", root.UUID) {
				continue
			}
			if fqdn != "" && providerValueEquals(meta, "bmc_fqdn", fqdn) {
				log.Printf("Matched existing device %s (%s) by provider metadata", dev.Name, dev.ID)
				return dev.ID
			}
			if host != "" && providerValueEquals(meta, "bmc_hostname", host) {
				log.Printf("Matched existing device %s (%s) by provider metadata", dev.Name, dev.ID)
				return dev.ID
			}
		}
		return uuid.New()
	}

	// Fallback: no BMC identity, match on UUID alone.
	checks := []devicetypes.ProviderKeyCheck{
		{Key: "redfish_uuid", Value: root.UUID},
	}
	if match := existing.FindDeviceByProviderKeys("redfish", checks); match != nil {
		log.Printf("Matched existing device %s (%s) by provider metadata", match.Name, match.ID)
		return match.ID
	}

	return uuid.New()
}

// providerValueEquals returns true when meta[key] equals val (as string).
func providerValueEquals(meta map[string]any, key, val string) bool {
	v, ok := meta[key]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s == val
}

// buildProviderMetadata extracts important metadata from a ServiceRoot.
// All Redfish-specific keys are nested under the "redfish" provider key.
func buildProviderMetadata(root import_.ServiceRoot) map[string]any {
	meta := map[string]any{
		"redfish_version": root.RedfishVersion,
		"redfish_uuid":    root.UUID,
		"vendor":          root.Vendor,
		"odata_type":      root.OdataType,
	}

	if bmc := root.ManagerType(); bmc != "" {
		meta["bmc_type"] = bmc
	}
	if fw := root.ManagerFirmwareVersion(); fw != "" {
		meta["bmc_firmware"] = fw
	}
	if fqdn := root.ManagerFQDN(); fqdn != "" {
		meta["bmc_fqdn"] = fqdn
	}
	if host := root.ManagerHostName(); host != "" {
		meta["bmc_hostname"] = host
	}
	if tag := root.ProductTag(); tag != "" {
		meta["product_tag"] = tag
	}
	if fam := root.SystemFamily(); fam != "" {
		meta["system_family"] = fam
	}

	// NOTE: Serial numbers and asset tags are not present in the
	// ServiceRoot response. They are available from /Systems/ and
	// /Chassis/ endpoints. A future enhancement could add --systems
	// and --chassis flags to import those resources.

	return map[string]any{"redfish": meta}
}
