package import_

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/commands"
	"github.com/google/uuid"
)

// TestImport verifies Import handles missing files, missing paths, and valid
// YAML inventory files through the file-based path.
//
// Why it matters: Import is the example provider entry point, so it must no-op
// cleanly without a file, fail on unreadable files, and merge valid YAML data.
// Inputs: empty FileFlag, a nonexistent path, and the inventory YAML fixture.
// Outputs: nil or non-nil errors according to the path state.
// Data choice: the fixture exercises the real YAML import path, while the empty
// and bad paths isolate the early-return and read-error branches.
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

// TestImportYAMLInvalidContent verifies Import returns a parse error for a YAML
// file with invalid syntax.
//
// Why it matters: malformed YAML should fail before any merge into inventory,
// preserving the existing inventory state.
// Inputs: a temp file containing invalid YAML. Outputs: a non-nil error naming
// YAML parsing.
// Data choice: the deliberately malformed token is the smallest local fixture
// that reaches yaml.Unmarshal's error branch.
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

// TestImportCSVRouting verifies Import gives CsvFlag precedence over FileFlag
// and stores parsed BOM records through the CSV provider path.
//
// Why it matters: CLI callers can pass both flags, and CSV imports should not be
// shadowed by a stale YAML file flag.
// Inputs: CsvFlag set to the cables fixture and FileFlag set to an unused path.
// Outputs: nil error and non-empty provider records.
// Data choice: the cables fixture has known valid BOM rows, proving the CSV
// route ran and persisted records.
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

// TestImportCSVEmptyFile verifies ImportCSV wraps the parser error for a
// header-only BOM CSV.
//
// Why it matters: a CSV with no data rows should be reported as invalid input,
// not treated as an empty successful import.
// Inputs: the empty.csv fixture containing only a header. Outputs: a non-nil
// wrapped parse error.
// Data choice: a header-only file isolates the parser's minimum-row guard.
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

// TestMergeYAMLInventory verifies mergeYAMLInventory merges each YAML section,
// migrates legacy rack slots, and rejects invalid UUID keys.
//
// Why it matters: file imports convert YAML string keys into typed inventory
// maps, so every section must land under the intended UUID and malformed keys
// must abort clearly.
// Inputs: YAML inventory structs with valid locations, racks, devices, cables,
// nil sections, and invalid keys for each section. Outputs: mutated inventory or
// an error.
// Data choice: one UUID per section proves each map merge, and the legacy slot
// map proves rack migration remains wired.
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

// TestImportCSV verifies ImportCSV stores valid BOM records and returns an error
// for an unreadable CSV path.
//
// Why it matters: BOM CSV import is the raw-record handoff to transform, so it
// must persist parsed records only after successful parsing.
// Inputs: the cables fixture and a nonexistent CSV path. Outputs: nil or
// non-nil errors.
// Data choice: the fixture proves the provider handoff path, while the missing
// file proves the wrapped open-error path.
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

