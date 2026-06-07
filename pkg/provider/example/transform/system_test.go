package transform

import (
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

func TestTransformSystem_Roles(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Roles: []import_.SystemRecord{
			{Section: "role", Name: "ComputeNode", ContentTypes: "dcim.device"},
			{Section: "role", Name: "Gateway", ContentTypes: "dcim.device,dcim.rack"},
		},
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metadata == nil {
		t.Fatal("expected Metadata to be set")
	}
	if len(result.Metadata.Roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(result.Metadata.Roles))
	}
	if result.Metadata.Roles[0].Name != "ComputeNode" {
		t.Errorf("first role = %q, want %q", result.Metadata.Roles[0].Name, "ComputeNode")
	}
	if len(result.Metadata.Roles[1].ContentTypes) != 2 {
		t.Errorf("Gateway should have 2 content types, got %d", len(result.Metadata.Roles[1].ContentTypes))
	}
}

func TestTransformSystem_Racks(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3701", Qty: 1, Status: "Available"},
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3702", Qty: 1, Status: "Available"},
		},
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Racks) != 2 {
		t.Fatalf("expected 2 racks, got %d", len(result.Racks))
	}

	names := make(map[string]bool)
	for _, rack := range result.Racks {
		names[rack.Name] = true
		if rack.Status != "Available" {
			t.Errorf("rack %q Status = %q, want %q", rack.Name, rack.Status, "Available")
		}
		if rack.UHeight < 1 {
			t.Errorf("rack %q UHeight = %d, want > 0", rack.Name, rack.UHeight)
		}
	}
	if !names["x3701"] || !names["x3702"] {
		t.Errorf("expected racks x3701 and x3702, got %v", names)
	}
}

func TestTransformSystem_Devices(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3701", Qty: 1, Status: "Available"},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "GH-x3701u34", Qty: 1, Rack: "x3701", Position: 34, Face: "front", Role: "ComputeNode", Status: "Active", Serial: "ABC123"},
		},
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(result.Devices))
	}

	for _, dev := range result.Devices {
		if dev.Name != "GH-x3701u34" {
			t.Errorf("Name = %q, want %q", dev.Name, "GH-x3701u34")
		}
		if dev.Role != "ComputeNode" {
			t.Errorf("Role = %q, want %q", dev.Role, "ComputeNode")
		}
		if dev.Status != "Active" {
			t.Errorf("Status = %q, want %q", dev.Status, "Active")
		}
		if dev.Serial != "ABC123" {
			t.Errorf("Serial = %q, want %q", dev.Serial, "ABC123")
		}
		if dev.RackPosition != 34 {
			t.Errorf("RackPosition = %d, want %d", dev.RackPosition, 34)
		}
		if dev.Face != "front" {
			t.Errorf("Face = %q, want %q", dev.Face, "front")
		}
		if dev.Parent == uuid.Nil {
			t.Error("expected Parent to be set to rack UUID")
		}
	}
}

func TestTransformSystem_DeviceUnknownRack(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "node1", Qty: 1, Rack: "nonexistent"},
		},
	}

	_, err := TransformSystem(devicetypes.Inventory{}, data)
	if err == nil {
		t.Fatal("expected error for unknown rack reference")
	}
}

func TestTransformSystem_InterfaceMAC(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3701", Qty: 1, Status: "Available"},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "GH-x3701u34", Qty: 1, Rack: "x3701", Position: 34, Face: "front", Role: "ComputeNode", Status: "Active"},
			{Section: "device", PartNumber: "hpe-xd670", Name: "GH-x3701u26", Qty: 1, Rack: "x3701", Position: 26, Face: "front", Role: "ComputeNode", Status: "Active"},
		},
		Interfaces: []import_.SystemRecord{
			// Hyphen form must be normalized to lowercase colon form.
			{Section: "interface", Device: "GH-x3701u34", Name: "iLO", MacAddress: "AA-BB-CC-DD-EE-01"},
		},
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	macFor := func(devName string) string {
		for _, d := range result.Devices {
			if d.Name != devName {
				continue
			}
			for _, ifc := range d.Interfaces {
				if ifc.Name == "iLO" {
					return ifc.MacAddress
				}
			}
		}
		return "<not found>"
	}

	if got := macFor("GH-x3701u34"); got != "aa:bb:cc:dd:ee:01" {
		t.Errorf("GH-x3701u34 iLO MAC = %q, want %q", got, "aa:bb:cc:dd:ee:01")
	}

	// Slice-clone isolation: setting MAC on one device of a type must not
	// leak into another device sharing the same device-type template.
	if got := macFor("GH-x3701u26"); got != "" {
		t.Errorf("GH-x3701u26 iLO MAC = %q, want empty (per-device isolation)", got)
	}
}

