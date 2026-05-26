package import_

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/example/commands"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// providerGetter is used to get the Example singleton from the parent package.
// Set by the parent package's init() to break the import cycle.
var providerGetter func() interface {
	ClearRecords()
	SetRecords(records []CsvRecord)
}

// SetProviderGetter allows the parent package to provide access to the singleton.
func SetProviderGetter(getter func() interface {
	ClearRecords()
	SetRecords(records []CsvRecord)
}) {
	providerGetter = getter
}

// GetProvider returns the Example singleton via the registered getter.
func GetProvider() interface {
	ClearRecords()
	SetRecords(records []CsvRecord)
} {
	if providerGetter == nil {
		panic("providerGetter not set; ensure example package init() calls SetProviderGetter")
	}
	return providerGetter()
}

// systemProviderGetter returns the Example singleton for system CSV operations.
var systemProviderGetter func() interface {
	SetSystemRecords(data *SystemCSV)
	ClearSystemRecords()
	IsSystemImport() bool
}

// SetSystemProviderGetter allows the parent package to provide system CSV access.
func SetSystemProviderGetter(getter func() interface {
	SetSystemRecords(data *SystemCSV)
	ClearSystemRecords()
	IsSystemImport() bool
}) {
	systemProviderGetter = getter
}

// GetSystemProvider returns the Example singleton for system CSV operations.
func GetSystemProvider() interface {
	SetSystemRecords(data *SystemCSV)
	ClearSystemRecords()
	IsSystemImport() bool
} {
	if systemProviderGetter == nil {
		panic("systemProviderGetter not set; ensure example package init() calls SetSystemProviderGetter")
	}
	return systemProviderGetter()
}

// peekCSVHeader reads the first line of a CSV file and returns the header fields.
func peekCSVHeader(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comment = '#'
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	return header, nil
}

// yamlInventory is an intermediate struct for YAML parsing with string keys
type yamlInventory struct {
	Locations map[string]*devicetypes.CaniLocationType `yaml:"locations"`
	Racks     map[string]*yamlRackType                 `yaml:"racks"`
	Devices   map[string]*devicetypes.CaniDeviceType   `yaml:"devices"`
	Modules   map[string]*devicetypes.CaniModuleType   `yaml:"modules"`
	Cables    map[string]*devicetypes.CaniCableType    `yaml:"cables"`
}

// yamlRackType handles parsing racks with legacy OccupiedSlots format (int → UUID)
type yamlRackType struct {
	ID               uuid.UUID         `yaml:"ID"`
	Name             string            `yaml:"Name"`
	Slug             string            `yaml:"RackTypeSlug,omitempty"`
	Location         uuid.UUID         `yaml:"Location,omitempty"`
	UHeight          int               `yaml:"UHeight"`
	Status           string            `yaml:"Status"`
	Devices          []uuid.UUID       `yaml:"Devices,omitempty"`
	OccupiedSlots    map[int]uuid.UUID `yaml:"OccupiedSlots,omitempty"` // legacy format
	ProviderMetadata map[string]any    `yaml:"ProviderMetadata,omitempty"`
}

