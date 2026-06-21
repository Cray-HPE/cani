package transform

import (
	"io"
	"os"
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

// --- Test helper ---

type fakeRecordProvider struct{ records []import_.CsvRecord }

func (f *fakeRecordProvider) GetRecords() []import_.CsvRecord { return f.records }

// withStdin redirects os.Stdin to a pipe preloaded with input for the duration
// of the test, then restores it. An empty input yields immediate EOF, which
// drives interactive prompts to return an error.
func withStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = old
		_ = r.Close()
	})
	go func() {
		_, _ = io.WriteString(w, input)
		_ = w.Close()
	}()
}

type fakeSystemProvider struct {
	data     *import_.SystemCSV
	isSystem bool
}

func (f *fakeSystemProvider) GetSystemRecords() *import_.SystemCSV { return f.data }
func (f *fakeSystemProvider) IsSystemImport() bool                 { return f.isSystem }

// setupTransformTest saves and restores providerGetter and config.Cfg, then
// sets a fakeRecordProvider with the given records.
func setupTransformTest(t *testing.T, records []import_.CsvRecord) {
	t.Helper()
	oldGetter := providerGetter
	oldCfg := config.Cfg
	t.Cleanup(func() {
		providerGetter = oldGetter
		config.Cfg = oldCfg
		resetRackPositionStates()
	})
	config.Cfg = &config.Config{}
	SetProviderGetter(func() interface{ GetRecords() []import_.CsvRecord } {
		return &fakeRecordProvider{records: records}
	})
}

// --- Transform entry point tests ---

// TestSetProviderGetter verifies the setter installs a non-nil record-provider
// accessor that the package can invoke.
//
// Why it matters: the transform layer is decoupled from the provider singleton
// to avoid an import cycle, so the parent package must inject CSV record access
// through this setter before a flat-CSV import can run.
// Inputs: a getter closure that records its invocation. Outputs: a non-nil
// providerGetter that runs the closure when called.
// Data choice: a boolean tripwire is the minimal observable proof the injected
// getter is the one stored and called.
func TestSetProviderGetter(t *testing.T) {
	old := providerGetter
	t.Cleanup(func() { providerGetter = old })

	called := false
	SetProviderGetter(func() interface{ GetRecords() []import_.CsvRecord } {
		called = true
		return &fakeRecordProvider{}
	})

	if providerGetter == nil {
		t.Fatal("providerGetter should not be nil after SetProviderGetter")
	}
	providerGetter()
	if !called {
		t.Error("expected providerGetter to be called")
	}
}

// TestSetSystemProviderGetter verifies the setter installs a non-nil system
// provider accessor that the package can invoke.
//
// Why it matters: the transform layer is decoupled from the provider singleton
// to avoid an import cycle, so the parent package must be able to inject system
// CSV access through this setter before any system import can run.
// Inputs: a getter closure that records its invocation. Outputs: a non-nil
// systemProviderGetter that, when called, runs the closure.
// Data choice: a boolean tripwire is the minimal observable proof the injected
// getter is the one actually stored and called.
func TestSetSystemProviderGetter(t *testing.T) {
	old := systemProviderGetter
	t.Cleanup(func() { systemProviderGetter = old })

	called := false
	SetSystemProviderGetter(func() interface {
		GetSystemRecords() *import_.SystemCSV
		IsSystemImport() bool
	} {
		called = true
		return &fakeSystemProvider{}
	})

	if systemProviderGetter == nil {
		t.Fatal("systemProviderGetter should not be nil after SetSystemProviderGetter")
	}
	systemProviderGetter()
	if !called {
		t.Error("expected systemProviderGetter to be called")
	}
}

// TestTransform_SystemCSVRouting verifies Transform dispatches to the system
// transform only when the provider reports a system import.
//
// Why it matters: a single Transform entry point serves both the flat CSV and
// the richer system CSV formats, so it must branch on IsSystemImport to avoid
// running the wrong pipeline against the wrong data shape.
// Inputs: a system provider toggled true then false, paired with matching
// records. Outputs: one rack from the system pipeline when true, one rack from
// the flat-CSV pipeline when false.
// Data choice: a single rack in each branch is the smallest output that proves
// which pipeline ran, and a real rack slug keeps the system path's library
// lookup realistic.
func TestTransform_SystemCSVRouting(t *testing.T) {
	t.Run("routes to system transform when IsSystemImport is true", func(t *testing.T) {
		oldSys := systemProviderGetter
		t.Cleanup(func() { systemProviderGetter = oldSys })

		data := &import_.SystemCSV{
			SectionDefaults: make(map[string]import_.SystemRecord),
			Racks: []import_.SystemRecord{
				{Section: "rack", PartNumber: "hpe-48u-800mmx1200mm-g2-enterprise-shock-rack", Name: "x3701", Qty: 1, Status: "Available"},
			},
		}
		SetSystemProviderGetter(func() interface {
			GetSystemRecords() *import_.SystemCSV
			IsSystemImport() bool
		} {
			return &fakeSystemProvider{data: data, isSystem: true}
		})

		result, err := Transform(devicetypes.Inventory{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Racks) != 1 {
			t.Errorf("expected 1 rack from system transform, got %d", len(result.Racks))
		}
	})

	t.Run("falls through to CSV transform when IsSystemImport is false", func(t *testing.T) {
		oldSys := systemProviderGetter
		t.Cleanup(func() { systemProviderGetter = oldSys })
		SetSystemProviderGetter(func() interface {
			GetSystemRecords() *import_.SystemCSV
			IsSystemImport() bool
		} {
			return &fakeSystemProvider{isSystem: false}
		})

		setupTransformTest(t, []import_.CsvRecord{
			{PartNumber: "FAKE-RACK-PN", Description: "48U Rack Cabinet", Quantity: 1, ConfigGroup: "0100"},
		})

		result, err := Transform(devicetypes.Inventory{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Racks) != 1 {
			t.Errorf("expected 1 rack from CSV transform, got %d", len(result.Racks))
		}
	})
}

// TestTransform verifies Transform guards a nil provider, no-ops on empty input,
// builds racks and devices from records, and propagates a classification error.
//
// Why it matters: Transform is the flat-CSV entry point, so it must reject a
// missing provider, stay a no-op when there is nothing to import, produce
// inventory from rows, and surface unclassifiable rows.
// Inputs: a nil providerGetter, nil records, a rack+server batch, and an
// unclassifiable memory-kit row. Outputs: errors or a populated TransformResult.
// Data choice: a 48U rack plus a quantity-2 server prove count expansion, while
// the DDR5 memory kit is a real description that classifies to nothing.
func TestTransform(t *testing.T) {
	t.Run("nil provider getter", func(t *testing.T) {
		old := providerGetter
		t.Cleanup(func() { providerGetter = old })
		providerGetter = nil

		if _, err := Transform(devicetypes.Inventory{}); err == nil {
			t.Fatal("expected error when providerGetter is nil")
		}
	})

	t.Run("empty records", func(t *testing.T) {
		setupTransformTest(t, nil)
		result, err := Transform(devicetypes.Inventory{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Racks) != 0 || len(result.Devices) != 0 || len(result.Cables) != 0 {
			t.Error("expected empty result for empty records")
		}
	})

	t.Run("with records", func(t *testing.T) {
		setupTransformTest(t, []import_.CsvRecord{
			{PartNumber: "FAKE-RACK-PN", Description: "48U Rack Cabinet", Quantity: 1, ConfigGroup: "0100"},
			{PartNumber: "FAKE-SRV-PN", Description: "ProLiant DL380 Server", Quantity: 2, ConfigGroup: "0300"},
		})
		result, err := Transform(devicetypes.Inventory{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Racks) != 1 {
			t.Errorf("expected 1 rack, got %d", len(result.Racks))
		}
		if len(result.Devices) != 2 {
			t.Errorf("expected 2 devices, got %d", len(result.Devices))
		}
	})

	t.Run("classify error", func(t *testing.T) {
		setupTransformTest(t, []import_.CsvRecord{
			{PartNumber: "FAKE-UNKNOWN", Description: "", Quantity: 1},
		})
		if _, err := Transform(devicetypes.Inventory{}); err == nil {
			t.Fatal("expected error for unclassifiable record")
		}
	})
}

// --- Classification tests ---

// TestClassifyRecords verifies records are bucketed into racks, devices, or
// cables, with an error for an unclassifiable row.
//
// Why it matters: classification is the first transform step that routes each
// row to its builder, so every hardware category and the reject path must be
// correct.
// Inputs: a table of single- and mixed-type record batches. Outputs: per-category
// counts or an error.
// Data choice: one row per category (rack, cable-by-endpoint,
// cable-by-description, switch, node), an unclassifiable memory kit, and a mixed
// batch cover every branch.
func TestClassifyRecords(t *testing.T) {
	tests := []struct {
		name        string
		records     []import_.CsvRecord
		wantRacks   int
		wantDevices int
		wantCables  int
		wantErr     bool
	}{
		{"empty records", nil, 0, 0, 0, false},
		{"rack by description", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE 48U 800mmx1200mm G2 Enterprise Shock Rack", Quantity: 1},
		}, 1, 0, 0, false},
		{"cable by explicit endpoints", []import_.CsvRecord{
			{Description: "link", Quantity: 1, SourceDevice: "sw1", DestDevice: "srv1"},
		}, 0, 0, 1, false},
		{"cable by description pattern", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE Cat6 RJ45 M/M 2m", Quantity: 1},
		}, 0, 0, 1, false},
		{"switch by description", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE Aruba Networking 8360-48Y6C", Quantity: 1},
		}, 0, 1, 0, false},
		{"node by description", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE ProLiant DL380 Gen11", Quantity: 1},
		}, 0, 1, 0, false},
		{"unclassifiable returns error", []import_.CsvRecord{
			{PartNumber: "X", Description: "HPE 64GB DDR5 Memory Kit", Quantity: 1},
		}, 0, 0, 0, true},
		{"mixed record types", []import_.CsvRecord{
			{PartNumber: "R", Description: "HPE 48U Rack", Quantity: 1},
			{PartNumber: "S", Description: "HPE Aruba Switch", Quantity: 1},
			{PartNumber: "C", Description: "HPE Cat6 Cable 2m", Quantity: 1},
		}, 1, 1, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := classifyRecords(tt.records)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.racks) != tt.wantRacks {
				t.Errorf("racks = %d, want %d", len(got.racks), tt.wantRacks)
			}
			if len(got.devices) != tt.wantDevices {
				t.Errorf("devices = %d, want %d", len(got.devices), tt.wantDevices)
			}
			if len(got.cables) != tt.wantCables {
				t.Errorf("cables = %d, want %d", len(got.cables), tt.wantCables)
			}
		})
	}
}

