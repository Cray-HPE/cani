/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package datastores

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// legacyInventory matches the v1alpha1 JSON shape written by cani-legacy.
type legacyInventory struct {
	SchemaVersion string                       `json:"SchemaVersion"`
	Provider      string                       `json:"Provider"`
	Hardware      map[uuid.UUID]legacyHardware `json:"Hardware"`
}

// legacyHardware matches a single Hardware entry from v1alpha1.
type legacyHardware struct {
	ID              uuid.UUID                  `json:"ID"`
	Name            string                     `json:"Name,omitempty"`
	Type            string                     `json:"Type,omitempty"`
	DeviceTypeSlug  string                     `json:"DeviceTypeSlug,omitempty"`
	Vendor          string                     `json:"Vendor,omitempty"`
	Architecture    string                     `json:"Architecture,omitempty"`
	Model           string                     `json:"Model,omitempty"`
	Status          string                     `json:"Status,omitempty"`
	Properties      map[string]interface{}     `json:"Properties,omitempty"`
	ProviderMeta    map[string]json.RawMessage `json:"ProviderMetadata,omitempty"`
	Parent          uuid.UUID                  `json:"Parent,omitempty"`
	Children        []uuid.UUID                `json:"Children,omitempty"`
	LocationPath    []legacyLocationToken      `json:"LocationPath,omitempty"`
	LocationOrdinal *int                       `json:"LocationOrdinal,omitempty"`
}

// legacyLocationToken matches a single token from a v1alpha1 LocationPath.
type legacyLocationToken struct {
	HardwareType string `json:"HardwareType"`
	Ordinal      int    `json:"Ordinal"`
}

// legacyCabinetMeta matches the CSM CabinetMetadata nested under ProviderMetadata["csm"]["Cabinet"].
type legacyCabinetMeta struct {
	HMNVlan *int `json:"HMNVlan,omitempty"`
}

// legacyNodeMeta matches the CSM NodeMetadata nested under ProviderMetadata["csm"]["Node"].
type legacyNodeMeta struct {
	Role    *string  `json:"Role,omitempty"`
	SubRole *string  `json:"SubRole,omitempty"`
	Nid     *int     `json:"Nid,omitempty"`
	Alias   []string `json:"Alias,omitempty"`
}

// legacyCsmMeta is the top-level CSM provider metadata envelope.
type legacyCsmMeta struct {
	Cabinet *legacyCabinetMeta `json:"Cabinet,omitempty"`
	Node    *legacyNodeMeta    `json:"Node,omitempty"`
}

