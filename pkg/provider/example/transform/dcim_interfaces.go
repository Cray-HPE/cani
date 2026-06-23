package transform

import (
	"log"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// transformDcimInterfaces applies per-interface metadata (currently MAC
// addresses) from the DCIM CSV's `interface` rows onto the matching device
// or module interface specs. An interface row uses Device as the owner name
// and Name as the interface name. The owner is resolved first among objects
// created in this import and then among devices and modules already present in
// the inventory, so an interface row can annotate hardware imported in an
// earlier run. Rows referencing an unknown owner or interface are logged and
// skipped so a single bad row does not abort import.
func transformDcimInterfaces(data *import_.DcimCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory) error {
	for _, rec := range data.Interfaces {
		rec = data.ApplyDefaults(rec)

		if rec.Device == "" {
			log.Printf("WARN: DCIM CSV interface row for %q has no Device, skipping", rec.Name)
			continue
		}
		if rec.Name == "" {
			log.Printf("WARN: DCIM CSV interface row for device %q has no Name, skipping", rec.Device)
			continue
		}

		spec := findResultInterfaceSpec(result, rec.Device, rec.Name)
		if spec == nil {
			spec = findInventoryInterfaceSpec(inv, rec.Device, rec.Name)
		}
		if spec == nil {
			log.Printf("WARN: DCIM CSV interface %q not found on device/module %q, skipping", rec.Name, rec.Device)
			continue
		}

		if rec.MacAddress != "" {
			mac, err := devicetypes.NormalizeMAC(rec.MacAddress)
			if err != nil {
				log.Printf("WARN: DCIM CSV interface %q on %q: %v, skipping", rec.Name, rec.Device, err)
				continue
			}
			spec.MacAddress = mac
		}
	}
	return nil
}

// findResultInterfaceSpec locates an InterfaceSpec by interface name on the
// device or module with the given name within the transform result.
func findResultInterfaceSpec(result *devicetypes.TransformResult, ownerName, ifaceName string) *devicetypes.InterfaceSpec {
	if spec := findInterfaceSpecInDevices(result.Devices, ownerName, ifaceName); spec != nil {
		return spec
	}
	return findInterfaceSpecInModules(result.Modules, ownerName, ifaceName)
}

// findInventoryInterfaceSpec locates an InterfaceSpec by interface name on a
// device or module already present in the inventory, allowing an interface row
// to annotate hardware imported in an earlier run.
func findInventoryInterfaceSpec(inv *devicetypes.Inventory, ownerName, ifaceName string) *devicetypes.InterfaceSpec {
	if inv == nil {
		return nil
	}
	if spec := findInterfaceSpecInDevices(inv.Devices, ownerName, ifaceName); spec != nil {
		return spec
	}
	return findInterfaceSpecInModules(inv.Modules, ownerName, ifaceName)
}

// findInterfaceSpecInDevices returns the named interface spec on the device with
// the given name, or nil when no device or interface matches.
func findInterfaceSpecInDevices(devices map[uuid.UUID]*devicetypes.CaniDeviceType, ownerName, ifaceName string) *devicetypes.InterfaceSpec {
	for _, dev := range devices {
		if dev == nil || dev.Name != ownerName {
			continue
		}
		for i := range dev.Interfaces {
			if dev.Interfaces[i].Name == ifaceName {
				return &dev.Interfaces[i]
			}
		}
	}
	return nil
}

// findInterfaceSpecInModules returns the named interface spec on the module with
// the given name, or nil when no module or interface matches.
func findInterfaceSpecInModules(modules map[uuid.UUID]*devicetypes.CaniModuleType, ownerName, ifaceName string) *devicetypes.InterfaceSpec {
	for _, mod := range modules {
		if mod == nil || mod.Name != ownerName {
			continue
		}
		for i := range mod.Interfaces {
			if mod.Interfaces[i].Name == ifaceName {
				return &mod.Interfaces[i]
			}
		}
	}
	return nil
}
