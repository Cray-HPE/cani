package transform

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// TransformDcim converts parsed DCIM CSV data into a TransformResult.
// Uses a 6-pass algorithm: roles → racks → devices → modules → interfaces → connections.
func TransformDcim(existing devicetypes.Inventory, data *import_.DcimCSV) (*devicetypes.TransformResult, error) {
	initInventoryMaps(&existing)
	isolateInventoryMaps(&existing)

	result := &devicetypes.TransformResult{
		Locations: make(map[uuid.UUID]*devicetypes.CaniLocationType),
		Racks:     make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices:   make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		Modules:   make(map[uuid.UUID]*devicetypes.CaniModuleType),
		Cables:    make(map[uuid.UUID]*devicetypes.CaniCableType),
	}

	// Pass 0: Roles and statuses (metadata catalog)
	meta, err := transformMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("transformMetadata: %w", err)
	}
	result.Metadata = meta

	// Pass 0b: Locations
	locationsByName, err := transformDcimLocations(data, result, &existing)
	if err != nil {
		return nil, fmt.Errorf("transformDcimLocations: %w", err)
	}

	// Pass 1: Racks
	racksByName, err := transformDcimRacks(data, result, &existing, locationsByName)
	if err != nil {
		return nil, fmt.Errorf("transformDcimRacks: %w", err)
	}

	// Pass 2: Devices
	err = transformDcimDevices(data, result, &existing, racksByName)
	if err != nil {
		return nil, fmt.Errorf("transformDcimDevices: %w", err)
	}

	// Pass 3: Modules
	err = transformDcimModules(data, result, &existing)
	if err != nil {
		return nil, fmt.Errorf("transformDcimModules: %w", err)
	}

	// Pass 3b: Interfaces (per-interface metadata such as MAC addresses)
	if err = transformDcimInterfaces(data, result, &existing); err != nil {
		return nil, fmt.Errorf("transformDcimInterfaces: %w", err)
	}

	// Pass 4: Connections
	err = transformDcimConnections(data, result, &existing)
	if err != nil {
		return nil, fmt.Errorf("transformDcimConnections: %w", err)
	}

	// Pass 5: Validate catalog references
	if err = validateReferences(result); err != nil {
		return nil, err
	}

	log.Printf("DCIM CSV transformed: %d roles, %d locations, %d racks, %d devices, %d modules, %d cables",
		len(data.Roles), len(result.Locations), len(result.Racks), len(result.Devices), len(result.Modules), len(result.Cables))

	return result, nil
}

// isolateInventoryMaps replaces the maps the transform mutates during resolution
// with shallow copies. Existing entries stay visible for intra-import lookups,
// but new entries added while transforming do not leak into the caller's
// inventory. This lets the merge phase deduplicate new objects by name instead
// of being short-circuited by a polluted UUID match, making re-import of the
// same DCIM CSV idempotent.
func isolateInventoryMaps(inv *devicetypes.Inventory) {
	locations := make(map[uuid.UUID]*devicetypes.CaniLocationType, len(inv.Locations))
	for k, v := range inv.Locations {
		locations[k] = v
	}
	inv.Locations = locations

	racks := make(map[uuid.UUID]*devicetypes.CaniRackType, len(inv.Racks))
	for k, v := range inv.Racks {
		racks[k] = v
	}
	inv.Racks = racks

	devices := make(map[uuid.UUID]*devicetypes.CaniDeviceType, len(inv.Devices))
	for k, v := range inv.Devices {
		devices[k] = v
	}
	inv.Devices = devices
}

// validateReferences checks that device, rack, and location role and status
// references resolve against the metadata catalog defined by the same import.
// Unknown references are logged as warnings; under strict mode any unresolved
// reference aborts the import. A catalog that is empty is treated as
// "not validated" so imports relying on export-time auto-creation still work.
func validateReferences(result *devicetypes.TransformResult) error {
	if result.Metadata == nil {
		return nil
	}
	roles := metadataNameSet(result.Metadata.Roles)
	statuses := metadataNameSet(result.Metadata.Statuses)

	var problems []string
	for _, dev := range result.Devices {
		if dev == nil {
			continue
		}
		problems = appendUnknownRef(problems, roles, dev.Role, "device", dev.Name, "role")
		problems = appendUnknownRef(problems, statuses, dev.Status, "device", dev.Name, "status")
	}
	for _, rack := range result.Racks {
		if rack != nil {
			problems = appendUnknownRef(problems, statuses, rack.Status, "rack", rack.Name, "status")
		}
	}
	for _, loc := range result.Locations {
		if loc != nil {
			problems = appendUnknownRef(problems, statuses, loc.Status, "location", loc.Name, "status")
		}
	}

	for _, p := range problems {
		log.Printf("WARN: %s", p)
	}
	if len(problems) > 0 && config.Cfg != nil && config.Cfg.Strict {
		return fmt.Errorf("DCIM CSV has %d unresolved catalog reference(s)", len(problems))
	}
	return nil
}

// metadataNameSet builds a set of catalog entry names for membership checks.
func metadataNameSet(entries []devicetypes.MetadataEntry) map[string]bool {
	set := make(map[string]bool, len(entries))
	for _, e := range entries {
		set[e.Name] = true
	}
	return set
}

// appendUnknownRef records a problem when ref is non-empty, the catalog is
// non-empty, and ref is absent from it; an empty catalog allows any reference.
func appendUnknownRef(problems []string, catalog map[string]bool, ref, kind, name, field string) []string {
	if ref == "" || len(catalog) == 0 || catalog[ref] {
		return problems
	}
	return append(problems, fmt.Sprintf("%s %q references unknown %s %q", kind, name, field, ref))
}