// --- Item creation tests ---

// TestCreateItemsFromRecord verifies a record expands into N items of its
// hardware type, registers them in inventory, and indexes them by config group.
//
// Why it matters: this is the per-row fan-out that builds inventory and the group
// indexes auto-cabling later relies on, so counts, inventory insertion, and
// grouping must all hold.
// Inputs: a quantity-2 rack row (group 0100), a quantity-3 node row (group
// 0200), and a node row with no config group. Outputs: created items, inventory
// entries, and group-index entries.
// Data choice: distinct quantities prove per-type expansion; the no-group row
// proves grouping is skipped when ConfigGroup is empty.
func TestCreateItemsFromRecord(t *testing.T) {
	t.Run("creates rack items", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
			Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		}
		racksByGroup := make(map[string][]uuid.UUID)
		devicesByGroup := make(map[string][]uuid.UUID)

		rec := import_.CsvRecord{PartNumber: "TEST-RACK", Description: "Test Rack", Quantity: 2, ConfigGroup: "0100"}
		result := createItemsFromRecord(inv, rec, "rack", racksByGroup, devicesByGroup)

		if len(result.Racks) != 2 {
			t.Errorf("expected 2 racks, got %d", len(result.Racks))
		}
		if len(result.Devices) != 0 {
			t.Errorf("expected 0 devices, got %d", len(result.Devices))
		}
		if len(inv.Racks) != 2 {
			t.Errorf("expected 2 racks in inventory, got %d", len(inv.Racks))
		}
		if len(racksByGroup["0100"]) != 2 {
			t.Errorf("expected 2 racks in group 0100, got %d", len(racksByGroup["0100"]))
		}
	})

	t.Run("creates device items", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
			Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		}
		racksByGroup := make(map[string][]uuid.UUID)
		devicesByGroup := make(map[string][]uuid.UUID)

		rec := import_.CsvRecord{PartNumber: "TEST-NODE", Description: "Test Server", Quantity: 3, ConfigGroup: "0200"}
		result := createItemsFromRecord(inv, rec, "node", racksByGroup, devicesByGroup)

		if len(result.Devices) != 3 {
			t.Errorf("expected 3 devices, got %d", len(result.Devices))
		}
		if len(result.Racks) != 0 {
			t.Errorf("expected 0 racks, got %d", len(result.Racks))
		}
		if len(inv.Devices) != 3 {
			t.Errorf("expected 3 devices in inventory, got %d", len(inv.Devices))
		}
		if len(devicesByGroup["0200"]) != 3 {
			t.Errorf("expected 3 devices in group 0200, got %d", len(devicesByGroup["0200"]))
		}
	})

	t.Run("no config group skips grouping", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
			Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		}
		racksByGroup := make(map[string][]uuid.UUID)
		devicesByGroup := make(map[string][]uuid.UUID)

		rec := import_.CsvRecord{PartNumber: "TEST", Description: "Standalone Device", Quantity: 1}
		createItemsFromRecord(inv, rec, "node", racksByGroup, devicesByGroup)

		if len(devicesByGroup) != 0 {
			t.Errorf("expected no config group entries, got %d", len(devicesByGroup))
		}
	})
}

// TestCreateRackWithoutLibraryMatch verifies a rack whose part number is not in
// the library is created with the default 48U height and still indexed.
//
// Why it matters: unknown rack parts must not break the import, so the builder
// falls back to a sane default height and registers the rack normally.
// Inputs: a rack record with a fake part number and group 0100. Outputs: a rack
// with the given ID, UHeight 48, present in inventory and the group index.
// Data choice: a fake part number guarantees the library miss, isolating the
// default-height fallback.
func TestCreateRackWithoutLibraryMatch(t *testing.T) {
	inv := &devicetypes.Inventory{Racks: make(map[uuid.UUID]*devicetypes.CaniRackType)}
	racksByGroup := make(map[string][]uuid.UUID)
	id := uuid.New()

	rec := import_.CsvRecord{PartNumber: "FAKE-RACK-PN", Description: "Custom 42U Rack", ConfigGroup: "0100"}
	rack := createRack(inv, id, "Custom 42U Rack", rec, racksByGroup)

	if rack.ID != id {
		t.Errorf("rack ID = %v, want %v", rack.ID, id)
	}
	if rack.UHeight != 48 {
		t.Errorf("default UHeight = %d, want 48", rack.UHeight)
	}
	if _, ok := inv.Racks[id]; !ok {
		t.Error("rack should be in inventory")
	}
	if len(racksByGroup["0100"]) != 1 {
		t.Error("rack should be tracked in racksByGroup")
	}
}

