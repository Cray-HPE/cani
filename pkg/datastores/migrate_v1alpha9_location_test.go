/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
	"maps"
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// v1Alpha9Site is the CRUD fixture as a v1alpha9 writer left it, laid out to
// tell a location that holds the rack's from one that does not: the rack
// stands in row-1 of site-a, and site-b is a second, unrelated site.
type v1Alpha9Site struct {
	inventory          *devicetypes.Inventory
	ids                crudFixtureIDs
	rowID, otherSiteID uuid.UUID
}

func newV1Alpha9Site(t *testing.T) v1Alpha9Site {
	t.Helper()
	inventory, ids := newCRUDInventory(t)
	inventory.SchemaVersion = devicetypes.SchemaVersionV1Alpha9
	row := &devicetypes.CaniLocationType{ID: uuid.New(), Name: "row-1", LocationType: "row", Parent: ids.locationID}
	other := &devicetypes.CaniLocationType{ID: uuid.New(), Name: "site-b", LocationType: "site"}
	inventory.Locations[row.ID], inventory.Locations[other.ID] = row, other
	inventory.Racks[ids.rackID].Location = row.ID
	// A v1alpha9 writer stored the rack's location on every mounted device.
	for _, device := range inventory.Devices {
		device.Location = row.ID
	}
	if err := inventory.RebuildDerivedState().Err(); err != nil {
		t.Fatal(err)
	}
	return v1Alpha9Site{inventory: inventory, ids: ids, rowID: row.ID, otherSiteID: other.ID}
}

// store puts a device into the fixture the way a datastore holds it, without
// the rebuild that follows a migration.
func (s v1Alpha9Site) store(device *devicetypes.CaniDeviceType) *devicetypes.CaniDeviceType {
	device.ID = uuid.New()
	s.inventory.Devices[device.ID] = device
	return device
}

// TestMigrateV1Alpha9PlacesDeviceByTypedLocation verifies a location typed
// with --metadata becomes the device's own location wherever the model can
// hold it.
//
// Why it matters: the exporter read that name ahead of the inventory's
// placement, and stops; a device whose placement was only typed would be
// exported to the rack's or the default location instead.
// Inputs: a device with no parent and no location, one with no parent placed
// elsewhere, and a racked device, each with a typed location name.
// Outputs: the location the device records, which the rebuild that follows
// a migration keeps, and a report only where a recorded location was
// replaced.
// Data choice: site-a holds the rack's row and site-b does not, so the
// racked case exercises the one typed location a racked device may keep.
func TestMigrateV1Alpha9PlacesDeviceByTypedLocation(t *testing.T) {
	tests := []struct {
		name       string
		device     func(site v1Alpha9Site) *devicetypes.CaniDeviceType
		typed      string
		want       func(site v1Alpha9Site) uuid.UUID
		wantReport []string
	}{
		{
			name: "device with no parent and no location",
			device: func(site v1Alpha9Site) *devicetypes.CaniDeviceType {
				return site.store(&devicetypes.CaniDeviceType{Name: "cdu-1", Type: devicetypes.TypeCDU})
			},
			typed: "site-b",
			want:  func(site v1Alpha9Site) uuid.UUID { return site.otherSiteID },
		},
		{
			name: "device with no parent placed elsewhere",
			device: func(site v1Alpha9Site) *devicetypes.CaniDeviceType {
				return site.store(&devicetypes.CaniDeviceType{Name: "cdu-1", Type: devicetypes.TypeCDU, Location: site.ids.locationID})
			},
			typed:      "site-b",
			want:       func(site v1Alpha9Site) uuid.UUID { return site.otherSiteID },
			wantReport: []string{`provider metadata "location" replaced the location the device recorded`},
		},
		{
			name:   "racked device and a location that holds its rack's",
			device: func(site v1Alpha9Site) *devicetypes.CaniDeviceType { return site.inventory.Devices[site.ids.nodeID] },
			typed:  "site-a",
			want:   func(site v1Alpha9Site) uuid.UUID { return site.ids.locationID },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			site := newV1Alpha9Site(t)
			device := tc.device(site)
			device.ProviderMetadata = typed(map[string]any{"location": tc.typed})

			report := foldProviderMetadata(site.inventory)
			rebuilt := site.inventory.RebuildDerivedState().Err()

			if rebuilt != nil {
				t.Fatalf("rebuild after the migration: %v", rebuilt)
			}
			if want := tc.want(site); device.Location != want {
				t.Errorf("location = %s, want %s", device.Location, want)
			}
			if device.ProviderMetadata != nil || device.CustomFields != nil {
				t.Errorf("provider metadata = %v, custom fields = %v, want the name to be held by neither",
					device.ProviderMetadata, device.CustomFields)
			}
			assertReported(t, report, tc.wantReport...)
		})
	}
}

