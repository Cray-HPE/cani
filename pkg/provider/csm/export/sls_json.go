package export

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

// validateSLSHardware checks that staged CANI devices with xnames are
// present in the expected hardware map and that every expected entry
// has a valid Class.  Devices created as implicit parents during the
// SLS import (status "active") are skipped when they are absent from
// expected, because buildExpectedHardware intentionally excludes them.
func validateSLSHardware(
	expected map[string]import_.SlsHardware,
	inventory devicetypes.Inventory,
) error {
	for _, dev := range inventory.Devices {
		xname := extractXname(dev)
		if xname == "" {
			continue
		}
		hw, ok := expected[xname]
		if !ok {
			// Only flag an error for staged devices that should
			// have been included in the expected map.  Implicit
			// parents from the import are not expected to exist
			// in SLS.
			if strings.EqualFold(dev.Status, "staged") {
				return fmt.Errorf("validation failed: device %s (xname %s) not found in SLS", dev.ID, xname)
			}
			continue
		}
		if hw.Class == "" {
			return fmt.Errorf("validation failed: %s has no Class", xname)
		}
	}
	return nil
}

// writeSLSJSON writes the expected hardware map to w as a JSON object
// keyed by xname, sorted alphabetically.
func writeSLSJSON(w io.Writer, expected map[string]import_.SlsHardware) error {
	// Sort xnames for deterministic output.
	keys := make([]string, 0, len(expected))
	for k := range expected {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ordered := make(map[string]import_.SlsHardware, len(expected))
	for _, k := range keys {
		ordered[k] = expected[k]
	}

	data, err := json.MarshalIndent(ordered, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling SLS JSON: %w", err)
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	return err
}
