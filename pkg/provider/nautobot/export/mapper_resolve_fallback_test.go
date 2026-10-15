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
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ---------- resolveDeviceType error paths ----------

// TestResolveDeviceTypeStrictNoSlug verifies that resolveDeviceType returns the
// error "device type slug is required" when Strict is true and the device has no
// slug or model.
//
// Why it matters: in strict mode the exporter must refuse to guess a device type,
// failing loudly so unclassified hardware is never silently mis-exported.
// Inputs: a CaniDeviceType with empty Slug and Model and a Strict mapper.
// Outputs: a non-nil error with the exact required-slug message.
// Data choice: empty slug+model with Strict=true is the precise condition that
// triggers the hard required-slug failure, and the message is asserted verbatim.
func TestResolveDeviceTypeStrictNoSlug(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{Strict: true})
	dev := &devicetypes.CaniDeviceType{Slug: "", Model: ""}

	_, err := mapper.resolveDeviceType(dev)
	if err == nil {
		t.Fatal("expected error for strict mode with no slug/model")
	}
	if err.Error() != "device type slug is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestResolveDeviceTypeNonStrictNoSlug verifies that resolveDeviceType returns a
// value satisfying errors.Is(err, ErrDeviceUnclassified) when Strict is false and
// the device has no slug or model.
//
// Why it matters: non-strict export must tag unclassifiable devices with a known
// sentinel so callers can skip them gracefully instead of erroring hard.
// Inputs: a CaniDeviceType with empty Slug and Model and a non-strict mapper.
// Outputs: an error wrapping ErrDeviceUnclassified.
// Data choice: empty slug+model with Strict=false isolates the sentinel-error
// branch; using errors.Is tolerates any wrapping around the sentinel.
func TestResolveDeviceTypeNonStrictNoSlug(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{Strict: false})
	dev := &devicetypes.CaniDeviceType{Slug: "", Model: ""}

	_, err := mapper.resolveDeviceType(dev)
	if !errors.Is(err, ErrDeviceUnclassified) {
		t.Errorf("expected ErrDeviceUnclassified, got %v", err)
	}
}

// ---------- resolveStatus error path ----------