func TestTransformSystem_InterfaceMACUnknownTarget(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3701", Qty: 1, Status: "Available"},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "GH-x3701u34", Qty: 1, Rack: "x3701", Position: 34, Face: "front", Role: "ComputeNode", Status: "Active"},
		},
		Interfaces: []import_.SystemRecord{
			// Unknown interface name -> warn + skip, not a hard error.
			{Section: "interface", Device: "GH-x3701u34", Name: "does-not-exist", MacAddress: "aa:bb:cc:dd:ee:ff"},
		},
	}

	if _, err := TransformSystem(devicetypes.Inventory{}, data); err != nil {
		t.Fatalf("unknown interface target should be skipped, got error: %v", err)
	}
}

func TestTransformSystem_Modules(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3701", Qty: 1, Status: "Available"},
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3507", Qty: 1, Status: "Available"},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "GH-x3701u34", Qty: 1, Rack: "x3701", Position: 34, Face: "front"},
			{Section: "device", PartNumber: "hpe-proliant-dl380-gen11-8sff", Name: "SERV-x3507u21", Qty: 1, Rack: "x3507", Position: 21, Face: "front"},
		},
		Modules: []import_.SystemRecord{
			{Section: "module", PartNumber: "nvidia-h100-sxm-gpu", Qty: 1, Device: "GH-x3701u34", Bay: "GPU0"},
			{Section: "module", PartNumber: "nvidia-connectx-6-dx-100gbe-2p-qsfp28", Qty: 1, Device: "SERV-x3507u21", Bay: "PCIe5"},
		},
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(result.Modules))
	}

	modsByDevice := make(map[uuid.UUID]*devicetypes.CaniModuleType)
	for _, mod := range result.Modules {
		if mod.ParentDevice == uuid.Nil {
			t.Error("expected ParentDevice to be set")
		}
		modsByDevice[mod.ParentDevice] = mod
	}

	deviceIDs := make(map[string]uuid.UUID)
	for id, dev := range result.Devices {
		deviceIDs[dev.Name] = id
	}

	gpuMod := modsByDevice[deviceIDs["GH-x3701u34"]]
	if gpuMod == nil {
		t.Fatal("expected GPU module for GH-x3701u34")
	}
	if gpuMod.ModuleBayName != "GPU 0" {
		t.Errorf("GPU ModuleBayName = %q, want %q", gpuMod.ModuleBayName, "GPU 0")
	}
	if gpuMod.Name != "gpu-GH-x3701u34-GPU 0" {
		t.Errorf("GPU Name = %q, want %q", gpuMod.Name, "gpu-GH-x3701u34-GPU 0")
	}

	nicMod := modsByDevice[deviceIDs["SERV-x3507u21"]]
	if nicMod == nil {
		t.Fatal("expected ConnectX-6 module for SERV-x3507u21")
	}
	if nicMod.ModuleBayName != "PCIe5" {
		t.Errorf("NIC ModuleBayName = %q, want %q", nicMod.ModuleBayName, "PCIe5")
	}
	if nicMod.Name != "CX6-SERV-x3507u21" {
		t.Errorf("NIC Name = %q, want %q", nicMod.Name, "CX6-SERV-x3507u21")
	}
}

