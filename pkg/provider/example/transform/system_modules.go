package transform

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// transformSystemModules creates modules from system CSV module records.
func transformSystemModules(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory) error {
	for _, rec := range data.Modules {
		rec = data.ApplyDefaults(rec)

		for i := 0; i < rec.Qty; i++ {
			id := uuid.New()
			var parentDevice *devicetypes.CaniDeviceType

			module := &devicetypes.CaniModuleType{
				ID:         id,
				Name:       rec.Name,
				Slug:       rec.PartNumber,
				PartNumber: rec.PartNumber,
				Serial:     rec.Serial,
				ObjectMeta: devicetypes.ObjectMeta{Status: rec.Status},
			}

			// Populate from module type library
			if rec.PartNumber != "" {
				if mt, ok := devicetypes.GetModuleBySlug(rec.PartNumber); ok {
					populateModuleFromType(module, &mt)
				} else if mt, ok := devicetypes.GetModuleTypeByPartNumber(rec.PartNumber); ok {
					populateModuleFromType(module, &mt)
				}
			}

			// Parent to device
			if rec.Device != "" {
				dev := inv.FindDeviceByNameOrID(rec.Device)
				if dev == nil {
					return fmt.Errorf("module %q references unknown device %q", rec.Name, rec.Device)
				}
				parentDevice = dev
				module.ParentDevice = dev.ID
			}

			if rec.Bay != "" {
				module.ModuleBayName = resolveSystemModuleBayName(parentDevice, rec.Bay)
			}

			if module.Name == "" {
				module.Name = synthesizeSystemModuleName(module, parentDevice)
			}

			result.Modules[id] = module
		}
	}

	return nil
}

// populateModuleFromType copies identity and hardware specs from a module
// type library entry, including a fresh copy of the interface specs so that
// per-interface details (e.g. MAC addresses) can be set on the module.
func populateModuleFromType(module *devicetypes.CaniModuleType, mt *devicetypes.CaniModuleType) {
	module.Slug = mt.Slug
	module.Manufacturer = mt.Manufacturer
	module.Model = mt.Model
	module.Type = mt.Type
	if mt.PartNumber != "" {
		module.PartNumber = mt.PartNumber
	}
	module.Interfaces = append([]devicetypes.InterfaceSpec(nil), mt.Interfaces...)
}

func resolveSystemModuleBayName(device *devicetypes.CaniDeviceType, requestedBay string) string {
	if device == nil || requestedBay == "" {
		return requestedBay
	}

	for _, bay := range device.ModuleBays {
		if bay.Name == requestedBay || bay.Position == requestedBay {
			return bay.Name
		}
	}

	return requestedBay
}

func synthesizeSystemModuleName(module *devicetypes.CaniModuleType, device *devicetypes.CaniDeviceType) string {
	if module == nil || module.Name != "" {
		return ""
	}

	lowerSlug := strings.ToLower(module.Slug)
	if device != nil && module.ModuleBayName != "" && (module.Type == devicetypes.TypeGPU || strings.Contains(lowerSlug, "gpu")) {
		return fmt.Sprintf("gpu-%s-%s", device.Name, module.ModuleBayName)
	}

	if device != nil && strings.Contains(lowerSlug, "connectx-6") {
		return fmt.Sprintf("CX6-%s", device.Name)
	}

	return fallbackSystemModuleName(module, device)
}

// fallbackSystemModuleName builds a deterministic name from the module slug,
// parent device, and bay so a module is never left unnamed. Unnamed modules are
// dropped from datastore summaries and collide when a row sets Qty > 1.
func fallbackSystemModuleName(module *devicetypes.CaniModuleType, device *devicetypes.CaniDeviceType) string {
	base := module.Slug
	if base == "" {
		base = "module"
	}
	if device != nil {
		base = fmt.Sprintf("%s-%s", base, device.Name)
	}
	if module.ModuleBayName != "" {
		base = fmt.Sprintf("%s-%s", base, module.ModuleBayName)
	}
	return base
}
