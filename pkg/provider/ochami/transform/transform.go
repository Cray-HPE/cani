package transform

import (
	"fmt"
	"log"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ochami/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/ochami/import"
	"github.com/google/uuid"
)

// providerGetter returns the Ochami singleton with raw records.
// Set by the parent package to break import cycles.
var providerGetter func() interface {
	GetRecords() []import_.JSONDeviceRecord
}

// SetProviderGetter allows the parent package to provide singleton access.
func SetProviderGetter(getter func() interface {
	GetRecords() []import_.JSONDeviceRecord
}) {
	providerGetter = getter
}

// hardwareTypeMap maps ochami deviceType strings to CANI canonical hardware types.
var hardwareTypeMap = map[string]string{
	"rack":          "rack",
	"chassis":       "chassis",
	"blade":         "blade",
	"node":          "node",
	"mgmt-switch":   "mgmt-switch",
	"hsn-switch":    "hsn-switch",
	"cabinet-pdu":   "pdu",
	"cdu":           "cdu",
	"cpu":           "cpu",
	"dimm":          "memory",
	"gpu":           "gpu",
	"nic":           "nic",
	"power-supply":  "psu",
	"cable":         "cable",
}

// normaliseHardwareType maps an ochami deviceType string to a CANI canonical hardware type.
func normaliseHardwareType(deviceType string) string {
	if canonical, ok := hardwareTypeMap[strings.ToLower(deviceType)]; ok {
		return canonical
	}
	return strings.ToLower(deviceType)
}

type classifiedRecords struct {
	racks   []import_.JSONDeviceRecord
	devices []import_.JSONDeviceRecord
	cables  []import_.JSONDeviceRecord
}

// classifyRecords categorises records into racks, devices, and cables.
func classifyRecords(records []import_.JSONDeviceRecord) (*classifiedRecords, error) {
	result := &classifiedRecords{}

	for i, rec := range records {
		hwType := normaliseHardwareType(rec.DeviceType)
		switch hwType {
		case "rack":
			result.racks = append(result.racks, rec)
		case "cable":
			result.cables = append(result.cables, rec)
		case "":
			return nil, fmt.Errorf("record %d: cannot classify hardware type for %q", i+1, rec.DeviceType)
		default:
			result.devices = append(result.devices, rec)
		}
	}

	return result, nil
}

// Transform converts imported data into CANI's inventory format.
// Returns a TransformResult containing devices, racks, and cables.
func Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	if providerGetter == nil {
		return nil, fmt.Errorf("transform: provider getter not set")
	}

	prov := providerGetter()
	allRecords := prov.GetRecords()

	classified, err := classifyRecords(allRecords)
	if err != nil {
		return nil, fmt.Errorf("classify records: %w", err)
	}

	initInventoryMaps(&existing)

	serialToRackID := make(map[string]uuid.UUID)
	serialToDeviceID := make(map[string]uuid.UUID)
	createdRacks := make(map[uuid.UUID]*devicetypes.CaniRackType)
	createdDevices := make(map[uuid.UUID]*devicetypes.CaniDeviceType)
	createdCables := make(map[uuid.UUID]*devicetypes.CaniCableType)

	// Pass 1: create racks
	for _, rec := range classified.racks {
		rack := createRack(rec)
		existing.Racks[rack.ID] = rack
		createdRacks[rack.ID] = rack
		serialToRackID[rec.SerialNumber] = rack.ID
	}

	// Pass 2: create devices
	for _, rec := range classified.devices {
		device := createDevice(rec)
		existing.Devices[device.ID] = device
		createdDevices[device.ID] = device
		serialToDeviceID[rec.SerialNumber] = device.ID
	}

	// Pass 3: assign parent-child relationships via ParentSerialNumber
	assignParentRelationships(&existing, classified.devices, serialToRackID, serialToDeviceID)

	// Pass 4: create cables
	for _, rec := range classified.cables {
		cable := createCable(rec)
		existing.Cables[cable.ID] = cable
		createdCables[cable.ID] = cable
	}

	log.Printf("Transformed: %d racks, %d devices, %d cables",
		len(createdRacks), len(createdDevices), len(createdCables))

	return &devicetypes.TransformResult{
		Racks:   createdRacks,
		Devices: createdDevices,
		Cables:  createdCables,
	}, nil
}

