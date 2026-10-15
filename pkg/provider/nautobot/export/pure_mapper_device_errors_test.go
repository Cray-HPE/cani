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
package export

import (
	"errors"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ---------- DeviceMapper nil device ----------

// TestMapToNautobotDeviceNilDevice verifies that mapping a nil device returns
// the exact error "device is nil" rather than panicking.
//
// Why it matters: the bulk-create path maps many devices in a loop; a nil entry
// must fail fast with a clear message instead of crashing the whole export.
// Inputs: a nil *CaniDeviceType. Outputs: nil request and the sentinel error.
// Data choice: asserting the exact error string locks the contract callers and
// tests rely on for nil handling.
func TestMapToNautobotDeviceNilDevice(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	_, err := mapper.MapToNautobotDevice(nil)
	if err == nil {
		t.Fatal("expected error for nil device")
	}
	if err.Error() != "device is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// TestMapToWritableDeviceRequestNilDevice verifies the single-create mapper
// also rejects a nil device with the "device is nil" error.
//
// Why it matters: the single-device create path is a separate method from the
// bulk path and must enforce the same nil guard so neither can panic.
// Inputs: a nil *CaniDeviceType. Outputs: nil request and the sentinel error.
// Data choice: mirrors TestMapToNautobotDeviceNilDevice to prove both mapper
// entry points share identical nil-safety behavior.
func TestMapToWritableDeviceRequestNilDevice(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})

	_, err := mapper.MapToWritableDeviceRequest(nil)
	if err == nil {
		t.Fatal("expected error for nil device")
	}
	if err.Error() != "device is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// ---------- MapToNautobotDevice error propagation ----------

// TestMapToNautobotDeviceNoDeviceType verifies that MapToNautobotDevice returns
// an error wrapping ErrDeviceUnclassified when the device type cannot be
// resolved (empty slug/model).
//
// Why it matters: the bulk export must abort a device whose type is unknown
// rather than emit a request with a missing/zero device-type reference.
// Inputs: a CaniDeviceType with empty Slug and a non-strict mapper whose
// location/status/role are seeded. Outputs: an error matching
// ErrDeviceUnclassified.
// Data choice: only the device-type source is removed (others seeded) so the
// failure is attributable to type resolution.
func TestMapToNautobotDeviceNoDeviceType(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	statusID := uuid.New()
	roleID := uuid.New()

	cache.locationsMu.Lock()
	cache.locations["DC"] = &CachedItem{ID: locID, Name: "DC"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()
	cache.rolesMu.Lock()
	cache.roles["Server"] = &CachedItem{ID: roleID, Name: "Server"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		Strict:          false,
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "no-type-device",
		Slug: "", // no slug
	}

	_, err := mapper.MapToNautobotDevice(dev)
	if err == nil {
		t.Fatal("expected error when device type can't be resolved")
	}
	if !errors.Is(err, ErrDeviceUnclassified) {
		t.Fatalf("expected ErrDeviceUnclassified, got %v", err)
	}
}

// TestMapToNautobotDeviceNoLocation verifies that MapToNautobotDevice returns a
// non-nil error when the location cannot be resolved.
//
// Why it matters: a device with a resolvable type but no location must not be
// exported, since Nautobot requires a location reference.
// Inputs: a CaniDeviceType (Slug "dt-noloc") with its device type cached but a
// mapper that has no DefaultLocation and no cached location. Outputs: a non-nil
// error.
// Data choice: caching the device type while leaving location unsatisfiable
// pinpoints location resolution as the failing step.
func TestMapToNautobotDeviceNoLocation(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["dt-noloc"] = &CachedItem{ID: dtID, Name: "DT-NoLoc"}
	cache.deviceTypesMu.Unlock()

	// No location cached, no default, no createLocations → error
	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "noloc-device",
		Slug: "dt-noloc",
	}

	_, err := mapper.MapToNautobotDevice(dev)
	if err == nil {
		t.Fatal("expected error when location can't be resolved")
	}
}

// TestMapToNautobotDeviceNoStatus verifies that MapToNautobotDevice returns a
// non-nil error when the status cannot be resolved.
//
// Why it matters: status is mandatory on Nautobot devices, so the exporter must
// stop before building a request that lacks one.
// Inputs: a CaniDeviceType (Slug "dt-nostat") with device type and location
// cached but a mapper with no DefaultStatus. Outputs: a non-nil error.
// Data choice: seeding type and location but not status isolates the status step
// as the cause of the failure.
func TestMapToNautobotDeviceNoStatus(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["dt-nostat"] = &CachedItem{ID: dtID, Name: "DT-NoStat"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-S"] = &CachedItem{ID: locID, Name: "DC-S"}
	cache.locationsMu.Unlock()

	// No status cached, no default → error
	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-S",
		DefaultStatus:   "",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "nostat-device",
		Slug: "dt-nostat",
	}

	_, err := mapper.MapToNautobotDevice(dev)
	if err == nil {
		t.Fatal("expected error when status can't be resolved")
	}
}

// TestMapToNautobotDeviceNoRole verifies that MapToNautobotDevice returns a
// non-nil error when the role cannot be resolved.
//
// Why it matters: role is required on Nautobot devices, so a device missing it
// must fail mapping rather than export an incomplete request.
// Inputs: a CaniDeviceType (Slug "dt-norole") with device type, location, and
// status cached but a mapper with no DefaultRole. Outputs: a non-nil error.
// Data choice: satisfying every reference except role pins the failure to the
// role resolution step.
func TestMapToNautobotDeviceNoRole(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()

	cache.deviceTypesMu.Lock()
	cache.deviceTypes["dt-norole"] = &CachedItem{ID: dtID, Name: "DT-NoRole"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-R"] = &CachedItem{ID: locID, Name: "DC-R"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	// No role cached, no default → error
	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-R",
		DefaultStatus:   "Active",
		DefaultRole:     "",
	})

	dev := &devicetypes.CaniDeviceType{
		Name: "norole-device",
		Slug: "dt-norole",
	}

	_, err := mapper.MapToNautobotDevice(dev)
	if err == nil {
		t.Fatal("expected error when role can't be resolved")
	}
}

