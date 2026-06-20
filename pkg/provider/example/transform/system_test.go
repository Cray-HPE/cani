package transform

import (
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// TestTransformSystem_Roles verifies the role pass populates result metadata
// with named roles and their parsed content types.
//
// Why it matters: roles are referenced by devices later in the import, so the
// importer must register them up front with their content-type associations
// intact.
// Inputs: a SystemCSV with two role rows, one single- and one multi-content-type.
// Outputs: result.Metadata.Roles.
// Data choice: ComputeNode (one content type) and Gateway
// ("dcim.device,dcim.rack") prove both name capture and the comma-split of
// content types.
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

// TestTransformSystem_Statuses verifies the metadata pass registers statuses
// from the `status` section, including when no roles are present.
//
// Why it matters: statuses are a first-class Nautobot catalog with content
// types, so the importer must register them up front even for a status-only
// file, mirroring how roles are handled.
// Inputs: a SystemCSV with two status rows and no roles. Outputs:
// result.Metadata.Statuses with parsed content types and empty Roles.
// Data choice: Active (one content type) and Planned ("dcim.device,dcim.rack")
// prove name capture and content-type splitting, while omitting roles proves the
// metadata is built from statuses alone.
func TestTransformSystem_Statuses(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Statuses: []import_.SystemRecord{
			{Section: "status", Name: "Active", ContentTypes: "dcim.device"},
			{Section: "status", Name: "Planned", ContentTypes: "dcim.device,dcim.rack"},
		},
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metadata == nil {
		t.Fatal("expected Metadata to be set from statuses alone")
	}
	if len(result.Metadata.Roles) != 0 {
		t.Errorf("expected 0 roles, got %d", len(result.Metadata.Roles))
	}
	if len(result.Metadata.Statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(result.Metadata.Statuses))
	}
	if result.Metadata.Statuses[0].Name != "Active" {
		t.Errorf("first status = %q, want %q", result.Metadata.Statuses[0].Name, "Active")
	}
	if len(result.Metadata.Statuses[1].ContentTypes) != 2 {
		t.Errorf("Planned should have 2 content types, got %d", len(result.Metadata.Statuses[1].ContentTypes))
	}
}

// TestTransformSystem_MetadataMissingName verifies a catalog row without a Name
// fails the metadata pass with a kind-specific error.
//
// Why it matters: roles and statuses are keyed by Name, so a nameless catalog
// row cannot be registered and must abort the import with a clear message naming
// the offending section kind.
// Inputs: a SystemCSV with one status row missing its Name. Outputs: a non-nil
// error equal to the wrapped "status record missing Name" message.
// Data choice: an empty-Name status isolates the shared missing-name guard and
// proves the kind label is threaded into the message.
func TestTransformSystem_MetadataMissingName(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Statuses: []import_.SystemRecord{
			{Section: "status", ContentTypes: "dcim.device"},
		},
	}
	_, err := TransformSystem(devicetypes.Inventory{}, data)
	if err == nil || err.Error() != "transformMetadata: status record missing Name" {
		t.Fatalf("error = %v, want 'transformMetadata: status record missing Name'", err)
	}
}

// TestTransformSystem_Racks verifies the rack pass creates one rack per row with
// the library U-height and the row's status.
//
// Why it matters: racks are the parents for device placement, so each row must
// resolve to a rack type sized from the library and carry its declared status.
// Inputs: two rack rows sharing a real rack slug. Outputs: result.Racks with
// names, statuses, and U-heights set.
// Data choice: two distinct names on the same 48U slug prove independent racks
// are created and the library U-height is applied to each.
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

// TestTransformSystem_Devices verifies the device pass creates a device parented
// to its rack with role, status, serial, position, and face copied from the row.
//
// Why it matters: devices are the core inventory objects, so the importer must
// place each one in its rack and preserve its operational fields.
// Inputs: a rack plus one fully specified hpe-xd670 device row. Outputs:
// result.Devices with placement and metadata set.
// Data choice: a real device part number at U34 front with a serial proves
// library typing plus a faithful copy of every per-device field and the rack
// parent link.
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

// TestTransformSystem_DeviceUnknownRack verifies the device pass errors when a
// device references a rack that does not exist.
//
// Why it matters: a device must land in a real rack, so a dangling rack
// reference aborts the import rather than orphaning the device.
// Inputs: a device row naming a rack absent from the batch and inventory.
// Outputs: a non-nil error.
// Data choice: omitting the rack section entirely guarantees the reference is
// unresolved, isolating the unknown-rack guard.
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

// TestTransformSystem_InterfaceMAC verifies an interface row's MAC is normalized
// to lowercase colon form and applied to only the named device.
//
// Why it matters: per-interface metadata personalizes a shared device-type
// template, so a MAC set on one device must normalize and must not leak to
// siblings of the same type.
// Inputs: two hpe-xd670 devices and one interface row setting iLO MAC
// "AA-BB-CC-DD-EE-01" on one. Outputs: the normalized MAC on that device's iLO,
// empty on the other.
// Data choice: the hyphen-uppercase input proves normalization; a second
// identical-type device proves the interface slice is cloned per device.
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