// TestMigrateV1Alpha9RefusesLocationTheModelCannotHold verifies a typed
// location is not applied where a device's location is not its own to
// choose, or where the exporter never read it.
//
// Why it matters: applying it would store a placement the rebuild undoes or
// Nautobot rejects, and a value the exporter passed over must not go live.
// Inputs: a racked device naming a site that does not hold its rack, a
// device installed in another device, a device whose HPCM bucket holds a
// location of its own, and a typed value that is not text. Outputs: the
// location the device already had, no trace of the typed value, one line
// saying why.
// Data choice: the HPCM value is the rack-and-chassis map that provider
// stores under the same key, which is what shadowed a typed name.
func TestMigrateV1Alpha9RefusesLocationTheModelCannotHold(t *testing.T) {
	tests := []struct {
		name         string
		device       func(site v1Alpha9Site) *devicetypes.CaniDeviceType
		provider     map[string]any
		want         func(site v1Alpha9Site) uuid.UUID
		wantProvider map[string]any
		wantReport   string
	}{
		{
			name: "racked device and a location that does not hold its rack's",
			device: func(site v1Alpha9Site) *devicetypes.CaniDeviceType {
				return site.inventory.Devices[site.ids.switchID]
			},
			provider:   typed(map[string]any{"location": "site-b"}),
			want:       func(site v1Alpha9Site) uuid.UUID { return site.rowID },
			wantReport: `provider metadata "location" was not applied: it does not hold the location of the device's rack`,
		},
		{
			name: "device installed in another device",
			device: func(site v1Alpha9Site) *devicetypes.CaniDeviceType {
				return site.store(&devicetypes.CaniDeviceType{Name: "blade-1", Type: devicetypes.TypeNode,
					Parent: site.ids.nodeID, Location: site.rowID})
			},
			provider:   typed(map[string]any{"location": "site-b"}),
			want:       func(v1Alpha9Site) uuid.UUID { return uuid.Nil },
			wantReport: `provider metadata "location" was not applied: a device installed in another device takes that device's location`,
		},
		{
			name: "another provider's value was read first",
			device: func(site v1Alpha9Site) *devicetypes.CaniDeviceType {
				return site.store(&devicetypes.CaniDeviceType{Name: "cdu-1", Type: devicetypes.TypeCDU})
			},
			provider: map[string]any{
				"hpcm":     map[string]any{"location": map[string]any{"rack": float64(1)}},
				"nautobot": map[string]any{"location": "site-b"},
			},
			want:         func(v1Alpha9Site) uuid.UUID { return uuid.Nil },
			wantProvider: map[string]any{"hpcm": map[string]any{"location": map[string]any{"rack": float64(1)}}},
			wantReport:   `provider metadata "location" was dropped: another copy of that key took precedence`,
		},
		{
			name: "typed value is not text",
			device: func(site v1Alpha9Site) *devicetypes.CaniDeviceType {
				return site.store(&devicetypes.CaniDeviceType{Name: "cdu-1", Type: devicetypes.TypeCDU})
			},
			provider:   typed(map[string]any{"location": float64(12)}),
			want:       func(v1Alpha9Site) uuid.UUID { return uuid.Nil },
			wantReport: `provider metadata "location" was dropped: it is not a location name`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			site := newV1Alpha9Site(t)
			device := tc.device(site)
			device.ProviderMetadata = tc.provider

			report := foldProviderMetadata(site.inventory)

			if want := tc.want(site); device.Location != want {
				t.Errorf("location = %s, want %s", device.Location, want)
			}
			if !reflect.DeepEqual(device.ProviderMetadata, tc.wantProvider) {
				t.Errorf("provider metadata = %v, want %v", device.ProviderMetadata, tc.wantProvider)
			}
			if device.CustomFields != nil {
				t.Errorf("custom fields = %v, want none: a device's location was never a custom field", device.CustomFields)
			}
			assertReported(t, report, tc.wantReport)
		})
	}
}

