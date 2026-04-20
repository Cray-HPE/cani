package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/google/uuid"
)

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

type fakeRecordProvider struct{ records []import_.CsvRecord }

func (f *fakeRecordProvider) GetRecords() []import_.CsvRecord { return f.records }

func TestTransformNoProviderGetter(t *testing.T) {
	old := providerGetter
	t.Cleanup(func() { providerGetter = old })
	providerGetter = nil

	_, err := Transform(devicetypes.Inventory{})
	if err == nil {
		t.Fatal("expected error when providerGetter is nil")
	}
}

func TestTransformEmptyRecords(t *testing.T) {
	old := providerGetter
	oldCfg := config.Cfg
	t.Cleanup(func() {
		providerGetter = old
		config.Cfg = oldCfg
	})

	config.Cfg = &config.Config{}
	SetProviderGetter(func() interface{ GetRecords() []import_.CsvRecord } {
		return &fakeRecordProvider{}
	})

	result, err := Transform(devicetypes.Inventory{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Racks) != 0 || len(result.Devices) != 0 || len(result.Cables) != 0 {
		t.Error("expected empty result for empty records")
	}
}

func TestTransformWithRecords(t *testing.T) {
	old := providerGetter
	oldCfg := config.Cfg
	t.Cleanup(func() {
		providerGetter = old
		config.Cfg = oldCfg
		resetRackPositionStates()
	})

	config.Cfg = &config.Config{}
	records := []import_.CsvRecord{
		{PartNumber: "FAKE-RACK-PN", Description: "48U Rack Cabinet", Quantity: 1, ConfigGroup: "0100"},
		{PartNumber: "FAKE-SRV-PN", Description: "ProLiant DL380 Server", Quantity: 2, ConfigGroup: "0300"},
	}
	SetProviderGetter(func() interface{ GetRecords() []import_.CsvRecord } {
		return &fakeRecordProvider{records: records}
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
}

func TestTransformClassifyError(t *testing.T) {
	old := providerGetter
	oldCfg := config.Cfg
	t.Cleanup(func() {
		providerGetter = old
		config.Cfg = oldCfg
	})

	config.Cfg = &config.Config{}
	records := []import_.CsvRecord{
		{PartNumber: "FAKE-UNKNOWN", Description: "", Quantity: 1},
	}
	SetProviderGetter(func() interface{ GetRecords() []import_.CsvRecord } {
		return &fakeRecordProvider{records: records}
	})

	_, err := Transform(devicetypes.Inventory{})
	if err == nil {
		t.Fatal("expected error for unclassifiable record")
	}
}

func TestBuildTransformStepInfo(t *testing.T) {
	rec := import_.CsvRecord{
		PartNumber:  "FAKE-PN",
		Description: "Test Device",
		Quantity:    2,
		ConfigGroup: "0300",
	}

	rack := &devicetypes.CaniRackType{
		ID:   uuid.New(),
		Name: "Rack-01",
	}
	device := &devicetypes.CaniDeviceType{
		ID:   uuid.New(),
		Name: "Server-01",
	}

	t.Run("rack items", func(t *testing.T) {
		created := CreatedItems{Racks: []*devicetypes.CaniRackType{rack}}
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
		created := CreatedItems{Devices: []*devicetypes.CaniDeviceType{device}}
		info := buildTransformStepInfo(rec, "node", created)

		if info.HwType != "node" {
			t.Errorf("HwType = %q, want %q", info.HwType, "node")
		}
		if len(info.CreatedItems) != 1 {
			t.Errorf("CreatedItems = %d, want 1", len(info.CreatedItems))
		}
	})
}

func TestBuildFieldMappings(t *testing.T) {
	rec := import_.CsvRecord{
		PartNumber:  "FAKE-PN",
		Description: "Some Device Description",
		ConfigGroup: "0200",
	}

	created := CreatedItems{}
	mappings := buildFieldMappings(rec, "switch", created)

	// Expect: PartNumber, Description→Name, ConfigGroup, HardwareType inferred
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
}

func TestBuildFieldMappingsNoConfigGroup(t *testing.T) {
	rec := import_.CsvRecord{
		PartNumber:  "FAKE-PN",
		Description: "Device",
	}

	mappings := buildFieldMappings(rec, "node", CreatedItems{})

	// Without ConfigGroup: PartNumber, Description→Name, HardwareType
	if len(mappings) != 3 {
		t.Fatalf("expected 3 mappings without ConfigGroup, got %d", len(mappings))
	}
}

func TestGetDeviceUHeight(t *testing.T) {
	tests := []struct {
		name   string
		device *devicetypes.CaniDeviceType
	}{
		{
			name:   "unknown part number and slug",
			device: &devicetypes.CaniDeviceType{PartNumber: "FAKE-UNKNOWN-PN", Slug: "fake-unknown-slug"},
		},
		{
			name:   "empty part number and slug",
			device: &devicetypes.CaniDeviceType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDeviceUHeight(tt.device)
			if got != 0 {
				t.Errorf("expected 0 for unknown device, got %d", got)
			}
		})
	}
}

func TestCreateRackWithoutLibraryMatch(t *testing.T) {
	inv := &devicetypes.Inventory{
		Racks: make(map[uuid.UUID]*devicetypes.CaniRackType),
	}
	racksByGroup := make(map[string][]uuid.UUID)
	id := uuid.New()

	rec := import_.CsvRecord{
		PartNumber:  "FAKE-RACK-PN",
		Description: "Custom 42U Rack",
		ConfigGroup: "0100",
	}

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

func TestLinkDeviceToRackMissingDevice(t *testing.T) {
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
		Racks:   map[uuid.UUID]*devicetypes.CaniRackType{},
	}

	// Should not panic when device doesn't exist
	linkDeviceToRack(inv, uuid.New(), uuid.New(), zoneMiddle)
}

func TestLinkDeviceToRackMissingRack(t *testing.T) {
	devID := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			devID: {ID: devID, Name: "srv"},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{},
	}

	// Should set parent but not panic when rack doesn't exist
	linkDeviceToRack(inv, devID, uuid.New(), zoneMiddle)
}
