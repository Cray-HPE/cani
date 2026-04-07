package import_

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ochami/commands"
)

type fakeProvider struct{ records []JSONDeviceRecord }

func (f *fakeProvider) ClearRecords()                   { f.records = nil }
func (f *fakeProvider) SetRecords(r []JSONDeviceRecord) { f.records = r }

func TestImport(t *testing.T) {
	fake := &fakeProvider{}
	SetProviderGetter(func() interface {
		ClearRecords()
		SetRecords(records []JSONDeviceRecord)
	} {
		return fake
	})
	t.Cleanup(func() {
		commands.JsonFileFlag = ""
		providerGetter = nil
	})

	tests := []struct {
		name         string
		jsonFileFlag string
		expectErr    bool
	}{
		{
			name:         "no file specified returns nil",
			jsonFileFlag: "",
			expectErr:    false,
		},
		{
			name:         "nonexistent file returns error",
			jsonFileFlag: "nonexistent.json",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands.JsonFileFlag = tt.jsonFileFlag
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

func TestImportOchamiDevices(t *testing.T) {
	fake := &fakeProvider{}
	SetProviderGetter(func() interface {
		ClearRecords()
		SetRecords(records []JSONDeviceRecord)
	} {
		return fake
	})
	t.Cleanup(func() {
		commands.JsonFileFlag = ""
		providerGetter = nil
	})

	tests := []struct {
		name         string
		jsonFileFlag string
		expectErr    bool
		expectCount  int
	}{
		{
			name:         "valid JSON imports records",
			jsonFileFlag: "../../../../testdata/fixtures/ochami/ochami_test_data.json",
			expectErr:    false,
			expectCount:  32,
		},
		{
			name:         "nonexistent file returns error",
			jsonFileFlag: "nonexistent.json",
			expectErr:    true,
			expectCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake.records = nil
			commands.JsonFileFlag = tt.jsonFileFlag
			inv := &devicetypes.Inventory{}
			err := ImportOchamiDevices(nil, nil, inv)
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && len(fake.records) != tt.expectCount {
				t.Errorf("expected %d records, got %d", tt.expectCount, len(fake.records))
			}
		})
	}
}