// TestResolveStatusNoDefaultNoCreate verifies that resolveStatus returns the
// error "status is required (use --default-status)" when the device has no status
// and the mapper has no DefaultStatus.
//
// Why it matters: Nautobot devices require a status; the exporter must fail with
// actionable guidance rather than push a device with no status.
// Inputs: an empty CaniDeviceType and a mapper with DefaultStatus "". Outputs: a
// non-nil error with the exact guidance message.
// Data choice: empty device status plus empty default removes every status source,
// forcing the required-status error, which is asserted verbatim.
func TestResolveStatusNoDefaultNoCreate(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultStatus: ""})
	dev := &devicetypes.CaniDeviceType{}

	_, err := mapper.resolveStatus(dev)
	if err == nil {
		t.Fatal("expected error when no status and no default")
	}
	if err.Error() != "status is required (use --default-status)" {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- resolveRole error path ----------

// TestResolveRoleNoDefaultNoCreate verifies that resolveRole returns the error
// "role is required (use --default-role)" when the device has no role and the
// mapper has no DefaultRole.
//
// Why it matters: Nautobot devices require a role; the exporter must fail with a
// clear hint instead of exporting a role-less device.
// Inputs: an empty CaniDeviceType and a mapper with DefaultRole "". Outputs: a
// non-nil error with the exact guidance message.
// Data choice: empty role plus empty default eliminates every role source so the
// required-role error path is the only outcome, asserted verbatim.
func TestResolveRoleNoDefaultNoCreate(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultRole: ""})
	dev := &devicetypes.CaniDeviceType{}

	_, err := mapper.resolveRole(dev)
	if err == nil {
		t.Fatal("expected error when no role and no default")
	}
	if err.Error() != "role is required (use --default-role)" {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- resolveLocation error path ----------

// TestResolveLocationNoDefaultNoCreate verifies that resolveLocation returns an
// actionable "location is required" error when the device has no location and
// the mapper has no DefaultLocation.
//
// Why it matters: every exported device must land at a Nautobot location; absent
// any source the exporter must fail with actionable guidance, not a blank location.
// Inputs: an empty CaniDeviceType and a mapper with DefaultLocation "". Outputs: a
// non-nil error naming both remedies (rack location, --default-location).
// Data choice: empty location plus empty default removes all location sources,
// forcing the required-location error.
func TestResolveLocationNoDefaultNoCreate(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{DefaultLocation: ""})
	dev := &devicetypes.CaniDeviceType{}

	_, err := mapper.resolveLocation(dev)
	if err == nil {
		t.Fatal("expected error when no location and no default")
	}
	if !strings.Contains(err.Error(), "location is required") || !strings.Contains(err.Error(), "--default-location") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- resolveStatus auto-create path ----------

// TestResolveStatusAutoCreateActive verifies resolveStatus falls back to the
// "Active" status when auto-create is enabled and the device sets no status.
//
// Why it matters: every Nautobot device requires a status, so cani must supply a
// sensible default rather than fail the export.
// Inputs: a device with no status, createStatuses=true, and a cache pre-seeded
// with "Active". Outputs: the cached "Active" CachedItem.
// Data choice: seeding the cache lets the default resolve without hitting the
// nil HTTP client that a real create would use.
func TestResolveStatusAutoCreateActive(t *testing.T) {
	cache := NewLookupCache(nil)
	cache.SetCreateStatuses(true)
	// Pre-populate cache with "Active" so the lookup succeeds
	cache.statusesMu.Lock()
	cache.statuses["Active"] = &CachedItem{ID: uuid.New(), Name: "Active"}
	cache.statusesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{} // no status set
	item, err := mapper.resolveStatus(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Active" {
		t.Errorf("expected auto-created 'Active', got %q", item.Name)
	}
}

// TestResolveStatusNoAutoCreateFails verifies resolveStatus errors when the
// device has no status and auto-create is disabled.
//
// Why it matters: without a status and without permission to create one, the
// export must stop rather than push an invalid device to Nautobot.
// Inputs: a device with no status and the default createStatuses=false cache.
// Outputs: a non-nil error.
// Data choice: leaving both the device status and default empty isolates the
// failure path.
func TestResolveStatusNoAutoCreateFails(t *testing.T) {
	cache := NewLookupCache(nil)
	// createStatuses is false by default
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{} // no status set, no default
	_, err := mapper.resolveStatus(dev)
	if err == nil {
		t.Fatal("expected error when no status and no auto-create")
	}
}

// ---------- resolveRole auto-create path ----------

// TestResolveRoleAutoCreateGeneric verifies resolveRole falls back to the
// "Generic" role when auto-create is enabled and the device sets no role.
//
// Why it matters: Nautobot devices need a role, so cani defaults unclassified
// devices to "Generic" instead of aborting the export.
// Inputs: a device with no role, createRoles=true, and a cache seeded with
// "Generic". Outputs: the cached "Generic" CachedItem.
// Data choice: the seeded cache resolves the default without a live API call.
func TestResolveRoleAutoCreateGeneric(t *testing.T) {
	cache := NewLookupCache(nil)
	cache.SetCreateRoles(true)
	cache.rolesMu.Lock()
	cache.roles["Generic"] = &CachedItem{ID: uuid.New(), Name: "Generic"}
	cache.rolesMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{} // no role set
	item, err := mapper.resolveRole(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Generic" {
		t.Errorf("expected auto-created 'Generic', got %q", item.Name)
	}
}

// TestResolveRoleNoAutoCreateFails verifies resolveRole errors when the device
// has no role and auto-create is disabled.
//
// Why it matters: a device with no resolvable role must fail fast rather than be
// exported into Nautobot without one.
// Inputs: a bare device and a cache with createRoles=false.
// Outputs: a non-nil error.
// Data choice: the empty device and default cache isolate the no-role error
// branch.
func TestResolveRoleNoAutoCreateFails(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{}
	_, err := mapper.resolveRole(dev)
	if err == nil {
		t.Fatal("expected error when no role and no auto-create")
	}
}

// ---------- resolveLocation without any location ----------

// TestResolveLocationWithoutLocationFails verifies resolveLocation returns an
// error, even with auto-create enabled, when a device has no location source.
//
// Why it matters: the exporter used to invent a "Default" location with no
// datastore counterpart; a device without a location must be reported instead.
// Inputs: a device with no location, createLocations=true, and a cache seeded
// with "Default". Outputs: a non-nil error and no lookup of "Default".
// Data choice: seeding "Default" proves the old fallback is no longer taken.
func TestResolveLocationWithoutLocationFails(t *testing.T) {
	cache := NewLookupCache(nil)
	cache.SetCreateLocations(true)
	cache.locationsMu.Lock()
	cache.locations["Default"] = &CachedItem{ID: uuid.New(), Name: "Default"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{} // no location
	if _, err := mapper.resolveLocation(dev); err == nil {
		t.Fatal("expected an error for a device with no location source")
	}
}

// TestResolveLocationNoAutoCreateFails verifies resolveLocation errors when no
// location is set and auto-create is disabled.
//
// Why it matters: exporting a device with no resolvable location would orphan it
// in Nautobot, so the mapper must refuse.
// Inputs: a bare device and a cache with createLocations=false.
// Outputs: a non-nil error.
// Data choice: the empty device and default cache isolate the failure path.
func TestResolveLocationNoAutoCreateFails(t *testing.T) {
	cache := NewLookupCache(nil)
	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{}
	_, err := mapper.resolveLocation(dev)
	if err == nil {
		t.Fatal("expected error when no location and no auto-create")
	}
}

// TestResolveLocationFromMetadataExplicit verifies resolveLocation honors an
// explicit ProviderMetadata["location"] over any default.
//
// Why it matters: when cani data names a specific location, the export must place
// the device there rather than at a fallback.
// Inputs: a device whose ProviderMetadata sets location="Lab-42", with "Lab-42"
// seeded in the cache. Outputs: the CachedItem whose ID matches the seeded one.
// Data choice: asserting on the seeded UUID proves the named location (not a
// default) was selected.
func TestResolveLocationFromMetadataExplicit(t *testing.T) {
	cache := NewLookupCache(nil)
	locID := uuid.New()
	cache.locationsMu.Lock()
	cache.locations["Lab-42"] = &CachedItem{ID: locID, Name: "Lab-42"}
	cache.locationsMu.Unlock()

	mapper := NewDeviceMapper(cache, &MapperOpts{})
	dev := &devicetypes.CaniDeviceType{
		ObjectMeta: devicetypes.ObjectMeta{
			ProviderMetadata: map[string]any{
				"location": "Lab-42",
			},
		},
	}
	item, err := mapper.resolveLocation(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != locID {
		t.Errorf("expected location ID %s, got %s", locID, item.ID)
	}
}