func Import(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	// Check CSV flag first (takes precedence)
	if commands.CsvFlag != "" {
		return ImportCSV(cmd, args, inventory)
	}

	filePath := commands.FileFlag
	if filePath == "" {
		log.Println("No file specified, skipping import")
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var parsed yamlInventory
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert and merge into inventory
	if err := mergeYAMLInventory(inventory, &parsed); err != nil {
		return err
	}

	log.Printf("Imported inventory from %s: %d locations, %d racks, %d devices, %d cables",
		filePath,
		len(inventory.Locations),
		len(inventory.Racks),
		len(inventory.Devices),
		len(inventory.Cables))

	return nil
}

func mergeYAMLInventory(dst *devicetypes.Inventory, src *yamlInventory) error {
	// Locations
	if src.Locations != nil {
		if dst.Locations == nil {
			dst.Locations = make(map[uuid.UUID]*devicetypes.CaniLocationType)
		}
		for k, v := range src.Locations {
			id, err := uuid.Parse(k)
			if err != nil {
				return fmt.Errorf("invalid location UUID %s: %w", k, err)
			}
			dst.Locations[id] = v
		}
	}

	// Racks - convert legacy slot format to face-aware format
	if src.Racks != nil {
		if dst.Racks == nil {
			dst.Racks = make(map[uuid.UUID]*devicetypes.CaniRackType)
		}
		for k, v := range src.Racks {
			id, err := uuid.Parse(k)
			if err != nil {
				return fmt.Errorf("invalid rack UUID %s: %w", k, err)
			}
			// Convert yamlRackType to CaniRackType with migration
			rack := &devicetypes.CaniRackType{
				ID:         v.ID,
				Name:       v.Name,
				Slug:       v.Slug,
				Location:   v.Location,
				UHeight:    v.UHeight,
				ObjectMeta: devicetypes.ObjectMeta{Status: v.Status, ProviderMetadata: v.ProviderMetadata},
				Devices:    v.Devices,
			}
			// Migrate legacy OccupiedSlots to face-aware format
			if v.OccupiedSlots != nil {
				rack.MigrateLegacySlots(v.OccupiedSlots)
			}
			dst.Racks[id] = rack
		}
	}

	// Devices
	if src.Devices != nil {
		if dst.Devices == nil {
			dst.Devices = make(map[uuid.UUID]*devicetypes.CaniDeviceType)
		}
		for k, v := range src.Devices {
			id, err := uuid.Parse(k)
			if err != nil {
				return fmt.Errorf("invalid device UUID %s: %w", k, err)
			}
			v.ID = id // ensure ID is set
			dst.Devices[id] = v
		}
	}

	// Cables
	if src.Cables != nil {
		if dst.Cables == nil {
			dst.Cables = make(map[uuid.UUID]*devicetypes.CaniCableType)
		}
		for k, v := range src.Cables {
			id, err := uuid.Parse(k)
			if err != nil {
				return fmt.Errorf("invalid cable UUID %s: %w", k, err)
			}
			dst.Cables[id] = v
		}
	}

	return nil
}

// ImportCSV parses a CSV file and stores raw records on the Example provider.
// Auto-detects system CSV format (multi-section with Section column) vs
// traditional BOM CSV format. Does NOT create inventory objects; that is the
// responsibility of the transform phase.
func ImportCSV(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	// Peek at header to detect format
	header, err := peekCSVHeader(commands.CsvFlag)
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	if IsSystemCSV(header) {
		return importSystemCSV(cmd, args, inventory)
	}
	return importBOMCSV(cmd, args, inventory)
}

// importSystemCSV parses a system CSV and stores the grouped data on the provider.
func importSystemCSV(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	data, err := ParseSystemCSV(commands.CsvFlag)
	if err != nil {
		return fmt.Errorf("failed to parse system CSV: %w", err)
	}

	total := len(data.Roles) + len(data.Racks) + len(data.Devices) + len(data.Modules) + len(data.Connections)
	if total == 0 {
		log.Println("No valid records found in system CSV")
		return nil
	}

	prov := GetSystemProvider()
	prov.ClearSystemRecords()
	prov.SetSystemRecords(data)

	log.Printf("Parsed system CSV from %s: %d roles, %d racks, %d devices, %d modules, %d connections",
		commands.CsvFlag,
		len(data.Roles), len(data.Racks), len(data.Devices), len(data.Modules), len(data.Connections))

	return nil
}

// importBOMCSV parses a traditional BOM CSV and stores raw records on the provider.
func importBOMCSV(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	records, err := ParseCSV(commands.CsvFlag)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		log.Println("No valid records found in CSV")
		return nil
	}

	// Show step-through output if step mode is enabled
	if config.Cfg.StepMode {
		opts := visual.ETLOptions{NoColor: config.Cfg.NoColor}
		for i, rec := range records {
			rawData := formatRawCSVRecord(rec)
			parsed := formatParsedRecord(rec)
			if err := visual.PromptCSVRowStepRaw(i+1, len(records), rawData, parsed, opts); err != nil {
				return fmt.Errorf("step interrupted: %w", err)
			}
		}
	}

	// Get the singleton Example provider to store raw records
	prov := GetProvider()
	prov.ClearRecords()
	prov.SetRecords(records)

	log.Printf("Parsed CSV from %s: %d records",
		commands.CsvFlag,
		len(records))

	return nil
}

// formatRawCSVRecord creates a compact string representation of the raw CSV data.
func formatRawCSVRecord(rec CsvRecord) string {
	parts := []string{rec.PartNumber, rec.Description}
	if rec.Quantity > 1 {
		parts = append(parts, fmt.Sprintf("qty=%d", rec.Quantity))
	}
	if rec.ConfigGroup != "" {
		parts = append(parts, fmt.Sprintf("grp=%s", rec.ConfigGroup))
	}
	if rec.SourceDevice != "" {
		parts = append(parts, fmt.Sprintf("%s:%s→%s:%s", rec.SourceDevice, rec.SourcePort, rec.DestDevice, rec.DestPort))
	}
	return fmt.Sprintf("%s", parts)
}

// formatParsedRecord creates a descriptive string of the parsed record type.
func formatParsedRecord(rec CsvRecord) string {
	if IsCableRecord(rec) {
		return fmt.Sprintf("cable: %s:%s ↔ %s:%s", rec.SourceDevice, rec.SourcePort, rec.DestDevice, rec.DestPort)
	}
	hwType := "device"
	if rec.Quantity > 1 {
		return fmt.Sprintf("%s × %d: %s", hwType, rec.Quantity, rec.Description)
	}
	return fmt.Sprintf("%s: %s", hwType, rec.Description)
}