// TestTransformSystem_InterfaceMACUnknownTarget verifies an interface row naming
// a non-existent interface is skipped with a warning rather than failing.
//
// Why it matters: a single bad interface row should not abort an otherwise valid
// import, so an unknown interface name is tolerated and skipped.
// Inputs: a device plus an interface row with an unknown Name. Outputs: no error
// from TransformSystem.
// Data choice: a valid device with one unmatched interface name isolates the
// skip path without any other failure source.
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

// TestTransformSystem_InterfaceMACExistingDevice verifies an interface row
// annotates a device that already exists in the inventory, not only devices
// created in the same import.
//
// Why it matters: operators run incremental imports, so a later interface-only
// file must set MAC addresses on hardware imported in an earlier run rather than
// silently skip it.
// Inputs: an inventory pre-seeded with device "node-existing" (interface iLO, no
// MAC) and a SystemCSV carrying only an interface row for that device. Outputs:
// the seeded device's iLO interface gains the normalized MAC.
// Data choice: the device lives only in the inventory (absent from the batch),
// so a set MAC can come only from the inventory fallback; the hyphen-uppercase
// input also proves normalization.
func TestTransformSystem_InterfaceMACExistingDevice(t *testing.T) {
	inv := *devicetypes.NewInventory()
	devID := uuid.New()
	inv.Devices[devID] = &devicetypes.CaniDeviceType{
		ID:         devID,
		Name:       "node-existing",
		Interfaces: []devicetypes.InterfaceSpec{{Name: "iLO"}},
	}

	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Interfaces: []import_.SystemRecord{
			{Section: "interface", Device: "node-existing", Name: "iLO", MacAddress: "AA-BB-CC-DD-EE-09"},
		},
	}

	if _, err := TransformSystem(inv, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := inv.Devices[devID].Interfaces[0].MacAddress; got != "aa:bb:cc:dd:ee:09" {
		t.Errorf("existing device iLO MAC = %q, want %q", got, "aa:bb:cc:dd:ee:09")
	}
}

// TestTransformSystem_Modules verifies the module pass attaches modules to their
// parent devices, canonicalizing bay names and synthesizing module names.
//
// Why it matters: modules such as GPUs and NICs hang off devices, so the
// importer must resolve the parent, normalize the bay label, and name the module
// deterministically.
// Inputs: two devices and a GPU module (bay GPU0) plus a ConnectX-6 NIC module
// (bay PCIe5). Outputs: result.Modules with parent, bay name, and name set.
// Data choice: GPU0→"GPU 0" with name "gpu-...-GPU 0" proves bay canonicalization
// and GPU naming; the ConnectX-6 proves the "CX6-<device>" naming branch.
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

// TestTransformSystem_ManualParity verifies a CSV-driven import produces the
// same racks, devices, and modules as a result hand-built from the library
// constructors.
//
// Why it matters: it pins the importer's output to the canonical
// library-constructed shape, guarding against drift in slug resolution,
// placement, or naming.
// Inputs: a full SystemCSV (2 racks, 2 devices, 2 modules) and an equivalent
// manually built TransformResult. Outputs: equality of normalized parity
// snapshots via reflect.DeepEqual.
// Data choice: real slugs (xd670, proliant, h100, connectx-6) with placement and
// bays let the snapshot compare slug, status, U-height, rack position, face,
// role, and bay across both construction paths.
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

// TestTransformSystem_Defaults verifies rows inherit the global _defaults status
// when they omit their own.
//
// Why it matters: operators set a status once in _defaults, so racks and devices
// that leave status blank must pick it up rather than ship blank.
// Inputs: a SystemCSV with _defaults Status "Active" and a rack and device that
// omit status. Outputs: both objects with Status "Active".
// Data choice: leaving status blank on both a rack and a device proves the
// global default reaches each object type.
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

// TestTransformSystem_RackQtyMultiple verifies a rack row with quantity N
// expands into N racks with sequential one-based name suffixes.
//
// Why it matters: operators request identical racks in bulk, so a quantity must
// produce that many uniquely named racks.
// Inputs: one rack row with Qty 3. Outputs: three racks named rack-1, rack-2,
// rack-3.
// Data choice: Qty 3 is the smallest quantity that proves both the count and the
// sequential one-based suffixing.
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

// TestTransformSystem_EmptyData verifies an empty system batch transforms into
// an empty result without error.
//
// Why it matters: an import with no rows is a valid no-op, so the importer must
// return cleanly rather than fail or fabricate objects.
// Inputs: a SystemCSV with only an empty SectionDefaults map. Outputs: zero
// racks and zero devices, nil error.
// Data choice: a wholly empty batch is the minimal input proving the no-op path.
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

// --- Location tests ---