func TestTransformSystem_ManualParity(t *testing.T) {
	data := &import_.SystemCSV{
		Defaults:        import_.SystemRecord{Section: "_defaults", Status: "Active"},
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3701", Qty: 1, Status: "Available"},
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3507", Qty: 1, Status: "Available"},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "GH-x3701u34", Qty: 1, Rack: "x3701", Position: 34, Face: "front", Role: "ComputeNode"},
			{Section: "device", PartNumber: "hpe-proliant-dl380-gen11-8sff", Name: "SERV-x3507u21", Qty: 1, Rack: "x3507", Position: 21, Face: "front", Role: "ServiceNode"},
		},
		Modules: []import_.SystemRecord{
			{Section: "module", PartNumber: "nvidia-h100-sxm-gpu", Qty: 1, Device: "GH-x3701u34", Bay: "GPU0"},
			{Section: "module", PartNumber: "nvidia-connectx-6-dx-100gbe-2p-qsfp28", Qty: 1, Device: "SERV-x3507u21", Bay: "PCIe5"},
		},
	}

	imported, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected import error: %v", err)
	}

	manual := buildManualParityResult(t)

	importedSnapshot := normalizeSystemParity(imported)
	manualSnapshot := normalizeSystemParity(manual)
	if !reflect.DeepEqual(importedSnapshot, manualSnapshot) {
		t.Fatalf("import/manual parity mismatch\nimported: %#v\nmanual: %#v", importedSnapshot, manualSnapshot)
	}
}

type systemParitySnapshot struct {
	Racks   map[string]rackParitySnapshot
	Devices map[string]deviceParitySnapshot
	Modules map[string]moduleParitySnapshot
}

type rackParitySnapshot struct {
	Slug    string
	Status  string
	UHeight int
}

type deviceParitySnapshot struct {
	Slug        string
	Rack        string
	RackPos     int
	Face        string
	Role        string
	Status      string
	UHeight     int
	IsFullDepth bool
}

type moduleParitySnapshot struct {
	Slug         string
	ParentDevice string
	ModuleBay    string
	Status       string
}

func buildManualParityResult(t *testing.T) *devicetypes.TransformResult {
	t.Helper()

	rack3701, err := devicetypes.NewRackFromSlug("hpe-48u-800mmx1200mm-g2-enterprise-shock-rack")
	if err != nil {
		t.Fatalf("NewRackFromSlug x3701: %v", err)
	}
	rack3701.Name = "x3701"
	rack3701.Status = "Available"

	rack3507, err := devicetypes.NewRackFromSlug("hpe-48u-800mmx1200mm-g2-enterprise-shock-rack")
	if err != nil {
		t.Fatalf("NewRackFromSlug x3507: %v", err)
	}
	rack3507.Name = "x3507"
	rack3507.Status = "Available"

	deviceGPU, err := devicetypes.NewDeviceFromSlug("hpe-xd670")
	if err != nil {
		t.Fatalf("NewDeviceFromSlug hpe-xd670: %v", err)
	}
	deviceGPU.Name = "GH-x3701u34"
	deviceGPU.Status = "Active"
	deviceGPU.Role = "ComputeNode"
	deviceGPU.Rack = rack3701.ID
	deviceGPU.Parent = rack3701.ID
	deviceGPU.RackPosition = 34
	deviceGPU.Face = devicetypes.FaceFront
	rack3701.Devices = append(rack3701.Devices, deviceGPU.ID)
	rack3701.PlaceDevice(deviceGPU.ID, deviceGPU.RackPosition, deviceGPU.UHeight, deviceGPU.Face, deviceGPU.IsFullDepth)

	deviceService, err := devicetypes.NewDeviceFromSlug("hpe-proliant-dl380-gen11-8sff")
	if err != nil {
		t.Fatalf("NewDeviceFromSlug hpe-proliant-dl380-gen11-8sff: %v", err)
	}
	deviceService.Name = "SERV-x3507u21"
	deviceService.Status = "Active"
	deviceService.Role = "ServiceNode"
	deviceService.Rack = rack3507.ID
	deviceService.Parent = rack3507.ID
	deviceService.RackPosition = 21
	deviceService.Face = devicetypes.FaceFront
	rack3507.Devices = append(rack3507.Devices, deviceService.ID)
	rack3507.PlaceDevice(deviceService.ID, deviceService.RackPosition, deviceService.UHeight, deviceService.Face, deviceService.IsFullDepth)

	gpuModule, err := devicetypes.NewModuleFromSlug("nvidia-h100-sxm-gpu")
	if err != nil {
		t.Fatalf("NewModuleFromSlug nvidia-h100-sxm-gpu: %v", err)
	}
	gpuModule.Name = "gpu-GH-x3701u34-GPU 0"
	gpuModule.ParentDevice = deviceGPU.ID
	gpuModule.ModuleBayName = "GPU 0"
	gpuModule.Status = "Active"

	nicModule, err := devicetypes.NewModuleFromSlug("nvidia-connectx-6-dx-100gbe-2p-qsfp28")
	if err != nil {
		t.Fatalf("NewModuleFromSlug nvidia-connectx-6-dx-100gbe-2p-qsfp28: %v", err)
	}
	nicModule.Name = "CX6-SERV-x3507u21"
	nicModule.ParentDevice = deviceService.ID
	nicModule.ModuleBayName = "PCIe5"
	nicModule.Status = "Active"

	return &devicetypes.TransformResult{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rack3701.ID: rack3701,
			rack3507.ID: rack3507,
		},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceGPU.ID:     deviceGPU,
			deviceService.ID: deviceService,
		},
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			gpuModule.ID: gpuModule,
			nicModule.ID: nicModule,
		},
	}
}