// TestBuildDeviceFromRecord verifies a device is built from a record with its
// ID, name, part number, and hardware type populated.
//
// Why it matters: even without a library match a device must carry its
// identifying fields so later passes can place and name it.
// Inputs: a device record with a fake part number and the name "DL380-001".
// Outputs: a device with matching ID, Name, PartNumber, and Type "node".
// Data choice: a fake part number keeps the test on the no-library path so only
// the record-derived fields are asserted.
func TestBuildDeviceFromRecord(t *testing.T) {
	id := uuid.New()
	rec := import_.CsvRecord{PartNumber: "TEST-FAKE-PN", Description: "HPE ProLiant DL380", ConfigGroup: "0200"}
	device := buildDeviceFromRecord(id, "DL380-001", rec, "node")

	if device.ID != id {
		t.Errorf("ID = %v, want %v", device.ID, id)
	}
	if device.Name != "DL380-001" {
		t.Errorf("Name = %q, want %q", device.Name, "DL380-001")
	}
	if device.PartNumber != "TEST-FAKE-PN" {
		t.Errorf("PartNumber = %q, want %q", device.PartNumber, "TEST-FAKE-PN")
	}
	if device.Type != devicetypes.Type("node") {
		t.Errorf("Type = %q, want %q", device.Type, "node")
	}
}

// TestPopulateFromDeviceType verifies library device-type fields are copied onto
// a device, but an empty type does not overwrite the existing type.
//
// Why it matters: library population enriches a bare device with manufacturer,
// model, ports, bays, and interfaces, so every field must copy while a blank
// library type must not clobber a known one.
// Inputs: a device plus a fully populated device type, and a switch device plus
// a device type with an empty Type. Outputs: the device with all fields copied,
// or its original type preserved.
// Data choice: a device type carrying one of every collection (interfaces,
// console/power ports, bays, identifications) proves each copy; the empty-Type
// case proves the guard.
func TestPopulateFromDeviceType(t *testing.T) {
	t.Run("copies all fields", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{Name: "original", Slug: "old-slug"}
		dt := &devicetypes.CaniDeviceType{
			Slug: "new-slug", Manufacturer: "HPE", Model: "DL380 Gen11",
			Description: "Rack server",
			Type:        devicetypes.Type("node"),
			UHeight:     2,
			IsFullDepth: true,
			Weight:      15.5,
			WeightUnit:  "kg",
			Comments:    "from library",
			Interfaces:  []devicetypes.InterfaceSpec{{Name: "eth0"}},
			ConsolePorts: []devicetypes.ConsolePortSpec{{
				Name: "console0",
			}},
			PowerPorts: []devicetypes.PowerPortSpec{{
				Name: "PSU1",
			}},
			ModuleBays: []devicetypes.ModuleBaySpec{{
				Name: "PCIe1",
			}},
			DeviceBays: []devicetypes.DeviceBaySpec{{
				Name: "bay1",
			}},
			Identifications: []devicetypes.Identification{{
				Manufacturer: "HPE",
				Model:        "DL380 Gen11",
			}},
		}
		populateFromDeviceType(device, dt)

		if device.Slug != "new-slug" {
			t.Errorf("Slug = %q, want %q", device.Slug, "new-slug")
		}
		if device.Manufacturer != "HPE" {
			t.Errorf("Manufacturer = %q, want %q", device.Manufacturer, "HPE")
		}
		if device.Model != "DL380 Gen11" {
			t.Errorf("Model = %q, want %q", device.Model, "DL380 Gen11")
		}
		if device.Type != devicetypes.Type("node") {
			t.Errorf("Type = %q, want %q", device.Type, "node")
		}
		if device.Description != "Rack server" {
			t.Errorf("Description = %q, want %q", device.Description, "Rack server")
		}
		if device.UHeight != 2 {
			t.Errorf("UHeight = %d, want 2", device.UHeight)
		}
		if !device.IsFullDepth {
			t.Fatal("expected IsFullDepth to be true")
		}
		if device.Weight != 15.5 {
			t.Errorf("Weight = %v, want 15.5", device.Weight)
		}
		if device.WeightUnit != "kg" {
			t.Errorf("WeightUnit = %q, want %q", device.WeightUnit, "kg")
		}
		if device.Comments != "from library" {
			t.Errorf("Comments = %q, want %q", device.Comments, "from library")
		}
		if len(device.Interfaces) != 1 {
			t.Errorf("Interfaces len = %d, want 1", len(device.Interfaces))
		}
		if len(device.ConsolePorts) != 1 {
			t.Errorf("ConsolePorts len = %d, want 1", len(device.ConsolePorts))
		}
		if len(device.PowerPorts) != 1 {
			t.Errorf("PowerPorts len = %d, want 1", len(device.PowerPorts))
		}
		if len(device.ModuleBays) != 1 {
			t.Errorf("ModuleBays len = %d, want 1", len(device.ModuleBays))
		}
		if len(device.DeviceBays) != 1 {
			t.Errorf("DeviceBays len = %d, want 1", len(device.DeviceBays))
		}
		if len(device.Identifications) != 1 {
			t.Errorf("Identifications len = %d, want 1", len(device.Identifications))
		}
	})

	t.Run("does not overwrite type when empty", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{Type: devicetypes.Type("switch")}
		dt := &devicetypes.CaniDeviceType{Slug: "new-slug", Type: ""}
		populateFromDeviceType(device, dt)

		if device.Type != devicetypes.Type("switch") {
			t.Errorf("Type = %q, want %q (should not overwrite with empty)", device.Type, "switch")
		}
	})
}

// TestGetDeviceUHeight verifies U-height resolution returns 0 when neither the
// part number nor the slug is in the library.
//
// Why it matters: placement needs a height, so an unknown device must report 0
// rather than guess, signaling the caller to clamp.
// Inputs: a device with unknown part number and slug, and a wholly empty device.
// Outputs: 0 in both cases.
// Data choice: fake and empty identifiers both miss the library, isolating the
// not-found return.
func TestGetDeviceUHeight(t *testing.T) {
	tests := []struct {
		name   string
		device *devicetypes.CaniDeviceType
	}{
		{"unknown part number and slug", &devicetypes.CaniDeviceType{PartNumber: "FAKE-UNKNOWN-PN", Slug: "fake-unknown-slug"}},
		{"empty part number and slug", &devicetypes.CaniDeviceType{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDeviceUHeight(tt.device); got != 0 {
				t.Errorf("expected 0 for unknown device, got %d", got)
			}
		})
	}
}

// --- Rack position and zone tests ---