// ---------- MapToWritableDeviceRequest error propagation ----------

// TestMapToWritableDeviceRequestNoDeviceType verifies that
// MapToWritableDeviceRequest returns a non-nil error when the device type cannot
// be resolved.
//
// Why it matters: the single-create export path must enforce the same device-type
// requirement as the bulk path so no under-specified device reaches Nautobot.
// Inputs: a CaniDeviceType with no slug and a non-strict mapper carrying default
// location/status/role names. Outputs: a non-nil error.
// Data choice: an unclassifiable device with defaults set elsewhere isolates the
// device-type resolution failure on the single-create code path.
func TestMapToWritableDeviceRequestNoDeviceType(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{
		Strict:          false,
		DefaultLocation: "DC",
		DefaultStatus:   "Active",
		DefaultRole:     "Server",
	})

	dev := &devicetypes.CaniDeviceType{Name: "fail-dt"}
	_, err := mapper.MapToWritableDeviceRequest(dev)
	if err == nil {
		t.Fatal("expected error when device type can't be resolved")
	}
}

// TestMapToWritableDeviceRequestNoLocation verifies that
// MapToWritableDeviceRequest returns a non-nil error when the location cannot be
// resolved.
//
// Why it matters: the single-create path must reject a device with no resolvable
// location, matching Nautobot's location requirement.
// Inputs: a CaniDeviceType (Slug "wr-noloc") with its device type cached but a
// mapper with empty DefaultLocation. Outputs: a non-nil error.
// Data choice: caching only the device type leaves location unsatisfiable,
// pinpointing the location step on the single-create path.
func TestMapToWritableDeviceRequestNoLocation(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["wr-noloc"] = &CachedItem{ID: dtID, Name: "WR-NoLoc"}
	cache.deviceTypesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: ""})
	dev := &devicetypes.CaniDeviceType{Name: "wr-fail-loc", Slug: "wr-noloc"}
	_, err := mapper.MapToWritableDeviceRequest(dev)
	if err == nil {
		t.Fatal("expected error when location can't be resolved")
	}
}

// TestMapToWritableDeviceRequestNoStatus verifies that
// MapToWritableDeviceRequest returns a non-nil error when the status cannot be
// resolved.
//
// Why it matters: the single-create path must enforce the mandatory status field
// just like the bulk path before issuing a create.
// Inputs: a CaniDeviceType (Slug "wr-nostat") with device type and location
// cached but a mapper with empty DefaultStatus. Outputs: a non-nil error.
// Data choice: seeding type and location but not status isolates status as the
// failing step on the single-create path.
func TestMapToWritableDeviceRequestNoStatus(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["wr-nostat"] = &CachedItem{ID: dtID, Name: "WR-NoStat"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-WS"] = &CachedItem{ID: locID, Name: "DC-WS"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-WS",
		DefaultStatus:   "",
	})
	dev := &devicetypes.CaniDeviceType{Name: "wr-fail-stat", Slug: "wr-nostat"}
	_, err := mapper.MapToWritableDeviceRequest(dev)
	if err == nil {
		t.Fatal("expected error when status can't be resolved")
	}
}

// TestMapToWritableDeviceRequestNoRole verifies that MapToWritableDeviceRequest
// returns a non-nil error when the role cannot be resolved.
//
// Why it matters: the single-create path must enforce the required role field so
// no role-less device is created in Nautobot.
// Inputs: a CaniDeviceType (Slug "wr-norole") with device type, location, and
// status cached but a mapper with empty DefaultRole. Outputs: a non-nil error.
// Data choice: satisfying every reference except role pins the failure to the
// role step on the single-create path.
func TestMapToWritableDeviceRequestNoRole(t *testing.T) {
	cache := NewLookupCache(nil)
	dtID := uuid.New()
	locID := uuid.New()
	statusID := uuid.New()
	cache.deviceTypesMu.Lock()
	cache.deviceTypes["wr-norole"] = &CachedItem{ID: dtID, Name: "WR-NoRole"}
	cache.deviceTypesMu.Unlock()
	cache.locationsMu.Lock()
	cache.locations["DC-WR2"] = &CachedItem{ID: locID, Name: "DC-WR2"}
	cache.locationsMu.Unlock()
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: statusID, Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{
		DefaultLocation: "DC-WR2",
		DefaultStatus:   "Active",
		DefaultRole:     "",
	})
	dev := &devicetypes.CaniDeviceType{Name: "wr-fail-role", Slug: "wr-norole"}
	_, err := mapper.MapToWritableDeviceRequest(dev)
	if err == nil {
		t.Fatal("expected error when role can't be resolved")
	}
}
