package export

import (
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

// hardwareChanges categorises the differences between two SLS states.
type hardwareChanges struct {
	Added   []import_.SlsHardware
	Removed []import_.SlsHardware
	Changed []import_.SlsHardware // expected version of entries that differ
}

// diffHardware compares expected against current SLS hardware and
// produces the set of additions, removals, and updates.
// Only hardware entries with CANI metadata are considered for changes;
// entries without CANI metadata in the expected set are left untouched.
func diffHardware(
	expected map[string]import_.SlsHardware,
	current map[string]import_.SlsHardware,
) hardwareChanges {
	var changes hardwareChanges

	for xname, exp := range expected {
		// Only reconcile entries that have CANI metadata.
		if !hasCaniMetadata(exp) {
			continue
		}

		cur, exists := current[xname]
		if !exists {
			changes.Added = append(changes.Added, exp)
			continue
		}

		if hardwareNeedsUpdate(cur, exp) {
			changes.Changed = append(changes.Changed, exp)
		}
	}

	// Removals: entries in current that have CANI metadata but are
	// missing from expected. For now we skip removals — CANI only
	// adds metadata, never removes hardware during reconcile.
	// This keeps the initial implementation safe and minimal.

	return changes
}

// hasCaniMetadata returns true if the hardware entry has the CANI
// schema version key set in ExtraProperties.
func hasCaniMetadata(hw import_.SlsHardware) bool {
	if hw.ExtraProperties == nil {
		return false
	}
	_, ok := hw.ExtraProperties["@cani.slsSchemaVersion"]
	return ok
}

// hardwareNeedsUpdate returns true when the expected entry differs
// from the current entry in a way that requires a PUT. We check for
// CANI metadata presence — if the current entry lacks CANI metadata,
// it needs updating.
func hardwareNeedsUpdate(
	current import_.SlsHardware,
	expected import_.SlsHardware,
) bool {
	if current.ExtraProperties == nil {
		return true
	}
	// If the current entry does not have CANI metadata, update it.
	if _, ok := current.ExtraProperties["@cani.slsSchemaVersion"]; !ok {
		return true
	}
	// If the CANI status changed, update it.
	curStatus, _ := current.ExtraProperties["@cani.status"].(string)
	expStatus, _ := expected.ExtraProperties["@cani.status"].(string)
	if curStatus != expStatus {
		return true
	}
	return false
}
