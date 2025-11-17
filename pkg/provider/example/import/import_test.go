package import_

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/commands"
	"github.com/google/uuid"
)

func TestImport(t *testing.T) {
	t.Cleanup(func() {
		commands.FileFlag = ""
		commands.CsvFlag = ""
	})

	tests := []struct {
		name      string
		fileFlag  string
		expectErr bool
	}{
		{
			name:      "no file specified returns nil",
			fileFlag:  "",
			expectErr: false,
		},
		{
			name:      "nonexistent file returns error",
			fileFlag:  "nonexistent.json",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands.CsvFlag = ""
			commands.FileFlag = tt.fileFlag
			inv := &devicetypes.Inventory{}
			err := Import(nil, nil, inv)
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMergeYAMLInventory(t *testing.T) {
	locID := uuid.New()
	rackID := uuid.New()
	deviceID := uuid.New()
	cableID := uuid.New()
	slotDeviceID := uuid.New()

	tests := []struct {
		name      string
		dst       *devicetypes.Inventory
		src       *yamlInventory
		expectErr bool
	}{
		{
			name: "merges all sections including legacy rack slots",
			dst:  &devicetypes.Inventory{},
			src: &yamlInventory{
				Locations: map[string]*devicetypes.CaniLocationType{
					locID.String(): {Name: "DC1"},
				},
				Racks: map[string]*yamlRackType{
					rackID.String(): {
						ID:            rackID,
						Name:          "Rack-01",
						UHeight:       42,
						OccupiedSlots: map[int]uuid.UUID{1: slotDeviceID},
					},
				},
				Devices: map[string]*devicetypes.CaniDeviceType{
					deviceID.String(): {Name: "server01"},
				},
				Cables: map[string]*devicetypes.CaniCableType{
					cableID.String(): {CableType: "cat6"},
				},
			},
			expectErr: false,
		},
		{
			name: "invalid location UUID returns error",
			dst:  &devicetypes.Inventory{},
			src: &yamlInventory{
				Locations: map[string]*devicetypes.CaniLocationType{
					"not-a-uuid": {Name: "Bad"},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mergeYAMLInventory(tt.dst, tt.src)
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr {
				if _, ok := tt.dst.Locations[locID]; !ok {
					t.Errorf("expected location %s to be merged", locID)
				}
				if _, ok := tt.dst.Racks[rackID]; !ok {
					t.Errorf("expected rack %s to be merged", rackID)
				}
				if _, ok := tt.dst.Devices[deviceID]; !ok {
					t.Errorf("expected device %s to be merged", deviceID)
				}
				if tt.dst.Devices[deviceID].ID != deviceID {
					t.Errorf("expected device ID to be set to %s", deviceID)
				}
				if _, ok := tt.dst.Cables[cableID]; !ok {
					t.Errorf("expected cable %s to be merged", cableID)
				}
			}
		})
	}
}

type fakeProvider struct{ records []CsvRecord }

func (f *fakeProvider) ClearRecords()            { f.records = nil }
func (f *fakeProvider) SetRecords(r []CsvRecord) { f.records = r }

func TestImportCSV(t *testing.T) {
	fake := &fakeProvider{}
	SetProviderGetter(func() interface {
		ClearRecords()
		SetRecords(records []CsvRecord)
	} {
		return fake
	})
	config.Cfg = &config.Config{}
	t.Cleanup(func() {
		providerGetter = nil
		config.Cfg = nil
	})

	tests := []struct {
		name      string
		csvPath   string
		expectErr bool
	}{
		{
			name:      "valid CSV",
			csvPath:   "../../../../testdata/fixtures/example/cables.csv",
			expectErr: false,
		},
		{
			name:      "invalid CSV path",
			csvPath:   "nonexistent.csv",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands.CsvFlag = tt.csvPath
			err := ImportCSV(nil, nil, nil)
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestFormatRawCSVRecord(t *testing.T) {

	tests := []struct {
		name     string
		record   CsvRecord
		expected string
	}{
		{
			name: "device record single quantity",
			record: CsvRecord{
				Description: "Network Card",
				Quantity:    1,
			},
			expected: "[ Network Card]",
		},
		{
			name: "device record greater than one",
			record: CsvRecord{
				Description: "Network Card",
				Quantity:    5,
			},
			expected: "[ Network Card qty=5]",
		},
		{
			name: "ConfigGroup has value",
			record: CsvRecord{
				ConfigGroup: "sample_config_group_name",
			},
			expected: "[  grp=sample_config_group_name]",
		},
		{
			name: "SourceDevice has value",
			record: CsvRecord{
				SourceDevice: "switch01",
				SourcePort:   "eth0",
				DestDevice:   "server01",
				DestPort:     "eth1",
			},
			expected: "[  switch01:eth0→server01:eth1]",
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			got := formatRawCSVRecord(tt.record)
			if got != tt.expected {
				t.Errorf("formatParsedRecord() = %q, want %q", got, tt.expected)
			}
		})
	}
}
func TestFormatParsedRecord(t *testing.T) {

	tests := []struct {
		name     string
		record   CsvRecord
		expected string
	}{
		{
			name: "cable record",
			record: CsvRecord{
				SourceDevice: "switch01",
				SourcePort:   "eth0",
				DestDevice:   "server01",
				DestPort:     "eth1",
			},
			expected: "cable: switch01:eth0 ↔ server01:eth1",
		},
		{
			name: "device record quantity of one",
			record: CsvRecord{
				Description: "HPE Server",
				Quantity:    1,
			},
			expected: "device: HPE Server",
		},
		{
			name: "device record quantity greater than one",
			record: CsvRecord{
				Description: "Network Card",
				Quantity:    5,
			},
			expected: "device × 5: Network Card",
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			got := formatParsedRecord(tt.record)
			if got != tt.expected {
				t.Errorf("formatParsedRecord() = %q, want %q", got, tt.expected)
			}
		})
	}
}
