package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/ochami/import"
	"github.com/google/uuid"
)

func TestNormaliseHardwareType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Rack", "rack"},
		{"Chassis", "chassis"},
		{"Blade", "blade"},
		{"Node", "node"},
		{"mgmt-switch", "mgmt-switch"},
		{"hsn-switch", "hsn-switch"},
		{"cabinet-pdu", "pdu"},
		{"cdu", "cdu"},
		{"CPU", "cpu"},
		{"DIMM", "memory"},
		{"GPU", "gpu"},
		{"NIC", "nic"},
		{"power-supply", "psu"},
		{"cable", "cable"},
		{"unknown-type", "unknown-type"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normaliseHardwareType(tt.input)
			if got != tt.want {
				t.Errorf("normaliseHardwareType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestClassifyRecords(t *testing.T) {
	records := []import_.JSONDeviceRecord{
		{DeviceType: "Rack", SerialNumber: "RACK-001"},
		{DeviceType: "Chassis", SerialNumber: "CHS-001", ParentSerialNumber: "RACK-001"},
		{DeviceType: "Node", SerialNumber: "NODE-001", ParentSerialNumber: "RACK-001"},
		{DeviceType: "mgmt-switch", SerialNumber: "SW-001", ParentSerialNumber: "RACK-001"},
		{DeviceType: "cabinet-pdu", SerialNumber: "PDU-001", ParentSerialNumber: "RACK-001"},
		{DeviceType: "CPU", SerialNumber: "CPU-001", ParentSerialNumber: "NODE-001"},
		{DeviceType: "DIMM", SerialNumber: "DIMM-001", ParentSerialNumber: "NODE-001"},
		{DeviceType: "cable", SerialNumber: "CBL-001"},
	}

	classified, err := classifyRecords(records)
	if err != nil {
		t.Fatalf("classifyRecords returned error: %v", err)
	}

	if len(classified.racks) != 1 {
		t.Errorf("racks = %d, want 1", len(classified.racks))
	}
	if len(classified.devices) != 6 {
		t.Errorf("devices = %d, want 6", len(classified.devices))
	}
	if len(classified.cables) != 1 {
		t.Errorf("cables = %d, want 1", len(classified.cables))
	}
}

func TestClassifyRecords_EmptyDeviceType(t *testing.T) {
	records := []import_.JSONDeviceRecord{
		{DeviceType: "", SerialNumber: "UNKNOWN-001"},
	}

	_, err := classifyRecords(records)
	if err == nil {
		t.Error("expected error for empty deviceType, got nil")
	}
}

func TestCreateRack(t *testing.T) {
	rec := import_.JSONDeviceRecord{
		DeviceType:   "Rack",
		SerialNumber: "RACK-SN-001",
		Manufacturer: "APC",
		PartNumber:   "AR3150",
		Properties:   import_.Properties{RedfishURI: "/Chassis/Rack001"},
	}

	rack := createRack(rec)

	if rack.Name != "RACK-SN-001" {
		t.Errorf("Name = %q, want %q", rack.Name, "RACK-SN-001")
	}
	if rack.Serial != "RACK-SN-001" {
		t.Errorf("Serial = %q, want %q", rack.Serial, "RACK-SN-001")
	}
	if rack.Manufacturer != "APC" {
		t.Errorf("Manufacturer = %q, want %q", rack.Manufacturer, "APC")
	}
	if rack.Status != "Active" {
		t.Errorf("Status = %q, want %q", rack.Status, "Active")
	}
	if rack.UHeight < 1 {
		t.Errorf("UHeight = %d, want >= 1", rack.UHeight)
	}
	if rack.ProviderMetadata == nil {
		t.Error("ProviderMetadata is nil")
	}
}

func TestCreateDevice(t *testing.T) {
	rec := import_.JSONDeviceRecord{
		DeviceType:         "Node",
		SerialNumber:       "NODE-SN-001",
		Manufacturer:       "HPE",
		PartNumber:         "P43357-B21",
		ParentSerialNumber: "RACK-SN-001",
		Properties:         import_.Properties{RedfishURI: "/Systems/NODE-SN-001"},
	}

	device := createDevice(rec)

	if device.Name != "NODE-SN-001" {
		t.Errorf("Name = %q, want %q", device.Name, "NODE-SN-001")
	}
	if device.Serial != "NODE-SN-001" {
		t.Errorf("Serial = %q, want %q", device.Serial, "NODE-SN-001")
	}
	if device.HardwareType != "node" {
		t.Errorf("HardwareType = %q, want %q", device.HardwareType, "node")
	}
	if device.Status != "Staged" {
		t.Errorf("Status = %q, want %q", device.Status, "Staged")
	}
}

func TestAssignParentRelationships_RackParent(t *testing.T) {
	inventory := devicetypes.NewInventory()

	rack := createRack(import_.JSONDeviceRecord{DeviceType: "Rack", SerialNumber: "RACK-001", Manufacturer: "APC"})
	device := createDevice(import_.JSONDeviceRecord{DeviceType: "Node", SerialNumber: "NODE-001", Manufacturer: "HPE", ParentSerialNumber: "RACK-001"})
	inventory.Racks[rack.ID] = rack
	inventory.Devices[device.ID] = device

	records := []import_.JSONDeviceRecord{
		{SerialNumber: "NODE-001", ParentSerialNumber: "RACK-001"},
	}

	serialToRackID := map[string]uuid.UUID{"RACK-001": rack.ID}
	serialToDeviceID := map[string]uuid.UUID{"NODE-001": device.ID}

	assignParentRelationships(inventory, records, serialToRackID, serialToDeviceID)

	if device.Parent != rack.ID {
		t.Errorf("device.Parent = %v, want rack.ID %v", device.Parent, rack.ID)
	}
	if device.Rack != rack.ID {
		t.Errorf("device.Rack = %v, want rack.ID %v", device.Rack, rack.ID)
	}
	if len(rack.Devices) != 1 || rack.Devices[0] != device.ID {
		t.Errorf("rack.Devices = %v, want [%v]", rack.Devices, device.ID)
	}
}

func TestAssignParentRelationships_DeviceParent(t *testing.T) {
	inventory := devicetypes.NewInventory()

	chassis := createDevice(import_.JSONDeviceRecord{DeviceType: "Chassis", SerialNumber: "CHS-001"})
	blade := createDevice(import_.JSONDeviceRecord{DeviceType: "Blade", SerialNumber: "BLD-001", ParentSerialNumber: "CHS-001"})
	inventory.Devices[chassis.ID] = chassis
	inventory.Devices[blade.ID] = blade

	records := []import_.JSONDeviceRecord{
		{SerialNumber: "BLD-001", ParentSerialNumber: "CHS-001"},
	}

	serialToRackID := map[string]uuid.UUID{}
	serialToDeviceID := map[string]uuid.UUID{
		"CHS-001": chassis.ID,
		"BLD-001": blade.ID,
	}

	assignParentRelationships(inventory, records, serialToRackID, serialToDeviceID)

	if blade.Parent != chassis.ID {
		t.Errorf("blade.Parent = %v, want chassis.ID %v", blade.Parent, chassis.ID)
	}
	if blade.ParentDevice != chassis.ID {
		t.Errorf("blade.ParentDevice = %v, want chassis.ID %v", blade.ParentDevice, chassis.ID)
	}
	if len(chassis.Children) != 1 || chassis.Children[0] != blade.ID {
		t.Errorf("chassis.Children = %v, want [%v]", chassis.Children, blade.ID)
	}
}

func TestCreateCable(t *testing.T) {
	rec := import_.JSONDeviceRecord{
		DeviceType:   "cable",
		SerialNumber: "CBL-SN-001",
		Manufacturer: "Mellanox",
		PartNumber:   "MCP1600-C003E30N",
	}

	cable := createCable(rec)

	if cable.Label != "CBL-SN-001" {
		t.Errorf("Label = %q, want %q", cable.Label, "CBL-SN-001")
	}
	if cable.Manufacturer != "Mellanox" {
		t.Errorf("Manufacturer = %q, want %q", cable.Manufacturer, "Mellanox")
	}
	if cable.PartNumber != "MCP1600-C003E30N" {
		t.Errorf("PartNumber = %q, want %q", cable.PartNumber, "MCP1600-C003E30N")
	}
	if cable.Slug == "" {
		t.Error("Slug must not be empty")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Rack", "rack"},
		{"mgmt-switch", "mgmt-switch"},
		{"Hello World", "hello-world"},
		{"CPU", "cpu"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