// TestDeviceTypePriority verifies each hardware type maps to a rack-fill
// priority (0 bottom, 1 middle, 2 top), with unknown defaulting to 1.
//
// Why it matters: rack placement orders devices by these priorities, so power
// gear sinks to the bottom, switches rise to the top, and compute fills the
// middle.
// Inputs: representative type strings including empty. Outputs: the integer
// priority.
// Data choice: pdu/cdu (0), node/blade/chassis and empty (1), and the switch
// family (2) cover every tier and the default.
func TestDeviceTypePriority(t *testing.T) {
	tests := []struct {
		hwType string
		want   int
	}{
		{"pdu", 0}, {"cdu", 0},
		{"node", 1}, {"blade", 1}, {"chassis", 1},
		{"switch", 2}, {"mgmt-switch", 2}, {"hsn-switch", 2},
		{"", 1},
	}
	for _, tt := range tests {
		t.Run(tt.hwType, func(t *testing.T) {
			if got := deviceTypePriority(tt.hwType); got != tt.want {
				t.Errorf("deviceTypePriority(%q) = %d, want %d", tt.hwType, got, tt.want)
			}
		})
	}
}

// TestSortDevicesByRackPriority verifies a device ID slice is sorted in place
// into non-decreasing rack priority.
//
// Why it matters: placement consumes a priority-ordered list, so the sort must
// group power gear first and switches last regardless of input order.
// Inputs: an unordered five-device slice spanning all three priority tiers.
// Outputs: the slice reordered so priority never decreases.
// Data choice: one device per type (switch, node, pdu, blade, cdu) proves the
// two priority-0 devices lead and the priority-2 switch trails.
func TestSortDevicesByRackPriority(t *testing.T) {
	inv := &devicetypes.Inventory{Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType)}

	switchID, nodeID, pduID, bladeID, cduID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch")}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node")}
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu")}
	inv.Devices[bladeID] = &devicetypes.CaniDeviceType{ID: bladeID, Type: devicetypes.Type("blade")}
	inv.Devices[cduID] = &devicetypes.CaniDeviceType{ID: cduID, Type: devicetypes.Type("cdu")}

	ids := []uuid.UUID{switchID, nodeID, pduID, bladeID, cduID}
	sortDevicesByRackPriority(inv, ids)

	// Verify ordering: PDUs/CDUs first, then nodes/blades, then switches
	for i := 1; i < len(ids); i++ {
		pri := deviceTypePriority(string(inv.Devices[ids[i]].Type))
		prevPri := deviceTypePriority(string(inv.Devices[ids[i-1]].Type))
		if pri < prevPri {
			t.Errorf("device at index %d (%s, priority %d) sorted before index %d (%s, priority %d)",
				i, inv.Devices[ids[i]].Type, pri, i-1, inv.Devices[ids[i-1]].Type, prevPri)
		}
	}

	for _, id := range ids[:2] {
		if deviceTypePriority(string(inv.Devices[id].Type)) != 0 {
			t.Errorf("expected pdu/cdu in first two positions, got %q", inv.Devices[id].Type)
		}
	}
	if deviceTypePriority(string(inv.Devices[ids[len(ids)-1]].Type)) != 2 {
		t.Errorf("expected switch in last position, got %q", inv.Devices[ids[len(ids)-1]].Type)
	}
}

// TestGroupDevicesByZone verifies devices are partitioned into bottom, middle,
// and top zones by hardware type.
//
// Why it matters: zone partitioning drives top-down vs bottom-up rack fill, so
// power gear, compute, and switches must land in the right zone.
// Inputs: a pdu, cdu, node, and switch. Outputs: bottom (2), middle (1), top (1).
// Data choice: one device per zone with two in the bottom proves both the
// multi-member and single-member partitions.
func TestGroupDevicesByZone(t *testing.T) {
	inv := &devicetypes.Inventory{Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType)}

	pduID, cduID, nodeID, switchID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu")}
	inv.Devices[cduID] = &devicetypes.CaniDeviceType{ID: cduID, Type: devicetypes.Type("cdu")}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node")}
	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch")}

	bottom, middle, top := groupDevicesByZone(inv, []uuid.UUID{pduID, cduID, nodeID, switchID})

	if len(bottom) != 2 {
		t.Errorf("expected 2 bottom devices, got %d", len(bottom))
	}
	if len(middle) != 1 {
		t.Errorf("expected 1 middle device, got %d", len(middle))
	}
	if len(top) != 1 {
		t.Errorf("expected 1 top device, got %d", len(top))
	}
}

// TestRackPositionOrdering verifies config-group parenting places a PDU at the
// rack bottom, a switch at the top, and a node just below the switch.
//
// Why it matters: this is the end-to-end placement contract operators expect, so
// a mixed device set must fill from both ends of the rack toward the middle.
// Inputs: a 48U rack and a switch/node/PDU group wired via
// assignConfigGroupParenting. Outputs: PDU at U1, switch at U48, node at U47.
// Data choice: a full 48U rack with one device per zone makes the exact
// bottom/top/below-top positions unambiguous.
func TestRackPositionOrdering(t *testing.T) {
	resetRackPositionStates()
	inv := &devicetypes.Inventory{
		Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &devicetypes.CaniRackType{ID: rackID, Name: "Rack-001", UHeight: 48, Devices: []uuid.UUID{}}

	switchID, nodeID, pduID := uuid.New(), uuid.New(), uuid.New()
	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch"), Slug: "switch-1u"}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node"), Slug: "node-1u"}
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu"), Slug: "pdu-1u"}

	assignConfigGroupParenting(inv,
		map[string][]uuid.UUID{"0100": {rackID}},
		map[string][]uuid.UUID{"0200": {switchID, nodeID, pduID}})

	if pos := inv.Devices[pduID].RackPosition; pos != 1 {
		t.Errorf("PDU position = %d, want 1 (bottom)", pos)
	}
	if pos := inv.Devices[switchID].RackPosition; pos != 48 {
		t.Errorf("switch position = %d, want 48 (top)", pos)
	}
	if pos := inv.Devices[nodeID].RackPosition; pos != 47 {
		t.Errorf("node position = %d, want 47 (below switch)", pos)
	}
}

// TestRackZonesFillCorrectly verifies a small rack fills bottom-up for power
// gear and top-down for switches, with compute below the switches.
//
// Why it matters: the zone fill must pack each end without collision so a dense
// rack lays out deterministically.
// Inputs: a 10U rack and two PDUs, two switches, and a node via
// assignConfigGroupParenting. Outputs: PDUs at U1/U2, switches at U10/U9, node
// at U8.
// Data choice: a tight 10U rack with five devices forces adjacent placements
// that expose any off-by-one in either fill direction.
func TestRackZonesFillCorrectly(t *testing.T) {
	resetRackPositionStates()
	inv := &devicetypes.Inventory{
		Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &devicetypes.CaniRackType{ID: rackID, Name: "Rack-001", UHeight: 10, Devices: []uuid.UUID{}}

	pdu1, pdu2, sw1, sw2, node1 := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Devices[pdu1] = &devicetypes.CaniDeviceType{ID: pdu1, Type: devicetypes.Type("pdu")}
	inv.Devices[pdu2] = &devicetypes.CaniDeviceType{ID: pdu2, Type: devicetypes.Type("cdu")}
	inv.Devices[sw1] = &devicetypes.CaniDeviceType{ID: sw1, Type: devicetypes.Type("switch")}
	inv.Devices[sw2] = &devicetypes.CaniDeviceType{ID: sw2, Type: devicetypes.Type("mgmt-switch")}
	inv.Devices[node1] = &devicetypes.CaniDeviceType{ID: node1, Type: devicetypes.Type("node")}

	assignConfigGroupParenting(inv,
		map[string][]uuid.UUID{"0100": {rackID}},
		map[string][]uuid.UUID{"0200": {pdu1, pdu2, sw1, sw2, node1}})

	checks := []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"pdu1", pdu1, 1}, {"pdu2", pdu2, 2},
		{"node", node1, 8},
		{"sw1", sw1, 10}, {"sw2", sw2, 9},
	}
	for _, c := range checks {
		if pos := inv.Devices[c.id].RackPosition; pos != c.want {
			t.Errorf("%s position = %d, want %d", c.name, pos, c.want)
		}
	}
}