// TestMigrateV1Alpha9KeepsLocationNameItCannotPlace verifies a typed
// location that is not the name of exactly one location in the inventory is
// kept where the operator can find it.
//
// Why it matters: that name did place the device on export, by looking the
// location up in Nautobot; it cannot become a placement here, and dropping
// it would lose the only record of where the device was sent.
// Inputs: a racked device naming a location the inventory lacks, and one
// naming a location two sites both have. Outputs: the device where it was,
// the name under the nautobot bucket beside whatever other bucket the device
// has, no custom field for it, one line saying why.
// Data choice: the second row-1 is under site-b, the way Nautobot allows one
// name under two parents; the node also carries a CSM bucket and the switch
// carries nothing else, so the name is kept whether or not the device's
// provider metadata outlives the migration.
func TestMigrateV1Alpha9KeepsLocationNameItCannotPlace(t *testing.T) {
	tests := []struct {
		name       string
		typed      string
		device     func(site v1Alpha9Site) *devicetypes.CaniDeviceType
		arrange    func(site v1Alpha9Site)
		wantReport string
	}{
		{
			name:       "name of no location",
			typed:      "dc-9",
			device:     func(site v1Alpha9Site) *devicetypes.CaniDeviceType { return site.inventory.Devices[site.ids.nodeID] },
			arrange:    func(v1Alpha9Site) {},
			wantReport: `provider metadata "location" names no location in the inventory; export no longer places the device by it`,
		},
		{
			name:   "name of two locations",
			typed:  "row-1",
			device: func(site v1Alpha9Site) *devicetypes.CaniDeviceType { return site.inventory.Devices[site.ids.switchID] },
			arrange: func(site v1Alpha9Site) {
				twin := &devicetypes.CaniLocationType{ID: uuid.New(), Name: "row-1", LocationType: "row", Parent: site.otherSiteID}
				site.inventory.Locations[twin.ID] = twin
			},
			wantReport: `provider metadata "location" names more than one location in the inventory; export no longer places the device by it`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			site := newV1Alpha9Site(t)
			tc.arrange(site)
			device := tc.device(site)
			wantProvider := maps.Clone(device.ProviderMetadata)
			if wantProvider == nil {
				wantProvider, device.ProviderMetadata = map[string]any{}, map[string]any{}
			}
			wantProvider["nautobot"] = map[string]any{"location": tc.typed}
			device.ProviderMetadata["nautobot"] = map[string]any{"location": tc.typed, "asset": "A-17"}

			report := foldProviderMetadata(site.inventory)

			if device.Location != site.rowID {
				t.Errorf("location = %s, want the rack's %s", device.Location, site.rowID)
			}
			if !reflect.DeepEqual(device.ProviderMetadata, wantProvider) {
				t.Errorf("provider metadata = %v, want %v", device.ProviderMetadata, wantProvider)
			}
			if wantCustom := map[string]any{"asset": "A-17"}; !reflect.DeepEqual(device.CustomFields, wantCustom) {
				t.Errorf("custom fields = %v, want %v", device.CustomFields, wantCustom)
			}
			assertReported(t, report, tc.wantReport)
		})
	}
}

// TestMigrateV1Alpha9DropsDerivedLocation verifies a location stored on a
// parented device survives only when it is its rack's.
//
// Why it matters: earlier generations re-derived that location on every
// load, so a stale one was harmless; from v1alpha10 it would be read as a
// location the device records for itself.
// Inputs: a racked device storing its rack's location, a racked device
// storing an unrelated site, and a device with no parent placed in that
// site. Outputs: the first and third keep their location, the second has
// none until the rebuild gives it the rack's; nothing is reported.
// Data choice: the stale value is a real location, so only the rule, and
// not a dangling reference, can be what clears it.
func TestMigrateV1Alpha9DropsDerivedLocation(t *testing.T) {
	site := newV1Alpha9Site(t)
	racked := site.inventory.Devices[site.ids.nodeID]
	stale := site.inventory.Devices[site.ids.switchID]
	stale.Location = site.otherSiteID
	standalone := site.store(&devicetypes.CaniDeviceType{Name: "cdu-1", Type: devicetypes.TypeCDU, Location: site.otherSiteID})

	report := foldProviderMetadata(site.inventory)

	assertReported(t, report)
	if racked.Location != site.rowID {
		t.Errorf("racked device location = %s, want its rack's %s kept", racked.Location, site.rowID)
	}
	if stale.Location != uuid.Nil {
		t.Errorf("stale location = %s, want it cleared", stale.Location)
	}
	if standalone.Location != site.otherSiteID {
		t.Errorf("device with no parent: location = %s, want %s kept", standalone.Location, site.otherSiteID)
	}
	if err := site.inventory.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("rebuild after the migration: %v", err)
	}
	if stale.Location != site.rowID {
		t.Errorf("after the rebuild the device is in %s, want its rack's %s", stale.Location, site.rowID)
	}
}