// TestFormatRawCSVRecord verifies formatRawCSVRecord builds the compact raw row
// summary used by step mode.
//
// Why it matters: step mode shows this string to operators before persisting
// records, so quantity, config group, and cable endpoint details must be visible.
// Inputs: device, multi-quantity, config-group, and cable CsvRecord values.
// Outputs: formatted raw summary strings.
// Data choice: each record sets one optional display component, isolating the
// formatter branches.
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
				t.Errorf("formatRawCSVRecord() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestFormatParsedRecord verifies formatParsedRecord describes BOM records as
// cable or device rows for step mode.
//
// Why it matters: operators rely on the parsed summary to confirm the importer's
// interpretation of each raw row.
// Inputs: an explicit cable record plus single- and multi-quantity device
// records. Outputs: formatted parsed summary strings.
// Data choice: the cable record proves endpoint formatting, and the two device
// quantities prove singular and multiplication text.
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

// withStdin redirects os.Stdin to a pipe preloaded with input for the duration
// of the test, then restores the original during cleanup.
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

// fakeDcimProvider records the DCIM CSV data handed to it so tests can assert
// the import wired ClearDcimRecords/SetDcimRecords correctly.
type fakeDcimProvider struct {
	data    *DcimCSV
	isDcim  bool
	cleared bool
}

func (f *fakeDcimProvider) SetDcimRecords(d *DcimCSV) { f.data = d }
func (f *fakeDcimProvider) ClearDcimRecords()         { f.cleared = true }
func (f *fakeDcimProvider) IsDcimImport() bool        { return f.isDcim }

// newDcimProvider registers fake as the DCIM provider and clears the getter
// during cleanup, sparing each test the verbose interface literal.
func newDcimProvider(t *testing.T, fake *fakeDcimProvider) {
	t.Helper()
	SetDcimProviderGetter(func() interface {
		SetDcimRecords(data *DcimCSV)
		ClearDcimRecords()
		IsDcimImport() bool
	} {
		return fake
	})
	t.Cleanup(func() { dcimProviderGetter = nil })
}

// setBOMProvider registers fake as the BOM record provider and clears the getter
// during cleanup.
func setBOMProvider(t *testing.T, fake *fakeProvider) {
	t.Helper()
	SetProviderGetter(func() interface {
		ClearRecords()
		SetRecords(records []CsvRecord)
	} {
		return fake
	})
	t.Cleanup(func() { providerGetter = nil })
}

// TestGetProvider_ErrorsWhenUnset verifies GetProvider returns an error when no
// getter has been registered.
//
// Why it matters: the import layer depends on the parent package injecting the
// provider singleton, so a missing registration is a programming error that must
// surface as an error rather than nil-deref later.
// Inputs: providerGetter set to nil. Outputs: a non-nil error.
// Data choice: nil is the only state that trips the guard, isolating the error
// path.
func TestGetProvider_ErrorsWhenUnset(t *testing.T) {
	old := providerGetter
	t.Cleanup(func() { providerGetter = old })
	providerGetter = nil
	if _, err := GetProvider(); err == nil {
		t.Error("expected error when providerGetter is unset")
	} else if !strings.Contains(err.Error(), "providerGetter not set") {
		t.Errorf("error = %q, want providerGetter not set", err)
	}
}

// TestDcimProviderGetter verifies the DCIM provider round-trips through the
// setter/getter and that retrieval errors when unset.
//
// Why it matters: DCIM CSV import resolves the provider singleton through this
// indirection to break an import cycle, so a registered getter must return the
// same instance and a missing one must surface an error.
// Inputs: a fakeDcimProvider via the setter, then a nil getter. Outputs: the
// identical provider instance, then a non-nil error.
// Data choice: identity comparison proves the exact registered value is
// returned; nil is the only state that trips the guard.
func TestDcimProviderGetter(t *testing.T) {
	t.Run("set and get round-trips the provider", func(t *testing.T) {
		fake := &fakeDcimProvider{isDcim: true}
		newDcimProvider(t, fake)
		got, err := GetDcimProvider()
		if err != nil {
			t.Fatalf("GetDcimProvider returned unexpected error: %v", err)
		}
		if got != fake {
			t.Error("GetDcimProvider did not return the registered provider")
		}
	})
	t.Run("errors when unset", func(t *testing.T) {
		old := dcimProviderGetter
		t.Cleanup(func() { dcimProviderGetter = old })
		dcimProviderGetter = nil
		if _, err := GetDcimProvider(); err == nil {
			t.Error("expected error when dcimProviderGetter is unset")
		} else if !strings.Contains(err.Error(), "dcimProviderGetter not set") {
			t.Errorf("error = %q, want dcimProviderGetter not set", err)
		}
	})
}

// TestImportCSV_DcimRouting verifies ImportCSV detects a DCIM CSV header and
// stores the parsed DCIM data on the provider.
//
// Why it matters: one import entry point serves both BOM and DCIM CSV formats,
// so a header carrying a Section column must route to the DCIM parser and
// persist its grouped records.
// Inputs: the dcim.csv fixture and a fakeDcimProvider. Outputs: nil error;
// the provider receives cleared-then-set data with 6 devices.
// Data choice: the shared fixture has a known shape (6 devices), making the
// routed-and-parsed assertion unambiguous.
func TestImportCSV_DcimRouting(t *testing.T) {
	fake := &fakeDcimProvider{}
	newDcimProvider(t, fake)
	t.Cleanup(func() { commands.CsvFlag = "" })
	commands.CsvFlag = "../../../../testdata/fixtures/example/dcim.csv"
	if err := ImportCSV(nil, nil, &devicetypes.Inventory{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.data == nil {
		t.Fatal("expected DCIM records to be set on the provider")
	}
	if !fake.cleared {
		t.Error("expected ClearDcimRecords before SetDcimRecords")
	}
	if len(fake.data.Devices) != 6 {
		t.Errorf("Devices = %d, want 6", len(fake.data.Devices))
	}
}

// TestImportDcimCSV_NoRecords verifies a DCIM CSV with no data sections is a
// no-op that leaves the provider untouched.
//
// Why it matters: a defaults-only file carries no inventory, so the import must
// skip persisting rather than store an empty payload.
// Inputs: a temp CSV with a header and a single _defaults row. Outputs: nil
// error; the provider's data is never set.
// Data choice: a lone _defaults row is the smallest input that parses cleanly yet
// yields zero section records, exercising the total==0 short-circuit.
func TestImportDcimCSV_NoRecords(t *testing.T) {
	fake := &fakeDcimProvider{}
	newDcimProvider(t, fake)
	t.Cleanup(func() { commands.CsvFlag = "" })
	dir := t.TempDir()
	path := dir + "/defaults-only.csv"
	if err := os.WriteFile(path, []byte("Section,Status\n_defaults,Active\n"), 0644); err != nil {
		t.Fatal(err)
	}
	commands.CsvFlag = path
	if err := ImportCSV(nil, nil, &devicetypes.Inventory{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.data != nil {
		t.Error("expected no records to be set for a defaults-only CSV")
	}
}

// TestImportDcimCSV_LocationsOnly verifies a DCIM CSV whose only data rows
// are locations is persisted rather than skipped as empty.
//
// Why it matters: locations and interfaces are valid inventory the transform
// consumes, so a locations-only (or interfaces-only) file must not be discarded
// by the empty-import short-circuit, which previously counted neither section.
// Inputs: a temp CSV with a header and two location rows. Outputs: nil error;
// the provider receives data carrying both locations.
// Data choice: two location rows and no other section prove the total now
// includes locations, where the previous count would have been zero.
func TestImportDcimCSV_LocationsOnly(t *testing.T) {
	fake := &fakeDcimProvider{}
	newDcimProvider(t, fake)
	t.Cleanup(func() { commands.CsvFlag = "" })
	dir := t.TempDir()
	path := dir + "/locations-only.csv"
	content := "Section,Name,Role\nlocation,DC01,dc\nlocation,L3,level\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	commands.CsvFlag = path
	if err := ImportCSV(nil, nil, &devicetypes.Inventory{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.data == nil {
		t.Fatal("expected locations-only CSV to persist records")
	}
	if len(fake.data.Locations) != 2 {
		t.Errorf("Locations = %d, want 2", len(fake.data.Locations))
	}
}

// TestImportBOMCSV_StepMode verifies BOM import drives the step-through prompt
// once per record and then stores the records.
//
// Why it matters: step mode is the operator's interactive review path, so each
// record must pause for confirmation before being persisted.
// Inputs: a one-row temp CSV, StepMode enabled, and a stdin holding one newline.
// Outputs: nil error; the provider receives the single record.
// Data choice: exactly one record keeps the run to a single prompt, which is all
// a fresh-per-row reader over a pipe can satisfy without hitting EOF.
func TestImportBOMCSV_StepMode(t *testing.T) {
	fake := &fakeProvider{}
	setBOMProvider(t, fake)
	config.Cfg = &config.Config{StepMode: true, NoColor: true}
	t.Cleanup(func() {
		config.Cfg = nil
		commands.CsvFlag = ""
	})
	dir := t.TempDir()
	path := dir + "/one.csv"
	if err := os.WriteFile(path, []byte("PartNumber,Description,Quantity\nP1,Widget,1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	commands.CsvFlag = path
	withStdin(t, "\n")
	if err := ImportCSV(nil, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.records) != 1 {
		t.Errorf("records = %d, want 1", len(fake.records))
	}
}

// TestImportBOMCSV_StepModeInterrupted verifies BOM import aborts when the step
// prompt's input stream closes.
//
// Why it matters: if the operator's input ends mid-review, the import must stop
// and surface the interruption rather than silently persist unreviewed records.
// Inputs: a one-row temp CSV, StepMode enabled, and an empty stdin that returns
// EOF. Outputs: a non-nil "step interrupted" error.
// Data choice: empty stdin is the minimal way to force the first prompt read to
// fail immediately.
func TestImportBOMCSV_StepModeInterrupted(t *testing.T) {
	fake := &fakeProvider{}
	setBOMProvider(t, fake)
	config.Cfg = &config.Config{StepMode: true, NoColor: true}
	t.Cleanup(func() {
		config.Cfg = nil
		commands.CsvFlag = ""
	})
	dir := t.TempDir()
	path := dir + "/one.csv"
	if err := os.WriteFile(path, []byte("PartNumber,Description,Quantity\nP1,Widget,1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	commands.CsvFlag = path
	withStdin(t, "")
	err := ImportCSV(nil, nil, nil)
	if err == nil {
		t.Fatal("expected error when step prompt is interrupted")
	}
	if !strings.Contains(err.Error(), "step interrupted") {
		t.Errorf("error = %q, want containing 'step interrupted'", err.Error())
	}
}

// TestImportDcimCSV_StepMode verifies DCIM CSV import drives the step-through
// prompt once and then stores the parsed data.
//
// Why it matters: step mode is the operator's interactive review path for DCIM
// imports too, so each record must pause for confirmation before being persisted.
// Inputs: a one-record DCIM CSV (single role), StepMode enabled, and a stdin
// holding one newline. Outputs: nil error; the provider receives the data.
// Data choice: exactly one record keeps the run to a single prompt, which is all
// a fresh-per-row reader over a pipe can satisfy without hitting EOF.
func TestImportDcimCSV_StepMode(t *testing.T) {
	fake := &fakeDcimProvider{}
	newDcimProvider(t, fake)
	config.Cfg = &config.Config{StepMode: true, NoColor: true}
	t.Cleanup(func() {
		config.Cfg = nil
		commands.CsvFlag = ""
	})
	dir := t.TempDir()
	path := dir + "/one.csv"
	if err := os.WriteFile(path, []byte("Section,Name\nrole,ComputeNode\n"), 0644); err != nil {
		t.Fatal(err)
	}
	commands.CsvFlag = path
	withStdin(t, "\n")
	if err := ImportCSV(nil, nil, &devicetypes.Inventory{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.data == nil {
		t.Fatal("expected DCIM records to be set after stepping")
	}
}

// TestImportDcimCSV_StepModeInterrupted verifies DCIM CSV import aborts when
// the step prompt's input stream closes.
//
// Why it matters: if the operator's input ends mid-review, the import must stop
// and surface the interruption rather than silently persist unreviewed records.
// Inputs: a one-record DCIM CSV, StepMode enabled, and an empty stdin that
// returns EOF. Outputs: a non-nil "step interrupted" error.
// Data choice: empty stdin is the minimal way to force the first prompt read to
// fail immediately.
func TestImportDcimCSV_StepModeInterrupted(t *testing.T) {
	fake := &fakeDcimProvider{}
	newDcimProvider(t, fake)
	config.Cfg = &config.Config{StepMode: true, NoColor: true}
	t.Cleanup(func() {
		config.Cfg = nil
		commands.CsvFlag = ""
	})
	dir := t.TempDir()
	path := dir + "/one.csv"
	if err := os.WriteFile(path, []byte("Section,Name\nrole,ComputeNode\n"), 0644); err != nil {
		t.Fatal(err)
	}
	commands.CsvFlag = path
	withStdin(t, "")
	err := ImportCSV(nil, nil, &devicetypes.Inventory{})
	if err == nil {
		t.Fatal("expected error when step prompt is interrupted")
	}
	if !strings.Contains(err.Error(), "step interrupted") {
		t.Errorf("error = %q, want containing 'step interrupted'", err.Error())
	}
}

// TestFormatDcimRecordParsed verifies each section renders its own descriptive
// parsed string for the step-through prompt.
//
// Why it matters: step mode shows operators a per-record summary, so each section
// must produce a readable, section-appropriate description rather than a generic
// fallback.
// Inputs: one record per section (connection, device, module, interface, rack,
// location) plus a role for the default branch. Outputs: the parsed string.
// Data choice: one representative record per branch covers every switch arm.
func TestFormatDcimRecordParsed(t *testing.T) {
	tests := []struct {
		name string
		rec  DcimRecord
		want string
	}{
		{"connection", DcimRecord{Section: "connection", PartNumber: "cat6", ADevice: "a", APort: "1", BDevice: "b", BPort: "2"}, "cable cat6: a:1 ↔ b:2"},
		{"device", DcimRecord{Section: "device", Name: "n1", PartNumber: "xd670", Rack: "r1", Position: 34, Face: "front"}, "device n1 (xd670) @ r1 U34 front"},
		{"module", DcimRecord{Section: "module", PartNumber: "h100", Device: "n1", Bay: "GPU0"}, "module h100 in n1 bay GPU0"},
		{"interface", DcimRecord{Section: "interface", Name: "iLO", Device: "n1", MacAddress: "aa:bb"}, "interface iLO on n1 mac=aa:bb"},
		{"rack", DcimRecord{Section: "rack", Name: "r1", PartNumber: "shock-rack"}, "rack r1 (shock-rack)"},
		{"location", DcimRecord{Section: "location", Name: "dc1", Role: "dc", Location: "site"}, "location dc1 type=dc parent=site"},
		{"role default", DcimRecord{Section: "role", Name: "ComputeNode"}, "role ComputeNode"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDcimRecordParsed(tt.rec); got != tt.want {
				t.Errorf("formatDcimRecordParsed(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// TestFormatDcimRecordRaw verifies the compact raw label includes the section
// and the available identity fields.
//
// Why it matters: the raw view orients the operator before the parsed detail, so
// it must include the section plus name and/or part number when present.
// Inputs: a fully identified record, a name-only record, and a part-number-only
// record. Outputs: the bracketed compact label.
// Data choice: the three shapes exercise the name and part-number branches and
// their omission.
func TestFormatDcimRecordRaw(t *testing.T) {
	tests := []struct {
		name string
		rec  DcimRecord
		want string
	}{
		{"name and part", DcimRecord{Section: "device", Name: "n1", PartNumber: "xd670"}, "[device n1 xd670]"},
		{"name only", DcimRecord{Section: "role", Name: "ComputeNode"}, "[role ComputeNode]"},
		{"part only", DcimRecord{Section: "module", PartNumber: "h100"}, "[module h100]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDcimRecordRaw(tt.rec); got != tt.want {
				t.Errorf("formatDcimRecordRaw(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// TestCollectDcimStepRecords verifies all section buckets are flattened in
// pipeline order.
//
// Why it matters: step mode walks every parsed record once, so the collector
// must include each section exactly once and preserve the role→device order.
// Inputs: a DcimCSV with one record in each of two buckets (role, device).
// Outputs: a two-element slice ordered roles-before-devices.
// Data choice: two distinct sections prove both inclusion and ordering without
// ambiguity.
func TestCollectDcimStepRecords(t *testing.T) {
	data := &DcimCSV{
		Roles:   []DcimRecord{{Section: "role", Name: "ComputeNode"}},
		Devices: []DcimRecord{{Section: "device", Name: "n1"}},
	}
	got := collectDcimStepRecords(data)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Section != "role" || got[1].Section != "device" {
		t.Errorf("order = [%s, %s], want [role, device]", got[0].Section, got[1].Section)
	}
}

// TestImport_MergeError verifies Import surfaces a merge failure from a YAML file
// that parses but contains an invalid UUID key.
//
// Why it matters: YAML import maps string UUID keys onto inventory, so a
// malformed key must abort with a clear error instead of producing a corrupt
// inventory.
// Inputs: a temp YAML file with a non-UUID location key. Outputs: a non-nil
// "invalid location UUID" error.
// Data choice: syntactically valid YAML with one bad key isolates merge-time
// validation from parse errors.
func TestImport_MergeError(t *testing.T) {
	t.Cleanup(func() {
		commands.FileFlag = ""
		commands.CsvFlag = ""
	})
	dir := t.TempDir()
	path := dir + "/bad-uuid.yaml"
	if err := os.WriteFile(path, []byte("locations:\n  not-a-uuid:\n    Name: DC1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	commands.CsvFlag = ""
	commands.FileFlag = path
	err := Import(nil, nil, &devicetypes.Inventory{})
	if err == nil {
		t.Fatal("expected error for invalid location UUID")
	}
	if !strings.Contains(err.Error(), "invalid location UUID") {
		t.Errorf("error = %q, want containing 'invalid location UUID'", err.Error())
	}
}

// TestPeekCSVHeader verifies the header peek returns the first row and errors on
// a missing file.
//
// Why it matters: format auto-detection reads only the header to choose between
// the BOM and DCIM parsers, so it must return the columns and fail cleanly when
// the file is absent.
// Inputs: the dcim.csv fixture, then a nonexistent path. Outputs: the header
// slice, then a non-nil error.
// Data choice: the fixture's first column is the known "Section" sentinel; a
// missing path is the simplest open failure.
func TestPeekCSVHeader(t *testing.T) {
	t.Run("returns header fields", func(t *testing.T) {
		got, err := peekCSVHeader("../../../../testdata/fixtures/example/dcim.csv")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) == 0 || got[0] != "Section" {
			t.Errorf("header = %v, want first column 'Section'", got)
		}
	})
	t.Run("errors on missing file", func(t *testing.T) {
		if _, err := peekCSVHeader("nonexistent.csv"); err == nil {
			t.Error("expected error for missing file")
		}
	})
	t.Run("errors on empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/empty.csv"
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := peekCSVHeader(path); err == nil {
			t.Error("expected error for empty file")
		}
	})
}

// TestImportCSV_DcimParseError verifies a DCIM-routed CSV that fails to parse
// surfaces a wrapped error.
//
// Why it matters: detection routes a Section header to the DCIM parser, so a
// truncated DCIM file must fail with a clear, wrapped message instead of a
// silent no-op.
// Inputs: a temp CSV with a Section header but no data rows. Outputs: a non-nil
// "failed to parse DCIM CSV" error.
// Data choice: a header-only file is the smallest input that detects as DCIM
// yet fails the parser's row-count guard.
func TestImportCSV_DcimParseError(t *testing.T) {
	fake := &fakeDcimProvider{}
	newDcimProvider(t, fake)
	t.Cleanup(func() { commands.CsvFlag = "" })
	dir := t.TempDir()
	path := dir + "/headeronly.csv"
	if err := os.WriteFile(path, []byte("Section,Name\n"), 0644); err != nil {
		t.Fatal(err)
	}
	commands.CsvFlag = path
	err := ImportCSV(nil, nil, &devicetypes.Inventory{})
	if err == nil {
		t.Fatal("expected error for unparseable DCIM CSV")
	}
	if !strings.Contains(err.Error(), "failed to parse DCIM CSV") {
		t.Errorf("error = %q, want containing 'failed to parse DCIM CSV'", err.Error())
	}
}

// TestImportBOMCSV_NoValidRecords verifies a BOM CSV whose data rows are all
// malformed is a no-op that stores nothing.
//
// Why it matters: when every row is dropped the import has nothing to persist, so
// it must log and return nil rather than store an empty record set on the
// provider.
// Inputs: a temp CSV with a valid header and a single zero-quantity row, and a
// fakeProvider. Outputs: nil error; the provider receives no records.
// Data choice: a lone invalid row guarantees ParseCSV returns zero records with
// no error, exercising the empty-records short-circuit.
func TestImportBOMCSV_NoValidRecords(t *testing.T) {
	fake := &fakeProvider{}
	setBOMProvider(t, fake)
	config.Cfg = &config.Config{}
	t.Cleanup(func() {
		config.Cfg = nil
		commands.CsvFlag = ""
	})
	dir := t.TempDir()
	path := dir + "/allbad.csv"
	if err := os.WriteFile(path, []byte("PartNumber,Description,Quantity\nP1,Widget,0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	commands.CsvFlag = path
	if err := ImportCSV(nil, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.records != nil {
		t.Errorf("expected no records stored, got %d", len(fake.records))
	}
}