// createRack builds a CaniRackType from a JSONDeviceRecord.
func createRack(rec import_.JSONDeviceRecord) *devicetypes.CaniRackType {
	id := uuid.New()
	slug := slugify(rec.DeviceType)
	uHeight := 42

	if rt, ok := devicetypes.GetRackTypeByPartNumber(rec.PartNumber); ok {
		if rt.Slug != "" {
			slug = rt.Slug
		}
		if rt.UHeight > 0 {
			uHeight = rt.UHeight
		}
	}

	return &devicetypes.CaniRackType{
		ID:               id,
		Name:             rec.SerialNumber,
		Slug:             slug,
		Serial:           rec.SerialNumber,
		Manufacturer:     rec.Manufacturer,
		PartNumber:       rec.PartNumber,
		UHeight:          uHeight,
		Status:           "active",
		Devices:          []uuid.UUID{},
		ProviderMetadata: buildProviderMetadata(rec),
	}
}

// createDevice builds a CaniDeviceType from a JSONDeviceRecord.
func createDevice(rec import_.JSONDeviceRecord) *devicetypes.CaniDeviceType {
	id := uuid.New()
	hwType := normaliseHardwareType(rec.DeviceType)

	device := &devicetypes.CaniDeviceType{
		ID:               id,
		Name:             rec.SerialNumber,
		Slug:             slugify(rec.DeviceType),
		Serial:           rec.SerialNumber,
		Manufacturer:     rec.Manufacturer,
		PartNumber:       rec.PartNumber,
		HardwareType:     hwType,
		Status:           "staged",
		ProviderMetadata: buildProviderMetadata(rec),
	}

	if dt, ok := devicetypes.GetByPartNumber(rec.PartNumber); ok {
		populateFromDeviceType(device, &dt)
	}

	return device
}

// populateFromDeviceType copies library fields into an inventory device instance.
func populateFromDeviceType(device *devicetypes.CaniDeviceType, dt *devicetypes.CaniDeviceType) {
	device.Slug = dt.Slug
	device.Manufacturer = dt.Manufacturer
	device.Model = dt.Model
	if dt.HardwareType != "" {
		device.HardwareType = dt.HardwareType
	}
	device.Interfaces = dt.Interfaces
}

// buildProviderMetadata creates ochami-keyed provider metadata from a record.
func buildProviderMetadata(rec import_.JSONDeviceRecord) map[string]any {
	return map[string]any{
		"ochami": map[string]any{
			"Source":             commands.JsonFileFlag,
			"SerialNumber":       rec.SerialNumber,
			"ParentSerialNumber": rec.ParentSerialNumber,
			"RedfishURI":         rec.Properties.RedfishURI,
		},
	}
}

// assignParentRelationships links devices to their parents via ParentSerialNumber.
// Rack-parented devices get Rack set and are added to the rack's Devices list.
// Device-parented children get ParentDevice set and are appended to parent's Children.
func assignParentRelationships(
	inventory *devicetypes.Inventory,
	deviceRecords []import_.JSONDeviceRecord,
	serialToRackID map[string]uuid.UUID,
	serialToDeviceID map[string]uuid.UUID,
) {
	for _, rec := range deviceRecords {
		if rec.ParentSerialNumber == "" {
			continue
		}

		deviceID, ok := serialToDeviceID[rec.SerialNumber]
		if !ok {
			continue
		}
		device := inventory.Devices[deviceID]
		if device == nil {
			continue
		}

		if rackID, ok := serialToRackID[rec.ParentSerialNumber]; ok {
			rack := inventory.Racks[rackID]
			if rack == nil {
				continue
			}
			device.Parent = rackID
			device.Rack = rackID
			rack.Devices = append(rack.Devices, deviceID)
			continue
		}

		if parentID, ok := serialToDeviceID[rec.ParentSerialNumber]; ok {
			parent := inventory.Devices[parentID]
			if parent == nil {
				continue
			}
			device.Parent = parentID
			device.ParentDevice = parentID
			parent.Children = append(parent.Children, deviceID)
		}
	}
}

// initInventoryMaps ensures inventory maps are initialized.
func initInventoryMaps(inventory *devicetypes.Inventory) {
	if inventory.Racks == nil {
		inventory.Racks = make(map[uuid.UUID]*devicetypes.CaniRackType)
	}
	if inventory.Devices == nil {
		inventory.Devices = make(map[uuid.UUID]*devicetypes.CaniDeviceType)
	}
	if inventory.Cables == nil {
		inventory.Cables = make(map[uuid.UUID]*devicetypes.CaniCableType)
	}
}

// slugify converts a string to a URL-safe slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
