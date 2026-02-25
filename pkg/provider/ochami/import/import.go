package import_

import (
	"log"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ochami/commands"
	"github.com/spf13/cobra"
)

// providerGetter is used to get the Example singleton from the parent package.
// Set by the parent package's init() to break the import cycle.
var providerGetter func() interface {
	ClearRecords()
	SetRecords(records []JSONDeviceRecord)
}

// SetProviderGetter allows the parent package to provide access to the singleton.
func SetProviderGetter(getter func() interface {
	ClearRecords()
	SetRecords(records []JSONDeviceRecord)
}) {
	providerGetter = getter
}

// GetProvider returns the Example singleton via the registered getter.
func GetProvider() interface {
	ClearRecords()
	SetRecords(records []JSONDeviceRecord)
} {
	if providerGetter == nil {
		panic("providerGetter not set; ensure example package init() calls SetProviderGetter")
	}
	return providerGetter()
}

func Import(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	// Common patterns:
	//   - Parse files or query APIs to get data
	//   - Store data in provider struct for later processing in Transform()
	//   - Report what was imported, skipped, or errored

	if commands.JsonFileFlag != "" {
		return ImportOchamiDevices(cmd, args, inventory)

	}

	return nil
}

func ImportOchamiDevices(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	records, err := ParseJson(commands.JsonFileFlag)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		log.Println("No valid records found in JSON")
		return nil
	}

	// TODO: add step through output option here

	// Get the singleton Example provider to store raw records
	prov := GetProvider()
	prov.ClearRecords()
	prov.SetRecords(records)

	return nil
}