// TestTransformSystem_Locations verifies the location pass builds a parented
// hierarchy with parsed content types from system CSV location rows.
//
// Why it matters: locations are the topmost dependency in an inventory; racks
// and devices hang off them, so the importer must create them first and wire
// each child to its parent by name before anything else resolves.
// Inputs: a zero-value inventory plus a SystemCSV with a top-level "dc" location
// and a "section" child that references it. Outputs: result.Locations populated
// with parent links set.
// Data choice: a two-level dc→section hierarchy is the smallest input that
// proves both top-level (nil parent) and child (resolved parent) handling, and
// the comma-separated ContentTypes proves the split-and-trim logic.
func TestTransformSystem_Locations(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Locations: []import_.SystemRecord{
			{Section: "location", Name: "dc1", Role: "dc", ContentTypes: "dcim.rack, dcim.device"},
			{Section: "location", Name: "row-a", Role: "section", Location: "dc1"},
		},
	}

	result, err := TransformSystem(devicetypes.Inventory{}, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Locations) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(result.Locations))
	}

	byName := make(map[string]*devicetypes.CaniLocationType)
	for _, loc := range result.Locations {
		byName[loc.Name] = loc
	}

	dc := byName["dc1"]
	if dc == nil {
		t.Fatal("expected location dc1 to be created")
	}
	if dc.LocationType != "dc" {
		t.Errorf("dc1 LocationType = %q, want %q", dc.LocationType, "dc")
	}
	if len(dc.ContentTypes) != 2 {
		t.Errorf("dc1 ContentTypes = %d, want 2", len(dc.ContentTypes))
	}
	if dc.Parent != uuid.Nil {
		t.Errorf("dc1 Parent = %v, want Nil (top-level)", dc.Parent)
	}

	row := byName["row-a"]
	if row == nil {
		t.Fatal("expected location row-a to be created")
	}
	if row.Parent != dc.ID {
		t.Errorf("row-a Parent = %v, want %v (dc1)", row.Parent, dc.ID)
	}
}

// TestTransformSystem_LocationParentFromInventory verifies a new location
// resolves its parent against locations already present in the inventory.
//
// Why it matters: imports are incremental, so a section added in a later batch
// must attach to a datacenter created in an earlier one; the resolver therefore
// falls back to the existing inventory when the parent is absent from the
// current batch.
// Inputs: an inventory pre-seeded with a "dc-existing" location and a SystemCSV
// whose only location references it. Outputs: the new location's Parent set to
// the pre-seeded UUID.
// Data choice: seeding the parent only in the inventory (never in the batch)
// forces the findLocationByName fallback path rather than the in-batch map.
func TestTransformSystem_LocationParentFromInventory(t *testing.T) {
	inv := devicetypes.NewInventory()
	existingID := uuid.New()
	inv.Locations[existingID] = &devicetypes.CaniLocationType{ID: existingID, Name: "dc-existing", LocationType: "dc"}

	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Locations: []import_.SystemRecord{
			{Section: "location", Name: "row-b", Role: "section", Location: "dc-existing"},
		},
	}

	result, err := TransformSystem(*inv, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var row *devicetypes.CaniLocationType
	for _, loc := range result.Locations {
		if loc.Name == "row-b" {
			row = loc
		}
	}
	if row == nil {
		t.Fatal("expected location row-b to be created")
	}
	if row.Parent != existingID {
		t.Errorf("row-b Parent = %v, want %v (resolved from existing inventory)", row.Parent, existingID)
	}
}

// TestTransformSystem_LocationErrors verifies the location pass rejects rows
// that are missing required fields or reference an unknown parent.
//
// Why it matters: a malformed location would silently break the parent chain
// for every rack and device beneath it, so the importer must fail fast with a
// clear error instead of producing a corrupt topology.
// Inputs: SystemCSV variants each containing one invalid location row. Outputs:
// a non-nil error from TransformSystem in every case.
// Data choice: the three rows isolate the three distinct guards — missing Name,
// missing Role, and an unresolvable parent — so each error branch is proven
// independently rather than masked by a single combined failure.
func TestTransformSystem_LocationErrors(t *testing.T) {
	tests := []struct {
		name string
		loc  import_.SystemRecord
	}{
		{"missing name", import_.SystemRecord{Section: "location", Role: "dc"}},
		{"missing role", import_.SystemRecord{Section: "location", Name: "dc1"}},
		{"unknown parent", import_.SystemRecord{Section: "location", Name: "row-a", Role: "section", Location: "ghost-dc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &import_.SystemCSV{
				SectionDefaults: make(map[string]import_.SystemRecord),
				Locations:       []import_.SystemRecord{tt.loc},
			}
			if _, err := TransformSystem(*devicetypes.NewInventory(), data); err == nil {
				t.Fatalf("expected error for %s location row", tt.name)
			}
		})
	}
}

