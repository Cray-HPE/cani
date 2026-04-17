package import_

import (
	"os"
	"strings"
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
		{
			name:      "valid YAML file",
			fileFlag:  "../../../../testdata/fixtures/example/inventory.yaml",
			expectErr: false,
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

func TestImportYAMLInvalidContent(t *testing.T) {
	t.Cleanup(func() {
		commands.FileFlag = ""
		commands.CsvFlag = ""
	})

	// Create a temp file with invalid YAML
	tmpDir := t.TempDir()
	badFile := tmpDir + "/bad.yaml"
	if err := os.WriteFile(badFile, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	commands.CsvFlag = ""
	commands.FileFlag = badFile
	inv := &devicetypes.Inventory{}
	err := Import(nil, nil, inv)
	if err == nil {
		t.Error("expected error for invalid YAML but got none")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("error = %q, want containing 'failed to parse YAML'", err.Error())
	}
}

func TestImportCSVRouting(t *testing.T) {
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
		commands.CsvFlag = ""
		commands.FileFlag = ""
	})

	// CSV flag takes precedence over FileFlag
	commands.CsvFlag = "../../../../testdata/fixtures/example/cables.csv"
	commands.FileFlag = "should-not-be-used.yaml"
	inv := &devicetypes.Inventory{}
	err := Import(nil, nil, inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.records) == 0 {
		t.Error("expected records to be set via CSV import")
	}
}

func TestImportCSVEmptyFile(t *testing.T) {
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
		commands.CsvFlag = ""
	})

	// empty.csv has header only → ParseCSV returns an error
	commands.CsvFlag = "../../../../testdata/fixtures/example/empty.csv"
	err := ImportCSV(nil, nil, nil)
	if err == nil {
		t.Error("expected error for header-only CSV but got none")
	}
	if !strings.Contains(err.Error(), "failed to parse CSV") {
		t.Errorf("error = %q, want containing 'failed to parse CSV'", err.Error())
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
		errMsg    string
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
			errMsg:    "invalid location UUID",
		},
		{
			name: "invalid rack UUID returns error",
			dst:  &devicetypes.Inventory{},
			src: &yamlInventory{
				Racks: map[string]*yamlRackType{
					"not-a-uuid": {Name: "Bad"},
				},
			},
			expectErr: true,
			errMsg:    "invalid rack UUID",
		},
		{
			name: "invalid device UUID returns error",
			dst:  &devicetypes.Inventory{},
			src: &yamlInventory{
				Devices: map[string]*devicetypes.CaniDeviceType{
					"not-a-uuid": {Name: "Bad"},
				},
			},
			expectErr: true,
			errMsg:    "invalid device UUID",
		},
		{
			name: "invalid cable UUID returns error",
			dst:  &devicetypes.Inventory{},
			src: &yamlInventory{
				Cables: map[string]*devicetypes.CaniCableType{
					"not-a-uuid": {CableType: "cat6"},
				},
			},
			expectErr: true,
			errMsg:    "invalid cable UUID",
		},
		{
			name: "nil sections are skipped",
			dst:  &devicetypes.Inventory{},
			src:  &yamlInventory{},
		},
		{
			name: "rack without legacy slots",
			dst:  &devicetypes.Inventory{},
			src: &yamlInventory{
				Racks: map[string]*yamlRackType{
					rackID.String(): {
						ID:      rackID,
						Name:    "Rack-02",
						UHeight: 48,
						Status:  "active",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mergeYAMLInventory(tt.dst, tt.src)
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if tt.expectErr && err != nil && tt.errMsg != "" {
				if got := err.Error(); !strings.Contains(got, tt.errMsg) {
					t.Errorf("error = %q, want containing %q", got, tt.errMsg)
				}
				return
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr {
				if _, ok := tt.dst.Locations[locID]; tt.src.Locations != nil && !ok {
					t.Errorf("expected location %s to be merged", locID)
				}
				if _, ok := tt.dst.Racks[rackID]; tt.src.Racks != nil && !ok {
					t.Errorf("expected rack %s to be merged", rackID)
				}
				if _, ok := tt.dst.Devices[deviceID]; tt.src.Devices != nil && !ok {
					t.Errorf("expected device %s to be merged", deviceID)
				}
				if d, ok := tt.dst.Devices[deviceID]; ok && d.ID != deviceID {
					t.Errorf("expected device ID to be set to %s", deviceID)
				}
				if _, ok := tt.dst.Cables[cableID]; tt.src.Cables != nil && !ok {
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