// isLegacyDatastore returns true when raw JSON looks like a v1alpha1 datastore
// (has a top-level "Hardware" key and SchemaVersion is "v1alpha1" or empty).
func isLegacyDatastore(raw []byte) bool {
	var probe struct {
		SchemaVersion string          `json:"SchemaVersion"`
		Hardware      json.RawMessage `json:"Hardware"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	if len(probe.Hardware) == 0 {
		return false
	}
	return probe.SchemaVersion == devicetypes.SchemaVersionV1Alpha1 ||
		probe.SchemaVersion == ""
}

// backupDatastore copies path to path.canisave before migration.
func backupDatastore(path string) error {
	src, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening datastore for backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(path + ".canisave")
	if err != nil {
		return fmt.Errorf("creating backup file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copying datastore to backup: %w", err)
	}
	return nil
}

// migrateV1Alpha1 converts a v1alpha1 JSON datastore into a v1alpha2 Inventory.
func migrateV1Alpha1(raw []byte) (*devicetypes.Inventory, error) {
	var legacy legacyInventory
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return nil, fmt.Errorf("unmarshaling legacy datastore: %w", err)
	}

	inv := devicetypes.NewInventory()
	inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha2
	inv.Provider = legacy.Provider

	for _, hw := range legacy.Hardware {
		if hw.Type == "System" {
			migrateSystem(inv, hw)
			continue
		}
		if hw.Type == "Cabinet" {
			migrateCabinet(inv, hw)
			continue
		}
		migrateDevice(inv, hw)
	}

	inv.RebuildProviderKeyIndex()
	return inv, nil
}

// migrateSystem converts the legacy System root to a CaniLocationType.
func migrateSystem(inv *devicetypes.Inventory, hw legacyHardware) {
	loc := &devicetypes.CaniLocationType{
		ID:           hw.ID,
		Name:         "System",
		LocationType: "system",
		ObjectMeta:   devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive)},
	}
	inv.Locations[hw.ID] = loc
}

// migrateCabinet creates both a CaniDeviceType and a CaniRackType for a cabinet,
// mirroring the CSM transform pattern.
func migrateCabinet(inv *devicetypes.Inventory, hw legacyHardware) {
	md := buildCsmMetadata(hw)

	dev := baseDevice(hw)
	dev.Type = devicetypes.TypeCabinet
	dev.ProviderMetadata = map[string]any{"csm": md}
	inv.Devices[hw.ID] = dev

	rackID := uuid.New()
	rack := &devicetypes.CaniRackType{
		ID:         rackID,
		Name:       hw.Name,
		ObjectMeta: devicetypes.ObjectMeta{Status: statusOrDefault(hw.Status), ProviderMetadata: map[string]any{"csm": md}},
		UHeight:    cabinetUHeight(md),
	}
	if hw.Vendor != "" {
		rack.Manufacturer = hw.Vendor
	}
	if hw.Model != "" {
		rack.Model = hw.Model
	}
	inv.Racks[rackID] = rack

	// Point cabinet device at its rack.
	dev.Parent = rackID
}

// migrateDevice handles all non-System, non-Cabinet hardware types.
func migrateDevice(inv *devicetypes.Inventory, hw legacyHardware) {
	dev := baseDevice(hw)
	dev.Type = mapLegacyType(hw.Type)
	dev.ProviderMetadata = map[string]any{"csm": buildCsmMetadata(hw)}
	inv.Devices[hw.ID] = dev
}

// baseDevice creates a CaniDeviceType with identity fields copied from legacy hardware.
func baseDevice(hw legacyHardware) *devicetypes.CaniDeviceType {
	dev := &devicetypes.CaniDeviceType{
		ID:         hw.ID,
		Name:       hw.Name,
		Slug:       hw.DeviceTypeSlug,
		Vendor:     hw.Vendor,
		Model:      hw.Model,
		ObjectMeta: devicetypes.ObjectMeta{Status: statusOrDefault(hw.Status)},
		Parent:     hw.Parent,
	}
	return dev
}

// buildCsmMetadata flattens legacy CSM provider metadata and location info
// into the map[string]any shape expected by the new ProviderMetadata["csm"].
func buildCsmMetadata(hw legacyHardware) map[string]any {
	md := map[string]any{}

	// Flatten location info.
	if hw.LocationOrdinal != nil {
		md["locationOrdinal"] = *hw.LocationOrdinal
	}
	if len(hw.LocationPath) > 0 {
		md["locationPath"] = locationPathString(hw.LocationPath)
	}

	// Decode CSM-specific nested metadata.
	raw, ok := hw.ProviderMeta["csm"]
	if !ok {
		return md
	}
	var csm legacyCsmMeta
	if err := json.Unmarshal(raw, &csm); err != nil {
		return md
	}

	if csm.Cabinet != nil && csm.Cabinet.HMNVlan != nil {
		md["hmnVlan"] = *csm.Cabinet.HMNVlan
	}
	if csm.Node != nil {
		if csm.Node.Role != nil {
			md["role"] = *csm.Node.Role
		}
		if csm.Node.SubRole != nil {
			md["subRole"] = *csm.Node.SubRole
		}
		if csm.Node.Nid != nil {
			md["nid"] = *csm.Node.Nid
		}
		if len(csm.Node.Alias) > 0 {
			md["aliases"] = csm.Node.Alias
		}
	}

	return md
}

// locationPathString converts a legacy LocationPath into a human-readable string
// (e.g. "System:0->Cabinet:3000->Chassis:0").
func locationPathString(tokens []legacyLocationToken) string {
	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = fmt.Sprintf("%s:%d", t.HardwareType, t.Ordinal)
	}
	return strings.Join(parts, "->")
}

// cabinetUHeight returns an appropriate rack height for a migrated cabinet.
func cabinetUHeight(md map[string]any) int {
	// Default to standard 42U; CSM transform uses class for this,
	// but the legacy datastore doesn't store class directly.
	return 42
}

// statusOrDefault returns the status string, defaulting to "Staged".
// If the input is a known Nautobot status, it is normalized to Title case.
func statusOrDefault(s string) string {
	if s == "" {
		return string(devicetypes.StatusStaged)
	}
	return devicetypes.NormalizeStatus(s)
}

// mapLegacyType maps a legacy HardwareType string to the new devicetypes.Type.
func mapLegacyType(legacyType string) devicetypes.Type {
	switch legacyType {
	case "Cabinet":
		return devicetypes.TypeCabinet
	case "Chassis":
		return devicetypes.TypeChassis
	case "NodeBlade":
		return devicetypes.TypeBlade
	case "NodeCard":
		return devicetypes.TypeNodeCard
	case "Node":
		return devicetypes.TypeNode
	case "ManagementSwitch":
		return devicetypes.TypeMgmtSwitch
	case "ManagementSwitchEnclosure":
		return devicetypes.TypeMgmtSwitch
	case "ManagementSwitchController":
		return devicetypes.TypeMgmtSwitch
	case "HighSpeedSwitch":
		return devicetypes.TypeHsnSwitch
	case "HighSpeedSwitchEnclosure":
		return devicetypes.TypeHsnSwitch
	case "HighSpeedSwitchController":
		return devicetypes.TypeHsnSwitch
	case "CabinetPDU":
		return devicetypes.TypeCabinetPDU
	case "CabinetPDUController":
		return devicetypes.TypeCabinetPDU
	case "CoolingDistributionUnit":
		return devicetypes.TypeCDU
	case "ChassisManagementModule":
		return devicetypes.TypeModule
	case "CabinetEnvironmentalController":
		return devicetypes.TypeModule
	case "NodeController":
		return devicetypes.TypeModule
	default:
		return devicetypes.Type(strings.ToLower(legacyType))
	}
}

// v1alpha2Probe extracts the pre-v1alpha3 device link fields from raw datastore
// JSON. Earlier schemas persisted the derived rack / parentDevice FKs; v1alpha3
// makes them non-serialized, so the migration must read them from the raw bytes
// to recover the authoritative Parent FK.
type v1alpha2Probe struct {
	Devices map[uuid.UUID]struct {
		Parent       uuid.UUID `json:"parent"`
		Rack         uuid.UUID `json:"rack"`
		ParentDevice uuid.UUID `json:"parentDevice"`
	} `json:"devices"`
}

// migrateV1Alpha2 upgrades a v1alpha2 inventory to v1alpha3 in place. It
// back-fills each device's authoritative Parent FK from the legacy derived
// fields (parent -> parentDevice -> rack, in priority order) when Parent is
// unset, then stamps the schema version. The immediate container wins, so a
// blade with both a parent chassis and a rack resolves to the chassis and lets
// RebuildDerivedState cascade the rack. Reverse indices are rebuilt separately.
func migrateV1Alpha2(raw []byte, inv *devicetypes.Inventory) {
	var probe v1alpha2Probe
	_ = json.Unmarshal(raw, &probe) // best-effort: absent fields decode as uuid.Nil

	for id, dev := range inv.Devices {
		if dev == nil || dev.Parent != uuid.Nil {
			continue
		}
		link := probe.Devices[id]
		switch {
		case link.Parent != uuid.Nil:
			dev.Parent = link.Parent
		case link.ParentDevice != uuid.Nil:
			dev.Parent = link.ParentDevice
		case link.Rack != uuid.Nil:
			dev.Parent = link.Rack
		}
	}

	inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha3
}

// migrateInventoryMetadata detects the old inventory-level "providerMetadata"
// key (used by intermediate builds before the ObjectMeta refactor) and
// converts it to the typed Metadata field. Returns true if migration occurred.
func migrateInventoryMetadata(raw []byte, inv *devicetypes.Inventory) bool {
	var probe struct {
		ProviderMeta map[string]json.RawMessage `json:"providerMetadata"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil || len(probe.ProviderMeta) == 0 {
		return false
	}

	// Already has typed metadata — nothing to do.
	if inv.Metadata != nil && (len(inv.Metadata.Roles) > 0 || len(inv.Metadata.Statuses) > 0 || len(inv.Metadata.Tags) > 0) {
		return false
	}

	if inv.Metadata == nil {
		inv.Metadata = &devicetypes.InventoryMetadata{}
	}

	// The old format stored entries under a provider key (e.g. "nautobot")
	// with sub-keys "roles", "statuses", "tags". Merge all providers.
	for _, providerRaw := range probe.ProviderMeta {
		var bucket struct {
			Roles    []devicetypes.MetadataEntry `json:"roles"`
			Statuses []devicetypes.MetadataEntry `json:"statuses"`
			Tags     []devicetypes.MetadataEntry `json:"tags"`
		}
		if err := json.Unmarshal(providerRaw, &bucket); err != nil {
			continue
		}
		inv.Metadata.Roles = append(inv.Metadata.Roles, bucket.Roles...)
		inv.Metadata.Statuses = append(inv.Metadata.Statuses, bucket.Statuses...)
		inv.Metadata.Tags = append(inv.Metadata.Tags, bucket.Tags...)
	}

	return len(inv.Metadata.Roles) > 0 || len(inv.Metadata.Statuses) > 0 || len(inv.Metadata.Tags) > 0
}