// TestTransformSystem_RackWithLocation verifies a rack is parented to a
// location created in the same import batch.
//
// Why it matters: racks must live inside a location for the exported topology
// to be valid, so the rack pass has to resolve the location name produced by
// the earlier location pass and stamp its UUID onto the rack.
// Inputs: a SystemCSV with one "dc1" location and one rack referencing it by
// name. Outputs: the rack's Location field set to the dc1 UUID.
// Data choice: a single location and rack keep the assertion unambiguous, and
// referencing the location by name (not UUID) exercises the name→ID resolution
// the CSV format relies on.
func TestTransformSystem_RackWithLocation(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Locations: []import_.SystemRecord{
			{Section: "location", Name: "dc1", Role: "dc"},
		},
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3701", Qty: 1, Location: "dc1", Status: "Available"},
		},
	}

	result, err := TransformSystem(*devicetypes.NewInventory(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var locID uuid.UUID
	for id, loc := range result.Locations {
		if loc.Name == "dc1" {
			locID = id
		}
	}
	if locID == uuid.Nil {
		t.Fatal("expected location dc1 to be created")
	}

	var rack *devicetypes.CaniRackType
	for _, r := range result.Racks {
		rack = r
	}
	if rack == nil {
		t.Fatal("expected rack to be created")
	}
	if rack.Location != locID {
		t.Errorf("rack Location = %v, want %v (dc1)", rack.Location, locID)
	}
}

// --- Connection tests ---

// TestTransformSystem_Connections verifies the connection pass resolves device
// names into a cable carrying the declared endpoints and cable properties.
//
// Why it matters: connections are the last pass and depend on every device
// already existing, so the importer must resolve both endpoint names to UUIDs
// and copy the cable's type, color, length, and unit onto the created cable.
// Inputs: a SystemCSV with two devices and one connection between them carrying
// type/color/length/unit. Outputs: a single cable with matching termination
// UUIDs, ports, and properties.
// Data choice: two distinct device names prove name→UUID resolution on both
// ends, and populating every cable property field proves each is propagated
// rather than dropped.
func TestTransformSystem_Connections(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "node-a", Qty: 1},
			{Section: "device", PartNumber: "hpe-xd670", Name: "node-b", Qty: 1},
		},
		Connections: []import_.SystemRecord{
			{Section: "connection", ADevice: "node-a", APort: "eth0", BDevice: "node-b", BPort: "eth0",
				PartNumber: "cat6", Color: "blue", Length: "5", LengthUnit: "m", Status: "Connected"},
		},
	}

	result, err := TransformSystem(*devicetypes.NewInventory(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Cables) != 1 {
		t.Fatalf("expected 1 cable, got %d", len(result.Cables))
	}

	deviceIDs := make(map[string]uuid.UUID)
	for id, dev := range result.Devices {
		deviceIDs[dev.Name] = id
	}

	var cable *devicetypes.CaniCableType
	for _, c := range result.Cables {
		cable = c
	}
	if cable.TerminationADevice != deviceIDs["node-a"] {
		t.Errorf("TerminationADevice = %v, want %v (node-a)", cable.TerminationADevice, deviceIDs["node-a"])
	}
	if cable.TerminationBDevice != deviceIDs["node-b"] {
		t.Errorf("TerminationBDevice = %v, want %v (node-b)", cable.TerminationBDevice, deviceIDs["node-b"])
	}
	if cable.TerminationAPort != "eth0" {
		t.Errorf("TerminationAPort = %q, want %q", cable.TerminationAPort, "eth0")
	}
	if cable.Slug != "cat6" {
		t.Errorf("cable Slug = %q, want %q", cable.Slug, "cat6")
	}
	if cable.Color != "blue" {
		t.Errorf("cable Color = %q, want %q", cable.Color, "blue")
	}
	if cable.Length == nil || *cable.Length != 5 {
		t.Errorf("cable Length = %v, want 5", cable.Length)
	}
	if cable.LengthUnit != "m" {
		t.Errorf("cable LengthUnit = %q, want %q", cable.LengthUnit, "m")
	}
}

// TestTransformSystem_ConnectionDefaults verifies cable defaults from the
// connection section fill in properties a connection row leaves blank.
//
// Why it matters: operators set section-wide defaults (e.g. a house color or
// length unit) once rather than on every row, so the importer must apply the
// connection section defaults to each cable that does not override them.
// Inputs: a SystemCSV with a connection section default for color/unit/status
// and one connection that omits those fields. Outputs: a cable that inherits
// the defaulted color, length unit, and status.
// Data choice: leaving exactly the defaulted fields blank on the connection row
// proves inheritance is sourced from the section defaults, not the row.
func TestTransformSystem_ConnectionDefaults(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: map[string]import_.SystemRecord{
			"connection": {Section: "connection", Color: "green", LengthUnit: "ft", Status: "Planned"},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "node-a", Qty: 1},
			{Section: "device", PartNumber: "hpe-xd670", Name: "node-b", Qty: 1},
		},
		Connections: []import_.SystemRecord{
			{Section: "connection", ADevice: "node-a", APort: "eth0", BDevice: "node-b", BPort: "eth0", PartNumber: "cat6"},
		},
	}

	result, err := TransformSystem(*devicetypes.NewInventory(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Cables) != 1 {
		t.Fatalf("expected 1 cable, got %d", len(result.Cables))
	}

	var cable *devicetypes.CaniCableType
	for _, c := range result.Cables {
		cable = c
	}
	if cable.Color != "green" {
		t.Errorf("cable Color = %q, want %q (from section defaults)", cable.Color, "green")
	}
	if cable.LengthUnit != "ft" {
		t.Errorf("cable LengthUnit = %q, want %q (from section defaults)", cable.LengthUnit, "ft")
	}
	if cable.Status != "Planned" {
		t.Errorf("cable Status = %q, want %q (from section defaults)", cable.Status, "Planned")
	}
}

// TestTransformSystem_ConnectionUnknownDevice verifies the connection pass
// fails when no connection can be resolved against the inventory.
//
// Why it matters: a connection naming a device that was never imported would
// otherwise yield zero cables silently; surfacing an error lets the operator
// fix the typo instead of shipping an incomplete topology.
// Inputs: a SystemCSV with one device and a connection whose B endpoint names a
// non-existent device. Outputs: a non-nil error from TransformSystem.
// Data choice: a single unresolvable endpoint drives the "no connections
// resolved" branch, the strictest failure the resolver reports.
func TestTransformSystem_ConnectionUnknownDevice(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "node-a", Qty: 1},
		},
		Connections: []import_.SystemRecord{
			{Section: "connection", ADevice: "node-a", APort: "eth0", BDevice: "ghost", BPort: "eth0"},
		},
	}
	if _, err := TransformSystem(*devicetypes.NewInventory(), data); err == nil {
		t.Fatal("expected error when all connections fail to resolve")
	}
}