// TestLinkDeviceToRack verifies linking tolerates a missing device or a missing
// rack without panicking.
//
// Why it matters: stale IDs can appear during incremental imports, so linking
// must degrade safely rather than dereference a nil device or rack.
// Inputs: an empty inventory with a random device ID, and an inventory with a
// device but a missing rack ID. Outputs: no panic.
// Data choice: the two missing-target cases isolate the device-nil and rack-nil
// guards independently.
func TestLinkDeviceToRack(t *testing.T) {
	t.Run("missing device does not panic", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
			Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
		}
		linkDeviceToRack(inv, uuid.New(), uuid.New(), zoneMiddle)
	})

	t.Run("missing rack sets parent without panic", func(t *testing.T) {
		devID := uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{devID: {ID: devID, Name: "srv"}},
			Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
		}
		linkDeviceToRack(inv, devID, uuid.New(), zoneMiddle)
	})
}

// --- Config group and parenting helpers ---

// TestFindParentRackIDs verifies only racks in the 01XX config group are
// returned as parent rack IDs.
//
// Why it matters: devices parent to racks in the rack group, so the helper must
// select 01XX racks and ignore other groups.
// Inputs: an empty map, a 0100 group with two racks, and a 0200 group. Outputs:
// the count of parent rack IDs.
// Data choice: 0100 with two racks proves selection; the 0200-only map proves
// non-rack groups are excluded.
func TestFindParentRackIDs(t *testing.T) {
	rack1, rack2, device1 := uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name         string
		racksByGroup map[string][]uuid.UUID
		wantLen      int
	}{
		{"empty map", map[string][]uuid.UUID{}, 0},
		{"racks in 01XX group", map[string][]uuid.UUID{"0100": {rack1, rack2}}, 2},
		{"racks in non-01XX group only", map[string][]uuid.UUID{"0200": {device1}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findParentRackIDs(tt.racksByGroup); len(got) != tt.wantLen {
				t.Errorf("findParentRackIDs() returned %d IDs, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// TestShouldLinkToRacks verifies a config group should link to racks only when
// it is a non-rack, well-formed group.
//
// Why it matters: rack-group devices are the racks themselves, so only other
// groups (devices, cables) should be parented to racks.
// Inputs: rack, device, cable, and malformed group strings. Outputs: a boolean.
// Data choice: 0100 (false), 0200/0300/0900 (true), and short/empty strings
// (false) cover the rack-group exclusion and the malformed guards.
func TestShouldLinkToRacks(t *testing.T) {
	tests := []struct {
		group string
		want  bool
	}{
		{"0100", false}, {"0200", true}, {"0300", true},
		{"0900", true}, {"1", false}, {"", false},
	}
	for _, tt := range tests {
		t.Run(tt.group, func(t *testing.T) {
			if got := shouldLinkToRacks(tt.group); got != tt.want {
				t.Errorf("shouldLinkToRacks(%q) = %t, want %t", tt.group, got, tt.want)
			}
		})
	}
}

// TestGetConfigGroupPrefixTransform verifies the config-group prefix is the
// first two characters, or empty for malformed groups.
//
// Why it matters: prefixes drive rack/device/cable grouping decisions, so a
// four-digit group must yield its two-digit prefix and short inputs must yield
// empty.
// Inputs: four-digit groups and short/empty strings. Outputs: the prefix string.
// Data choice: 0100/0200/0315 prove the two-char slice; "1" and "" prove the
// too-short guard.
func TestGetConfigGroupPrefixTransform(t *testing.T) {
	tests := []struct{ input, want string }{
		{"0100", "01"}, {"0200", "02"}, {"0315", "03"},
		{"1", ""}, {"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := getConfigGroupPrefix(tt.input); got != tt.want {
				t.Errorf("getConfigGroupPrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- Step info and field mapping tests ---

// TestBuildTransformStepInfo verifies step-through display info reports the
// hardware type, record quantity, and created items for racks and devices.
//
// Why it matters: step mode shows operators what each row produced, so the info
// must reflect the type and the created objects.
// Inputs: a quantity-2 record with one created rack, and the same with one
// created device. Outputs: a step-info struct.
// Data choice: one created item per case with Quantity 2 proves the display
// reads both the record quantity and the created-items list.
func TestBuildTransformStepInfo(t *testing.T) {
	rec := import_.CsvRecord{PartNumber: "FAKE-PN", Description: "Test Device", Quantity: 2, ConfigGroup: "0300"}

	t.Run("rack items", func(t *testing.T) {
		created := CreatedItems{Racks: []*devicetypes.CaniRackType{{ID: uuid.New(), Name: "Rack-01"}}}
		info := buildTransformStepInfo(rec, "rack", created)

		if info.HwType != "rack" {
			t.Errorf("HwType = %q, want %q", info.HwType, "rack")
		}
		if info.Quantity != 2 {
			t.Errorf("Quantity = %d, want 2", info.Quantity)
		}
		if len(info.CreatedItems) != 1 {
			t.Errorf("CreatedItems = %d, want 1", len(info.CreatedItems))
		}
	})

	t.Run("device items", func(t *testing.T) {
		created := CreatedItems{Devices: []*devicetypes.CaniDeviceType{{ID: uuid.New(), Name: "Server-01"}}}
		info := buildTransformStepInfo(rec, "node", created)

		if info.HwType != "node" {
			t.Errorf("HwType = %q, want %q", info.HwType, "node")
		}
		if len(info.CreatedItems) != 1 {
			t.Errorf("CreatedItems = %d, want 1", len(info.CreatedItems))
		}
	})
}

// TestBuildFieldMappings verifies field mappings include a ConfigGroup row only
// when the record has one.
//
// Why it matters: step mode shows the source-to-target field mapping per row, so
// an absent ConfigGroup must drop its mapping rather than show a blank.
// Inputs: a record with ConfigGroup 0200 and one without. Outputs: four mappings
// vs three, with the expected source/target labels.
// Data choice: asserting the PartNumber/Name/ConfigGroup/HardwareType order and
// the 4-vs-3 count proves the conditional ConfigGroup mapping.
func TestBuildFieldMappings(t *testing.T) {
	t.Run("with config group", func(t *testing.T) {
		rec := import_.CsvRecord{PartNumber: "FAKE-PN", Description: "Some Device Description", ConfigGroup: "0200"}
		mappings := buildFieldMappings(rec, "switch", CreatedItems{})

		if len(mappings) != 4 {
			t.Fatalf("expected 4 mappings, got %d", len(mappings))
		}
		if mappings[0].SourceField != "PartNumber" {
			t.Errorf("first mapping source = %q, want PartNumber", mappings[0].SourceField)
		}
		if mappings[1].TargetField != "Name" {
			t.Errorf("second mapping target = %q, want Name", mappings[1].TargetField)
		}
		if mappings[2].SourceField != "ConfigGroup" {
			t.Errorf("third mapping source = %q, want ConfigGroup", mappings[2].SourceField)
		}
		if mappings[3].TargetField != "HardwareType" {
			t.Errorf("fourth mapping target = %q, want HardwareType", mappings[3].TargetField)
		}
	})

	t.Run("without config group", func(t *testing.T) {
		rec := import_.CsvRecord{PartNumber: "FAKE-PN", Description: "Device"}
		mappings := buildFieldMappings(rec, "node", CreatedItems{})

		if len(mappings) != 3 {
			t.Fatalf("expected 3 mappings without ConfigGroup, got %d", len(mappings))
		}
	})
}

// --- Metadata tests ---

// TestBuildProviderMetadata verifies provider metadata records the source CSV,
// part number, and config group under an "example" key.
//
// Why it matters: the example provider stamps origin metadata on every item so a
// later export can trace it back, so these fields must be captured.
// Inputs: commands.CsvFlag "test.csv" and a record with PartNumber/ConfigGroup.
// Outputs: an "example" metadata map with Source, PartNumber, ConfigGroup.
// Data choice: a set CsvFlag plus a concrete part number and group prove all
// three provenance fields are wired through.
func TestBuildProviderMetadata(t *testing.T) {
	commands.CsvFlag = "test.csv"
	t.Cleanup(func() { commands.CsvFlag = "" })

	rec := import_.CsvRecord{PartNumber: "P12345", ConfigGroup: "0200"}
	meta := buildProviderMetadata(rec)

	example, ok := meta["example"].(map[string]any)
	if !ok {
		t.Fatal("expected 'example' key in metadata")
	}
	if example["Source"] != "test.csv" {
		t.Errorf("Source = %v, want %q", example["Source"], "test.csv")
	}
	if example["PartNumber"] != "P12345" {
		t.Errorf("PartNumber = %v, want %q", example["PartNumber"], "P12345")
	}
	if example["ConfigGroup"] != "0200" {
		t.Errorf("ConfigGroup = %v, want %q", example["ConfigGroup"], "0200")
	}
}

// TestInitInventoryMaps verifies inventory map initialization creates the
// Locations/Racks/Devices/Cables maps when nil and leaves populated maps intact.
//
// Why it matters: the transform writes into these maps, so they must exist even
// when callers pass a zero-value inventory, but re-initializing must never drop
// an existing entry during an incremental import.
// Inputs: an empty inventory and one with a pre-populated Racks map. Outputs:
// non-nil maps, or the preserved entry.
// Data choice: a pre-seeded rack proves the preserve path while the empty
// inventory proves the create path.
func TestInitInventoryMaps(t *testing.T) {
	t.Run("initializes nil maps", func(t *testing.T) {
		inv := devicetypes.Inventory{}
		initInventoryMaps(&inv)
		if inv.Locations == nil {
			t.Error("expected Locations map to be initialized")
		}
		if inv.Racks == nil {
			t.Error("expected Racks map to be initialized")
		}
		if inv.Devices == nil {
			t.Error("expected Devices map to be initialized")
		}
		if inv.Cables == nil {
			t.Error("expected Cables map to be initialized")
		}
	})

	t.Run("preserves existing maps", func(t *testing.T) {
		rackID := uuid.New()
		inv := devicetypes.Inventory{
			Racks: map[uuid.UUID]*devicetypes.CaniRackType{rackID: {Name: "existing"}},
		}
		initInventoryMaps(&inv)
		if _, ok := inv.Racks[rackID]; !ok {
			t.Error("initInventoryMaps overwrote existing Racks map")
		}
	})
}

// --- Summary tests ---

// TestBuildTransformSummary verifies the transform summary lists rack names and
// groups devices under their parent rack.
//
// Why it matters: the summary is the operator-facing rollup after an import, so
// it must name each rack and attribute devices to the right one.
// Inputs: an inventory with one rack and a device parented to it. Outputs: a
// summary with one rack name and one device under "Rack-001".
// Data choice: a single parented device is the smallest input proving the
// rack-to-device grouping.
func TestBuildTransformSummary(t *testing.T) {
	rackID, deviceID := uuid.New(), uuid.New()
	inv := &devicetypes.Inventory{
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{rackID: {ID: rackID, Name: "Rack-001"}},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{deviceID: {ID: deviceID, Name: "Server-001", Parent: rackID}},
	}

	summary := BuildTransformSummary(inv)
	if len(summary.RackNames) != 1 {
		t.Errorf("RackNames len = %d, want 1", len(summary.RackNames))
	}
	if devices, ok := summary.DevicesByRack["Rack-001"]; !ok || len(devices) != 1 {
		t.Errorf("expected 1 device in Rack-001, got %v", summary.DevicesByRack)
	}
}

// TestBuildNewItemsSummary verifies the new-items summary groups created devices
// under their parent rack, filing parentless devices under an empty key.
//
// Why it matters: after a partial import the operator needs to see which new
// devices landed in which rack and which are orphaned.
// Inputs: created items with a rack-parented device, and created items with an
// unparented device. Outputs: a summary keyed by rack name, or by empty string.
// Data choice: one parented and one orphan device prove both the named-rack
// grouping and the empty-key fallback.
func TestBuildNewItemsSummary(t *testing.T) {
	t.Run("with parent rack", func(t *testing.T) {
		rackID := uuid.New()
		inv := &devicetypes.Inventory{
			Racks: map[uuid.UUID]*devicetypes.CaniRackType{rackID: {ID: rackID, Name: "Rack-001"}},
		}
		created := CreatedItems{
			Racks:   []*devicetypes.CaniRackType{{ID: rackID, Name: "Rack-001"}},
			Devices: []*devicetypes.CaniDeviceType{{Name: "Server-001", Parent: rackID}},
		}
		summary := buildNewItemsSummary(created, inv)

		if len(summary.RackNames) != 1 || summary.RackNames[0] != "Rack-001" {
			t.Errorf("RackNames = %v, want [Rack-001]", summary.RackNames)
		}
		if devices, ok := summary.DevicesByRack["Rack-001"]; !ok || len(devices) != 1 {
			t.Errorf("expected 1 device in Rack-001, got %v", summary.DevicesByRack)
		}
	})

	t.Run("orphan device", func(t *testing.T) {
		inv := &devicetypes.Inventory{Racks: map[uuid.UUID]*devicetypes.CaniRackType{}}
		created := CreatedItems{Devices: []*devicetypes.CaniDeviceType{{Name: "Orphan-001"}}}
		summary := buildNewItemsSummary(created, inv)

		if devices, ok := summary.DevicesByRack[""]; !ok || len(devices) != 1 {
			t.Errorf("expected 1 orphan device under empty key, got %v", summary.DevicesByRack)
		}
	})
}

// --- Utility function tests ---

// TestGenerateName verifies a name is built from a description, truncated to 30
// characters and suffixed with an ordinal only when more than one item shares it.
//
// Why it matters: device and rack names must be unique and length-bounded, so
// bulk rows get -NNN suffixes and long descriptions are clipped.
// Inputs: short and long descriptions with various index/total values. Outputs:
// the name string.
// Data choice: single vs first/second-of-three prove suffixing; the 48-char
// description with total 1 and 2 proves truncation with and without a suffix.
func TestGenerateName(t *testing.T) {
	tests := []struct {
		name, description string
		index, total      int
		want              string
	}{
		{"single item", "HPE Server", 0, 1, "HPE Server"},
		{"first of three", "HPE Server", 0, 3, "HPE Server-001"},
		{"second of three", "HPE Server", 1, 3, "HPE Server-002"},
		{"truncates long name single", "HPE ProLiant DL380 Gen11 High Performance Server", 0, 1, "HPE ProLiant DL380 Gen11 High "},
		{"truncates long name multi", "HPE ProLiant DL380 Gen11 High Performance Server", 0, 2, "HPE ProLiant DL380 Gen11 High -001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateName(tt.description, tt.index, tt.total); got != tt.want {
				t.Errorf("generateName(%q, %d, %d) = %q, want %q",
					tt.description, tt.index, tt.total, got, tt.want)
			}
		})
	}
}

// TestSlugify verifies a string is slugified to lowercase with spaces becoming
// dashes and special characters removed.
//
// Why it matters: slugs key library lookups and URLs, so the transform must
// produce stable, lowercase, punctuation-free slugs.
// Inputs: mixed-case strings with spaces, dashes, and special characters.
// Outputs: the slug.
// Data choice: caps, existing dashes, "Special!@#$Characters", empty, and
// multi-space runs cover lowering, dash mapping, stripping, and the empty case.
func TestSlugify(t *testing.T) {
	tests := []struct{ input, want string }{
		{"HPE Cat6 Cable", "hpe-cat6-cable"},
		{"Simple", "simple"},
		{"ALLCAPS", "allcaps"},
		{"with-dashes", "with-dashes"},
		{"Special!@#$Characters", "specialcharacters"},
		{"", ""},
		{"spaces  and   tabs", "spaces--and---tabs"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := slugify(tt.input); got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestTruncateName verifies a name longer than 30 characters is clipped to 30,
// while shorter names pass through unchanged.
//
// Why it matters: downstream systems bound name length, so the transform must
// clip without altering compliant names.
// Inputs: short, exactly-30, over-30, and empty strings. Outputs: the (possibly
// clipped) name.
// Data choice: the exactly-30 and 31-char cases pin the boundary; empty proves
// the no-op.
func TestTruncateName(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{"short string", "Hello", "Hello"},
		{"exactly 30 chars", "123456789012345678901234567890", "123456789012345678901234567890"},
		{"over 30 chars", "1234567890123456789012345678901", "123456789012345678901234567890"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateName(tt.input); got != tt.want {
				t.Errorf("truncateName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- Library-backed classification and creation ---

// TestClassifyRecords_FromLibrary verifies a part number resolved by the device
// library drives classification, including the default device branch.
//
// Why it matters: the library is the authoritative source of hardware type, so
// a recognized part number must take precedence over description heuristics, and
// any library type outside the explicit rack/cable buckets must still land as a
// device rather than being dropped.
// Inputs: one record whose part number resolves to a mgmt-switch in the library.
// Outputs: exactly one classified device.
// Data choice: FG-4401F is a real Fortinet entry typed "mgmt-switch", which is
// not in the explicit case list, so it exercises the default-to-device branch.
func TestClassifyRecords_FromLibrary(t *testing.T) {
	got, err := classifyRecords([]import_.CsvRecord{
		{PartNumber: "FG-4401F", Description: "ignored description", Quantity: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.devices) != 1 {
		t.Errorf("expected 1 device, got %d (racks=%d cables=%d)", len(got.devices), len(got.racks), len(got.cables))
	}
}

// TestCreateRack_FromLibrary verifies a rack record whose part number resolves
// in the device registry inherits that entry's slug and U-height.
//
// Why it matters: when a record's part number is known, the created rack should
// reflect library specs instead of defaulting, so downstream placement uses the
// correct height.
// Inputs: a rack record carrying the real part number P67287-B21. Outputs: a
// rack with the library slug hpe-xd670 and U-height 5.
// Data choice: P67287-B21 is a real registry entry with a non-default U-height,
// so a successful lookup is observably different from the 48U fallback.
func TestCreateRack_FromLibrary(t *testing.T) {
	inv := &devicetypes.Inventory{Racks: map[uuid.UUID]*devicetypes.CaniRackType{}}
	id := uuid.New()
	rec := import_.CsvRecord{PartNumber: "P67287-B21", Description: "Custom Rack", ConfigGroup: "0100"}
	rack := createRack(inv, id, "r", rec, map[string][]uuid.UUID{})
	if rack.Slug != "hpe-xd670" {
		t.Errorf("Slug = %q, want %q", rack.Slug, "hpe-xd670")
	}
	if rack.UHeight != 5 {
		t.Errorf("UHeight = %d, want 5", rack.UHeight)
	}
}

// TestBuildDeviceFromRecord_FromLibrary verifies a device record with a known
// part number is populated from the library entry.
//
// Why it matters: library population fills in manufacturer, model, interfaces,
// and U-height that the bare CSV record lacks, so a recognized part number must
// trigger that copy.
// Inputs: a device record carrying part number P67287-B21. Outputs: a device
// with the library slug hpe-xd670 and U-height 5.
// Data choice: P67287-B21 resolves to hpe-xd670 (U-height 5), making the
// library copy observable against the record's own empty specs.
func TestBuildDeviceFromRecord_FromLibrary(t *testing.T) {
	id := uuid.New()
	rec := import_.CsvRecord{PartNumber: "P67287-B21", Description: "HPE XD670"}
	device := buildDeviceFromRecord(id, "xd670-001", rec, "blade")
	if device.Slug != "hpe-xd670" {
		t.Errorf("Slug = %q, want %q", device.Slug, "hpe-xd670")
	}
	if device.UHeight != 5 {
		t.Errorf("UHeight = %d, want 5", device.UHeight)
	}
}

// TestGetDeviceUHeight_FromLibrary verifies U-height resolution by part number
// and by slug against the device library.
//
// Why it matters: rack placement needs each device's true height; the resolver
// must try part number first and fall back to slug so devices identified either
// way are sized correctly.
// Inputs: one device keyed only by part number, another keyed only by slug.
// Outputs: U-height 5 for both.
// Data choice: P67287-B21 and hpe-xd670 are the part number and slug of the same
// real 5U entry, isolating each lookup path while expecting the same height.
func TestGetDeviceUHeight_FromLibrary(t *testing.T) {
	t.Run("by part number", func(t *testing.T) {
		d := &devicetypes.CaniDeviceType{PartNumber: "P67287-B21"}
		if got := getDeviceUHeight(d); got != 5 {
			t.Errorf("getDeviceUHeight = %d, want 5", got)
		}
	})
	t.Run("by slug", func(t *testing.T) {
		d := &devicetypes.CaniDeviceType{Slug: "hpe-xd670"}
		if got := getDeviceUHeight(d); got != 5 {
			t.Errorf("getDeviceUHeight = %d, want 5", got)
		}
	})
}

// --- Placement edge cases ---

// TestGroupDevicesByZone_NilDevice verifies an ID absent from inventory is
// treated as a middle-zone device.
//
// Why it matters: stale or dangling device IDs must not crash zone partitioning;
// the safe default is the middle zone so the device still gets a placement
// attempt.
// Inputs: a device ID that is not present in the inventory map. Outputs: that ID
// sorted into the middle zone, with bottom and top empty.
// Data choice: a never-registered UUID guarantees the nil-lookup branch without
// depending on any device state.
func TestGroupDevicesByZone_NilDevice(t *testing.T) {
	inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{}}
	bottom, middle, top := groupDevicesByZone(inv, []uuid.UUID{uuid.New()})
	if len(bottom) != 0 || len(top) != 0 || len(middle) != 1 {
		t.Errorf("unknown device zoning = bottom %d, middle %d, top %d; want 0,1,0", len(bottom), len(middle), len(top))
	}
}

// TestDistributeDevicesToRacks_NilRack verifies a target rack ID missing from
// inventory is skipped without panicking.
//
// Why it matters: rack distribution must tolerate a stale rack reference rather
// than dereference a nil rack, so a bad ID simply yields no placement.
// Inputs: one device and a rack-ID list whose only entry is absent from the
// inventory. Outputs: the device left unplaced (RackPosition 0) and no panic.
// Data choice: a single missing rack ID isolates the nil-rack guard in both the
// state-init loop and the link step.
func TestDistributeDevicesToRacks_NilRack(t *testing.T) {
	resetRackPositionStates()
	devID := uuid.New()
	inv := &devicetypes.Inventory{
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{devID: {ID: devID, Type: devicetypes.Type("node")}},
	}
	distributeDevicesToRacks(inv, []uuid.UUID{devID}, []uuid.UUID{uuid.New()})
	if inv.Devices[devID].RackPosition != 0 {
		t.Errorf("expected device unplaced, got position %d", inv.Devices[devID].RackPosition)
	}
}

// TestAssignRackPosition_Overflow verifies devices that exceed remaining rack
// space are left unplaced in both the top-down and bottom-up fill directions.
//
// Why it matters: a full rack must refuse further devices rather than assign
// invalid or overlapping U positions, protecting downstream layout integrity.
// Inputs: a 1U rack and successive switch (top) and PDU (bottom) placements.
// Outputs: the first switch placed at U1, the second switch and the PDU left at
// position 0.
// Data choice: a 1U rack is the smallest space that fits exactly one device,
// forcing the very next placement in each direction to hit the won't-fit guard.
func TestAssignRackPosition_Overflow(t *testing.T) {
	resetRackPositionStates()
	rack := &devicetypes.CaniRackType{ID: uuid.New(), UHeight: 1}

	first := &devicetypes.CaniDeviceType{ID: uuid.New(), Type: devicetypes.Type("switch")}
	assignRackPosition(first, rack, zoneTop)
	if first.RackPosition != 1 {
		t.Fatalf("first switch position = %d, want 1", first.RackPosition)
	}

	second := &devicetypes.CaniDeviceType{ID: uuid.New(), Type: devicetypes.Type("switch")}
	assignRackPosition(second, rack, zoneTop)
	if second.RackPosition != 0 {
		t.Errorf("second switch should not fit, got position %d", second.RackPosition)
	}

	pdu := &devicetypes.CaniDeviceType{ID: uuid.New(), Type: devicetypes.Type("pdu")}
	assignRackPosition(pdu, rack, zoneBottom)
	if pdu.RackPosition != 0 {
		t.Errorf("pdu should not fit in full rack, got position %d", pdu.RackPosition)
	}
}

// --- Transform entry: step mode and cable errors ---

// TestTransform_StepModeRack verifies Transform drives the rack step-through
// prompt when step mode is enabled and a rack is created.
//
// Why it matters: step mode is the operator's interactive review path, so each
// created rack must pause for confirmation rather than streaming past unseen.
// Inputs: a single rack record, StepMode enabled, and a stdin holding one
// newline. Outputs: one rack with no error.
// Data choice: exactly one record keeps the run to a single prompt, which is the
// most a fresh-per-call bufio reader over a pipe can satisfy without EOF; P9K58A
// is a real rack part number so classification routes it through the rack pass.
func TestTransform_StepModeRack(t *testing.T) {
	setupTransformTest(t, []import_.CsvRecord{
		{PartNumber: "P9K58A", Description: "HPE 48U Rack", Quantity: 1, ConfigGroup: "0100"},
	})
	config.Cfg.StepMode = true
	config.Cfg.NoColor = true
	withStdin(t, "\n")

	result, err := Transform(devicetypes.Inventory{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Racks) != 1 {
		t.Errorf("expected 1 rack, got %d", len(result.Racks))
	}
}

// TestTransform_StepModeDevice verifies Transform drives the device step-through
// prompt and resolves device hardware type from the library.
//
// Why it matters: the device pass must both pause for review in step mode and
// classify a recognized part number via the library so the created device is
// correctly typed.
// Inputs: a single device record keyed by a real part number, StepMode enabled,
// and a stdin holding one newline. Outputs: one device with no error.
// Data choice: P67287-B21 resolves to a blade in the library, exercising the
// library hardware-type branch, and one record keeps the run to a single prompt.
func TestTransform_StepModeDevice(t *testing.T) {
	setupTransformTest(t, []import_.CsvRecord{
		{PartNumber: "P67287-B21", Description: "HPE XD670", Quantity: 1, ConfigGroup: "0300"},
	})
	config.Cfg.StepMode = true
	config.Cfg.NoColor = true
	withStdin(t, "\n")

	result, err := Transform(devicetypes.Inventory{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(result.Devices))
	}
}

// TestTransform_StepModeRackInterrupted verifies Transform aborts when the rack
// step prompt is interrupted.
//
// Why it matters: if the operator's input stream closes mid-review, Transform
// must stop and surface the interruption rather than silently continue past an
// unconfirmed rack.
// Inputs: a single rack record, StepMode enabled, and an empty stdin that
// returns EOF on the first read. Outputs: a non-nil error from Transform.
// Data choice: empty stdin is the minimal way to force the prompt's read to fail
// immediately, exercising the rack-step interrupt branch.
func TestTransform_StepModeRackInterrupted(t *testing.T) {
	setupTransformTest(t, []import_.CsvRecord{
		{PartNumber: "P9K58A", Description: "HPE 48U Rack", Quantity: 1, ConfigGroup: "0100"},
	})
	config.Cfg.StepMode = true
	config.Cfg.NoColor = true
	withStdin(t, "")

	if _, err := Transform(devicetypes.Inventory{}); err == nil {
		t.Fatal("expected error when rack step prompt is interrupted")
	}
}

// TestTransform_StepModeDeviceInterrupted verifies Transform aborts when the
// device step prompt is interrupted.
//
// Why it matters: an interrupted input stream during the device review must halt
// the import so a half-reviewed device set is never committed.
// Inputs: a single device record, StepMode enabled, and an empty stdin that
// returns EOF on the first read. Outputs: a non-nil error from Transform.
// Data choice: P67287-B21 routes through the device pass, and empty stdin forces
// the device-step prompt read to fail, exercising the device-step interrupt branch.
func TestTransform_StepModeDeviceInterrupted(t *testing.T) {
	setupTransformTest(t, []import_.CsvRecord{
		{PartNumber: "P67287-B21", Description: "HPE XD670", Quantity: 1, ConfigGroup: "0300"},
	})
	config.Cfg.StepMode = true
	config.Cfg.NoColor = true
	withStdin(t, "")

	if _, err := Transform(devicetypes.Inventory{}); err == nil {
		t.Fatal("expected error when device step prompt is interrupted")
	}
}

// TestTransform_CableError verifies Transform surfaces a cable-pass failure as a
// wrapped error.
//
// Why it matters: a cable referencing a missing device must abort the import so
// the operator fixes the data rather than receiving a partial, inconsistent
// inventory.
// Inputs: one explicit cable record naming devices that do not exist. Outputs: a
// non-nil error from Transform.
// Data choice: both endpoints are ghosts so the cable pass fails on the first
// lookup, exercising the transformCables error-wrap path in Transform.
func TestTransform_CableError(t *testing.T) {
	setupTransformTest(t, []import_.CsvRecord{
		{Description: "link", Quantity: 1, SourceDevice: "ghost-a", SourcePort: "e0", DestDevice: "ghost-b", DestPort: "e0"},
	})
	if _, err := Transform(devicetypes.Inventory{}); err == nil {
		t.Fatal("expected error from transformCables for missing devices")
	}
}