func normalizeSystemParity(result *devicetypes.TransformResult) systemParitySnapshot {
	snapshot := systemParitySnapshot{
		Racks:   make(map[string]rackParitySnapshot),
		Devices: make(map[string]deviceParitySnapshot),
		Modules: make(map[string]moduleParitySnapshot),
	}

	rackNames := make(map[uuid.UUID]string, len(result.Racks))
	for id, rack := range result.Racks {
		if rack == nil {
			continue
		}
		rackNames[id] = rack.Name
		snapshot.Racks[rack.Name] = rackParitySnapshot{
			Slug:    rack.Slug,
			Status:  rack.Status,
			UHeight: rack.UHeight,
		}
	}

	deviceNames := make(map[uuid.UUID]string, len(result.Devices))
	for id, dev := range result.Devices {
		if dev == nil {
			continue
		}
		deviceNames[id] = dev.Name
		snapshot.Devices[dev.Name] = deviceParitySnapshot{
			Slug:        dev.Slug,
			Rack:        rackNames[dev.Rack],
			RackPos:     dev.RackPosition,
			Face:        dev.Face,
			Role:        dev.Role,
			Status:      dev.Status,
			UHeight:     dev.UHeight,
			IsFullDepth: dev.IsFullDepth,
		}
	}

	for _, mod := range result.Modules {
		if mod == nil {
			continue
		}
		snapshot.Modules[mod.Name] = moduleParitySnapshot{
			Slug:         mod.Slug,
			ParentDevice: deviceNames[mod.ParentDevice],
			ModuleBay:    mod.ModuleBayName,
			Status:       mod.Status,
		}
	}

	return snapshot
}

func TestTransformSystem_Defaults(t *testing.T) {
	data := &import_.SystemCSV{
		Defaults:        import_.SystemRecord{Section: "_defaults", Status: "Active"},
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3701", Qty: 1},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "node1", Qty: 1, Rack: "x3701", Position: 10, Face: "front"},
		},
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Rack should inherit Active status from defaults
	for _, rack := range result.Racks {
		if rack.Status != "Active" {
			t.Errorf("rack Status = %q, want %q (from defaults)", rack.Status, "Active")
		}
	}

	// Device should inherit Active status from defaults
	for _, dev := range result.Devices {
		if dev.Status != "Active" {
			t.Errorf("device Status = %q, want %q (from defaults)", dev.Status, "Active")
		}
	}
}

func TestTransformSystem_RackQtyMultiple(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "rack", Qty: 3, Status: "Available"},
		},
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Racks) != 3 {
		t.Fatalf("expected 3 racks, got %d", len(result.Racks))
	}

	names := make(map[string]bool)
	for _, rack := range result.Racks {
		names[rack.Name] = true
	}
	for _, expected := range []string{"rack-1", "rack-2", "rack-3"} {
		if !names[expected] {
			t.Errorf("expected rack %q, not found in %v", expected, names)
		}
	}
}

func TestTransformSystem_EmptyData(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Racks) != 0 {
		t.Errorf("expected 0 racks, got %d", len(result.Racks))
	}
	if len(result.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(result.Devices))
	}
}