// TestTransformSystem_RoleMissingName verifies the role pass rejects a role row
// without a Name.
//
// Why it matters: roles are referenced by name from device rows, so an unnamed
// role is unreferenceable and must be rejected rather than stored as a dangling
// entry.
// Inputs: a SystemCSV with one role row that has content types but no Name.
// Outputs: a non-nil error from TransformSystem.
// Data choice: supplying ContentTypes while omitting Name isolates the missing
// Name guard so the failure cannot be attributed to an otherwise empty row.
func TestTransformSystem_RoleMissingName(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Roles: []import_.SystemRecord{
			{Section: "role", ContentTypes: "dcim.device"},
		},
	}
	if _, err := TransformSystem(*devicetypes.NewInventory(), data); err == nil {
		t.Fatal("expected error for role missing Name")
	}
}

// TestTransformSystem_RackByPartNumber verifies a rack row whose identifier is a
// catalog part number (not a slug) is resolved through the part-number lookup.
//
// Why it matters: operators may identify racks by orderable part number; the
// importer must fall back to part-number resolution so those rows still inherit
// the correct rack-type U-height.
// Inputs: one rack row keyed by the real part number P9K58A. Outputs: a single
// rack with a U-height greater than zero sourced from the rack type.
// Data choice: P9K58A is the orderable part number of a real 48U rack type whose
// slug differs, so success proves the part-number branch ran, not slug lookup.
func TestTransformSystem_RackByPartNumber(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "P9K58A", Name: "x4000", Qty: 1, Status: "Available"},
		},
	}
	result, err := TransformSystem(*devicetypes.NewInventory(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Racks) != 1 {
		t.Fatalf("expected 1 rack, got %d", len(result.Racks))
	}
	for _, rack := range result.Racks {
		if rack.UHeight < 1 {
			t.Errorf("rack UHeight = %d, want > 0 (resolved by part number)", rack.UHeight)
		}
	}
}

// TestTransformSystem_RackMissingPartNumber verifies a rack row without a part
// number is rejected and surfaced as a wrapped TransformSystem error.
//
// Why it matters: a rack with no part number cannot be sized, so the importer
// must fail loudly rather than emit an unusable zero-height rack.
// Inputs: one rack row with an empty PartNumber. Outputs: a non-nil error from
// TransformSystem.
// Data choice: leaving only PartNumber blank isolates the missing-part-number
// guard from any other validation.
func TestTransformSystem_RackMissingPartNumber(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", Name: "x4000", Qty: 1},
		},
	}
	if _, err := TransformSystem(*devicetypes.NewInventory(), data); err == nil {
		t.Fatal("expected error for rack missing PartNumber")
	}
}

// TestTransformSystem_RackLocationFromInventory verifies a rack inherits a
// location that exists only in the pre-existing inventory, not the current batch.
//
// Why it matters: incremental imports reference locations created in prior runs,
// so the rack pass must fall back to the inventory lookup to keep parenting
// intact across batches.
// Inputs: an inventory pre-seeded with a "dc1" location and a rack row naming
// that location. Outputs: a rack whose Location points at the seeded location ID.
// Data choice: the location is placed only in the inventory (absent from the
// batch) so a correct parent can only come from the inventory fallback path.
func TestTransformSystem_RackLocationFromInventory(t *testing.T) {
	inv := *devicetypes.NewInventory()
	locID := uuid.New()
	inv.Locations[locID] = &devicetypes.CaniLocationType{ID: locID, Name: "dc1"}

	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x4000", Qty: 1, Location: "dc1"},
		},
	}
	result, err := TransformSystem(inv, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var found bool
	for _, rack := range result.Racks {
		if rack.Name == "x4000" {
			found = true
			if rack.Location != locID {
				t.Errorf("rack Location = %v, want %v (from inventory)", rack.Location, locID)
			}
		}
	}
	if !found {
		t.Fatal("expected rack x4000 in result")
	}
}

// TestTransformSystem_DeviceMissingPartNumber verifies a device row without a
// part number is rejected.
//
// Why it matters: a device with no part number cannot be typed or sized, so the
// importer must fail rather than create an untyped device.
// Inputs: a rack plus a device row with an empty PartNumber. Outputs: a non-nil
// error from TransformSystem.
// Data choice: a valid rack ensures the failure is attributable to the device's
// missing part number, not a rack problem.
func TestTransformSystem_DeviceMissingPartNumber(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x4000", Qty: 1},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", Name: "node1", Qty: 1, Rack: "x4000"},
		},
	}
	if _, err := TransformSystem(*devicetypes.NewInventory(), data); err == nil {
		t.Fatal("expected error for device missing PartNumber")
	}
}

// TestTransformSystem_DeviceByPartNumberAndQty verifies device resolution by
// part number and multi-quantity naming in a single pass.
//
// Why it matters: operators identify devices by orderable part number and often
// request several identical units, so the importer must resolve the type by part
// number and disambiguate the resulting device names.
// Inputs: one device row keyed by part number P67287-B21 with Qty 2. Outputs:
// two devices whose names are unique.
// Data choice: P67287-B21 is a real part number whose slug differs, and Qty 2 is
// the smallest quantity that forces name suffixing.
func TestTransformSystem_DeviceByPartNumberAndQty(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x4000", Qty: 1},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "P67287-B21", Name: "node", Qty: 2, Rack: "x4000"},
		},
	}
	result, err := TransformSystem(*devicetypes.NewInventory(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(result.Devices))
	}
	names := make(map[string]bool)
	for _, dev := range result.Devices {
		if names[dev.Name] {
			t.Errorf("duplicate device name %q; expected unique names for Qty>1", dev.Name)
		}
		names[dev.Name] = true
		if dev.Slug != "hpe-xd670" {
			t.Errorf("Slug = %q, want hpe-xd670 (resolved by part number)", dev.Slug)
		}
	}
}

// TestTransformSystem_DevicePlacementEdges verifies default face selection,
// minimum-height clamping, and the placement-failure warning path.
//
// Why it matters: a device row may omit a face, reference an unknown type with
// no U-height, or request a U position outside the rack; the importer must apply
// a default face, clamp height to at least 1U, and continue past a failed
// placement without aborting the whole import.
// Inputs: a 48U rack and two unknown-part-number devices — one at U10 with no
// face (default face, 1U clamp) and one at U100 (out-of-range placement).
// Outputs: two devices created with no error, the first faced "front".
// Data choice: an unknown part number guarantees a zero U-height to exercise the
// clamp, and U100 in a 48U rack guarantees the placement-failure warning.
func TestTransformSystem_DevicePlacementEdges(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x4000", Qty: 1},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "UNKNOWN-PN", Name: "noface", Qty: 1, Rack: "x4000", Position: 10},
			{Section: "device", PartNumber: "UNKNOWN-PN", Name: "overflow", Qty: 1, Rack: "x4000", Position: 100, Face: "front"},
		},
	}
	result, err := TransformSystem(*devicetypes.NewInventory(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(result.Devices))
	}
	for _, dev := range result.Devices {
		if dev.Name == "noface" && dev.Face != devicetypes.FaceFront {
			t.Errorf("noface Face = %q, want default %q", dev.Face, devicetypes.FaceFront)
		}
	}
}

// TestTransformSystem_ModuleByPartNumber verifies a module row keyed by part
// number resolves through the part-number lookup.
//
// Why it matters: modules are commonly identified by orderable part number, so
// the importer must resolve their type that way to populate bays and interfaces.
// Inputs: a device plus a module row keyed by part number P42351-B21. Outputs: a
// single module attached to the device.
// Data choice: P42351-B21 is a real NIC module part number whose slug differs,
// proving the part-number branch resolved the type.
func TestTransformSystem_ModuleByPartNumber(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x4000", Qty: 1},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "node1", Qty: 1, Rack: "x4000", Position: 1},
		},
		Modules: []import_.SystemRecord{
			{Section: "module", PartNumber: "P42351-B21", Qty: 1, Device: "node1", Bay: "PCIe1"},
		},
	}
	result, err := TransformSystem(*devicetypes.NewInventory(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(result.Modules))
	}
}

// TestTransformSystem_ModuleUnknownDevice verifies a module referencing a
// non-existent device fails as a wrapped TransformSystem error.
//
// Why it matters: a module must attach to a real parent device; a dangling
// reference indicates a data error the operator needs to fix before import.
// Inputs: a module row whose Device names a device that was never created.
// Outputs: a non-nil error from TransformSystem.
// Data choice: omitting the device section entirely guarantees the reference is
// unresolved, isolating the unknown-device guard.
func TestTransformSystem_ModuleUnknownDevice(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Modules: []import_.SystemRecord{
			{Section: "module", PartNumber: "P42351-B21", Qty: 1, Device: "ghost", Bay: "PCIe1"},
		},
	}
	if _, err := TransformSystem(*devicetypes.NewInventory(), data); err == nil {
		t.Fatal("expected error for module referencing unknown device")
	}
}

// TestTransformSystem_InterfaceSkips verifies interface rows missing an owner,
// missing a name, or carrying an invalid MAC are warned and skipped, not fatal.
//
// Why it matters: a single malformed interface row should not abort an otherwise
// valid import; the pass must tolerate and skip each malformed shape.
// Inputs: a device plus three interface rows — no Device, no Name, and a valid
// target with an unparseable MAC. Outputs: no error from TransformSystem.
// Data choice: each row isolates one skip branch (missing owner, missing name,
// MAC normalization failure) while the valid device keeps the run otherwise sound.
func TestTransformSystem_InterfaceSkips(t *testing.T) {
	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Racks: []import_.SystemRecord{
			{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x4000", Qty: 1},
		},
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "node1", Qty: 1, Rack: "x4000", Position: 1},
		},
		Interfaces: []import_.SystemRecord{
			{Section: "interface", Name: "orphan", MacAddress: "aa:bb:cc:dd:ee:ff"},
			{Section: "interface", Device: "node1", MacAddress: "aa:bb:cc:dd:ee:ff"},
			{Section: "interface", Device: "node1", Name: "iLO", MacAddress: "ZZ-NOT-A-MAC"},
		},
	}
	if _, err := TransformSystem(*devicetypes.NewInventory(), data); err != nil {
		t.Fatalf("malformed interface rows should be skipped, got error: %v", err)
	}
}

// TestTransformSystem_ConnectionGlobalStatusDefault verifies a global default
// status is applied to cables produced by the connection pass.
//
// Why it matters: operators may set a single global status (e.g. "Planned") for
// an entire import; the connection pass must thread that default into the cable
// defaults so every cable inherits it.
// Inputs: a SystemCSV whose global Defaults set Status and a connection that
// omits status. Outputs: a single cable with Status "Planned".
// Data choice: the status lives only in global Defaults (no connection section
// defaults), isolating the global-status branch of the cable defaults builder.
func TestTransformSystem_ConnectionGlobalStatusDefault(t *testing.T) {
	data := &import_.SystemCSV{
		Defaults:        import_.SystemRecord{Status: "Planned"},
		SectionDefaults: make(map[string]import_.SystemRecord),
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "node-a", Qty: 1},
			{Section: "device", PartNumber: "hpe-xd670", Name: "node-b", Qty: 1},
		},
		Connections: []import_.SystemRecord{
			{Section: "connection", ADevice: "node-a", APort: "eth0", BDevice: "node-b", BPort: "eth0", PartNumber: "cat6"},
		},
	}
	result, err := TransformSystem(*devicetypes.NewInventory(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Cables) != 1 {
		t.Fatalf("expected 1 cable, got %d", len(result.Cables))
	}
	for _, cable := range result.Cables {
		if cable.Status != "Planned" {
			t.Errorf("cable Status = %q, want %q (from global defaults)", cable.Status, "Planned")
		}
	}
}

// TestTransformSystem_DeviceRackFromInventory verifies a device parents to a
// rack that exists only in the pre-existing inventory.
//
// Why it matters: incremental imports place new devices into racks from prior
// runs, so the device pass must resolve the parent rack from the inventory when
// the batch does not contain it.
// Inputs: an inventory pre-seeded with rack "existing-rack" and a device row
// naming that rack. Outputs: a device whose Parent is the seeded rack ID.
// Data choice: the rack is only in the inventory (absent from the batch) so a
// correct parent can come only from the inventory fallback.
func TestTransformSystem_DeviceRackFromInventory(t *testing.T) {
	inv := *devicetypes.NewInventory()
	rackID := uuid.New()
	inv.Racks[rackID] = &devicetypes.CaniRackType{ID: rackID, Name: "existing-rack", UHeight: 48}

	data := &import_.SystemCSV{
		SectionDefaults: make(map[string]import_.SystemRecord),
		Devices: []import_.SystemRecord{
			{Section: "device", PartNumber: "hpe-xd670", Name: "node1", Qty: 1, Rack: "existing-rack", Position: 1},
		},
	}
	result, err := TransformSystem(inv, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var found bool
	for _, dev := range result.Devices {
		if dev.Name == "node1" {
			found = true
			if dev.Parent != rackID {
				t.Errorf("device Parent = %v, want %v (from inventory rack)", dev.Parent, rackID)
			}
		}
	}
	if !found {
		t.Fatal("expected device node1 in result")
	}
}

// TestResolveSystemModuleBayName verifies bay-name canonicalization, including
// the nil-device, empty-bay, matched, and unmatched fallthrough paths.
//
// Why it matters: module bay names from CSV must be normalized to the device's
// canonical bay label when one matches and otherwise passed through unchanged,
// so module placement is consistent without dropping unknown bays.
// Inputs: combinations of nil/real device, empty/known/unknown requested bay.
// Outputs: requested bay echoed for nil-device, empty, and unmatched cases; the
// canonical bay name for a position match.
// Data choice: a device whose bay Position "GPU0" maps to Name "GPU 0" exercises
// the position-match branch, while "ZZZ" exercises the unmatched fallthrough.
func TestResolveSystemModuleBayName(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		ModuleBays: []devicetypes.ModuleBaySpec{{Name: "GPU 0", Position: "GPU0"}},
	}
	tests := []struct {
		name      string
		device    *devicetypes.CaniDeviceType
		requested string
		want      string
	}{
		{"nil device echoes requested", nil, "GPU0", "GPU0"},
		{"empty bay echoes empty", dev, "", ""},
		{"position match returns canonical name", dev, "GPU0", "GPU 0"},
		{"name match returns canonical name", dev, "GPU 0", "GPU 0"},
		{"unmatched bay echoes requested", dev, "ZZZ", "ZZZ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveSystemModuleBayName(tt.device, tt.requested); got != tt.want {
				t.Errorf("resolveSystemModuleBayName(%v, %q) = %q, want %q", tt.device, tt.requested, got, tt.want)
			}
		})
	}
}

// TestSynthesizeSystemModuleName verifies synthesized module names across the
// guard, GPU, ConnectX, and deterministic-fallback branches.
//
// Why it matters: a module without an explicit name must never be left blank
// (blank-named modules are dropped from summaries and collide when Qty > 1), so
// only the nil-module and already-named guards return empty; every other shape
// yields a deterministic name from the GPU, ConnectX, or slug/device/bay
// fallback.
// Inputs: nil module, already-named module, a nil-device GPU, a GPU with and
// without a bay, a ConnectX-6 module, and an unrelated module. Outputs: empty
// for the two guards and a formatted name for every other case.
// Data choice: the GPU type, a "connectx-6" slug, and a plain slug select each
// branch, while the nil-device and no-bay rows exercise the fallback's optional
// device and bay segments.
func TestSynthesizeSystemModuleName(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Name: "node-a"}
	tests := []struct {
		name   string
		module *devicetypes.CaniModuleType
		device *devicetypes.CaniDeviceType
		want   string
	}{
		{"nil module", nil, dev, ""},
		{"already named", &devicetypes.CaniModuleType{Name: "preset"}, dev, ""},
		{"nil device falls back to slug and bay", &devicetypes.CaniModuleType{Type: devicetypes.TypeGPU, ModuleBayName: "GPU 0"}, nil, "module-GPU 0"},
		{"gpu without bay falls back to device", &devicetypes.CaniModuleType{Type: devicetypes.TypeGPU}, dev, "module-node-a"},
		{"gpu with bay", &devicetypes.CaniModuleType{Type: devicetypes.TypeGPU, ModuleBayName: "GPU 0"}, dev, "gpu-node-a-GPU 0"},
		{"connectx six", &devicetypes.CaniModuleType{Slug: "nvidia-connectx-6-dx-100gbe"}, dev, "CX6-node-a"},
		{"other module falls back to slug and device", &devicetypes.CaniModuleType{Slug: "some-other-nic"}, dev, "some-other-nic-node-a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := synthesizeSystemModuleName(tt.module, tt.device); got != tt.want {
				t.Errorf("synthesizeSystemModuleName = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFindResultInterfaceSpec verifies interface-spec lookup across device-name
// mismatch, device match, module match, and not-found paths.
//
// Why it matters: applying per-interface metadata requires locating the spec on
// either a device or a module by owner name; the resolver must skip non-matching
// devices, search modules, and report nil when nothing matches.
// Inputs: a TransformResult holding one device (dev1/eth0) and one module
// (mod1/port0), queried for the device interface, the module interface, and a
// missing owner. Outputs: non-nil specs for the two hits, nil for the miss.
// Data choice: distinct owner and interface names make each lookup target a
// single, unambiguous branch (device match, module match after device mismatch,
// and total miss).
func TestFindResultInterfaceSpec(t *testing.T) {
	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {Name: "dev1", Interfaces: []devicetypes.InterfaceSpec{{Name: "eth0"}}},
		},
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			uuid.New(): {Name: "mod1", Interfaces: []devicetypes.InterfaceSpec{{Name: "port0"}}},
		},
	}
	t.Run("device interface", func(t *testing.T) {
		if spec := findResultInterfaceSpec(result, "dev1", "eth0"); spec == nil {
			t.Error("expected to find device interface eth0")
		}
	})
	t.Run("module interface after device mismatch", func(t *testing.T) {
		if spec := findResultInterfaceSpec(result, "mod1", "port0"); spec == nil {
			t.Error("expected to find module interface port0")
		}
	})
	t.Run("not found", func(t *testing.T) {
		if spec := findResultInterfaceSpec(result, "nope", "eth0"); spec != nil {
			t.Error("expected nil for unknown owner")
		}
	})
}

// TestFindInventoryInterfaceSpec verifies inventory interface-spec lookup across
// the nil-inventory, device-match, module-match, and not-found paths.
//
// Why it matters: applying interface metadata to pre-existing hardware relies on
// resolving the spec from the inventory's device or module maps, and a nil
// inventory must be tolerated rather than panic.
// Inputs: a nil inventory, then an inventory holding device dev1/eth0 and module
// mod1/port0, queried for each owner and a missing one. Outputs: nil for the nil
// inventory and the miss; non-nil specs for the device and module hits.
// Data choice: distinct owner and interface names make each lookup a single
// unambiguous branch, and the nil inventory isolates the guard clause.
func TestFindInventoryInterfaceSpec(t *testing.T) {
	if spec := findInventoryInterfaceSpec(nil, "dev1", "eth0"); spec != nil {
		t.Error("expected nil for a nil inventory")
	}
	inv := devicetypes.NewInventory()
	inv.Devices[uuid.New()] = &devicetypes.CaniDeviceType{Name: "dev1", Interfaces: []devicetypes.InterfaceSpec{{Name: "eth0"}}}
	inv.Modules[uuid.New()] = &devicetypes.CaniModuleType{Name: "mod1", Interfaces: []devicetypes.InterfaceSpec{{Name: "port0"}}}
	t.Run("device interface", func(t *testing.T) {
		if spec := findInventoryInterfaceSpec(inv, "dev1", "eth0"); spec == nil {
			t.Error("expected to find device interface eth0")
		}
	})
	t.Run("module interface", func(t *testing.T) {
		if spec := findInventoryInterfaceSpec(inv, "mod1", "port0"); spec == nil {
			t.Error("expected to find module interface port0")
		}
	})
	t.Run("not found", func(t *testing.T) {
		if spec := findInventoryInterfaceSpec(inv, "nope", "eth0"); spec != nil {
			t.Error("expected nil for unknown owner")
		}
	})
}
