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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// deviceTypeCreatedJSON renders the object Nautobot returns from a successful
// device-type create. Note the field is "model" (not "name").
func deviceTypeCreatedJSON(id uuid.UUID, model string) string {
	return fmt.Sprintf(`{"id":%q,"model":%q,"display":%q}`, id.String(), model, model)
}

// seedManufacturer caches a manufacturer so GetOrCreateManufacturer resolves it
// without issuing any HTTP request.
func seedManufacturer(e *Exporter, name string) {
	seedManufacturerWithID(e, name, uuid.New())
}

// seedManufacturerWithID caches a manufacturer with a caller-chosen UUID so
// tests can assert the outbound device-type request references that exact ID.
func seedManufacturerWithID(e *Exporter, name string, id uuid.UUID) {
	e.Cache.manufacturersMu.Lock()
	e.Cache.manufacturers[name] = &CachedItem{ID: id, Name: name}
	e.Cache.manufacturersMu.Unlock()
}

// -----------------------------------------------------------------------------
// CreateDeviceTypeFromCaniDevice — create a device type from inventory data.
// -----------------------------------------------------------------------------

// TestCreateDeviceTypeFromCaniDevice_CreatesAndCaches verifies that
// CreateDeviceTypeFromCaniDevice POSTs a device-type built from inventory
// fields and caches the result under the device slug.
//
// Why it matters: device-types must exist before the devices that reference
// them; caching by slug lets the rest of the export resolve the type without
// re-querying or duplicating it.
// Inputs: a CaniDeviceType{Slug, Model, Manufacturer:"HPE", UHeight:2,
// IsFullDepth}, a seeded manufacturer, and a 201 response. Outputs: a
// *CachedItem plus a deviceTypes[slug] cache entry.
// Data choice: a realistic 2U full-depth HPE server exercises the optional
// UHeight/IsFullDepth fields, and asserting the slug cache key proves the
// idempotency hook downstream lookups rely on.
func TestCreateDeviceTypeFromCaniDevice_CreatesAndCaches(t *testing.T) {
	dtID := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated,
		deviceTypeCreatedJSON(dtID, "ProLiant DL380")))
	defer cleanup()
	seedManufacturer(e, "HPE")

	device := &devicetypes.CaniDeviceType{
		Slug:         "hpe-dl380",
		Model:        "ProLiant DL380",
		Manufacturer: "HPE",
		UHeight:      2,
		IsFullDepth:  true,
	}

	item, err := e.Cache.CreateDeviceTypeFromCaniDevice(device)
	if err != nil {
		t.Fatalf("CreateDeviceTypeFromCaniDevice() error = %v", err)
	}
	if item == nil || item.ID != dtID || item.Name != "ProLiant DL380" {
		t.Errorf("expected device type %s, got %+v", dtID, item)
	}

	e.Cache.deviceTypesMu.RLock()
	cached, ok := e.Cache.deviceTypes["hpe-dl380"]
	e.Cache.deviceTypesMu.RUnlock()
	if !ok || cached.ID != dtID {
		t.Errorf("expected device type cached under slug, got %+v (ok=%v)", cached, ok)
	}
}

// TestCreateDeviceTypeFromCaniDevice_DefaultsMissingManufacturer verifies that
// when the device has no Manufacturer, CreateDeviceTypeFromCaniDevice falls
// back to "Unknown" and still creates the type.
//
// Why it matters: inventory data is often incomplete; defaulting keeps the
// export resilient so a missing manufacturer does not block device-type
// creation in Nautobot.
// Inputs: a CaniDeviceType with an empty Manufacturer, with "Unknown" seeded so
// the fallback resolves from cache. Outputs: the created *CachedItem and a POST
// body whose manufacturer.id matches the cached Unknown manufacturer.
// Data choice: seeding "Unknown" with a fixed UUID makes the fallback visible in
// the generated Nautobot request payload, not merely in a successful response.
func TestCreateDeviceTypeFromCaniDevice_DefaultsMissingManufacturer(t *testing.T) {
	dtID := uuid.New()
	unknownID := uuid.New()
	var calls int
	var captured []byte
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		captured, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(deviceTypeCreatedJSON(dtID, "generic")))
	})
	defer cleanup()
	seedManufacturerWithID(e, "Unknown", unknownID)

	device := &devicetypes.CaniDeviceType{Slug: "generic", Model: "generic"}
	item, err := e.Cache.CreateDeviceTypeFromCaniDevice(device)
	if err != nil {
		t.Fatalf("CreateDeviceTypeFromCaniDevice() error = %v", err)
	}
	if item == nil || item.ID != dtID {
		t.Errorf("expected device type %s, got %+v", dtID, item)
	}
	if calls != 1 {
		t.Fatalf("expected exactly one device-type create POST, got %d", calls)
	}

	var payload struct {
		Model        string    `json:"model"`
		Manufacturer wireIDRef `json:"manufacturer"`
	}
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("decode device-type create payload: %v\nbody: %s", err, captured)
	}
	if payload.Model != "generic" {
		t.Errorf("model = %q, want generic", payload.Model)
	}
	if payload.Manufacturer.ID != unknownID.String() {
		t.Errorf("manufacturer.id = %q, want %s", payload.Manufacturer.ID, unknownID)
	}
}

// TestCreateDeviceTypeFromCaniDevice_ReturnsErrorOnNon201 verifies that
// CreateDeviceTypeFromCaniDevice returns an error when the create responds with
// 400.
//
// Why it matters: a failed device-type create must abort so devices are not
// exported referencing a type that was never persisted in Nautobot.
// Inputs: a seeded manufacturer and a 400 response. Outputs: a non-nil error.
// Data choice: a 400 with a {"detail"} body models a Nautobot validation
// rejection, and the seeded manufacturer isolates the failure to the create
// call.
func TestCreateDeviceTypeFromCaniDevice_ReturnsErrorOnNon201(t *testing.T) {
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusBadRequest, `{"detail":"bad"}`))
	defer cleanup()
	seedManufacturer(e, "HPE")

	device := &devicetypes.CaniDeviceType{Slug: "hpe-dl380", Model: "ProLiant DL380", Manufacturer: "HPE"}
	if _, err := e.Cache.CreateDeviceTypeFromCaniDevice(device); err == nil {
		t.Fatal("expected an error when device-type create responds with 400")
	}
}

// TestCreateDeviceTypeFromCaniDevice_ReturnsErrorForNilDevice verifies that
// CreateDeviceTypeFromCaniDevice returns an error for a nil device without
// making any HTTP call.
//
// Why it matters: guarding nil input prevents a panic mid-export and signals a
// programming or data error early.
// Inputs: a nil device on a cache built with a nil client. Outputs: a non-nil
// error.
// Data choice: a nil client guarantees the test fails loudly (panic) if the
// guard is missing and any request is attempted.
func TestCreateDeviceTypeFromCaniDevice_ReturnsErrorForNilDevice(t *testing.T) {
	cache := NewLookupCache(nil) // nil client: any HTTP would panic
	if _, err := cache.CreateDeviceTypeFromCaniDevice(nil); err == nil {
		t.Fatal("expected an error for a nil device")
	}
}

// TestCreateDeviceTypeFromCaniDevice_ReturnsErrorForEmptySlug verifies that
// CreateDeviceTypeFromCaniDevice returns an error when the device's Slug is
// empty.
//
// Why it matters: the slug is the cache key device-types are stored and
// resolved under; without it the export could not link devices to their type,
// so the input must be rejected.
// Inputs: a CaniDeviceType{Slug:""} on a nil-client cache. Outputs: a non-nil
// error.
// Data choice: an empty slug with a nil client confirms the guard fires before
// any HTTP work.
func TestCreateDeviceTypeFromCaniDevice_ReturnsErrorForEmptySlug(t *testing.T) {
	cache := NewLookupCache(nil)
	device := &devicetypes.CaniDeviceType{Slug: "", Model: "x"}
	if _, err := cache.CreateDeviceTypeFromCaniDevice(device); err == nil {
		t.Fatal("expected an error for an empty slug")
	}
}

// -----------------------------------------------------------------------------
// CreateDeviceTypeFromLocal — create a device type from the local YAML library.
// -----------------------------------------------------------------------------

// TestCreateDeviceTypeFromLocal_ReturnsErrorForUnknownSlug verifies that
// CreateDeviceTypeFromLocal errors when the slug is absent from the builtin
// device-type library.
//
// Why it matters: this path seeds Nautobot from cani's local YAML library; an
// unknown slug means there is no definition to export, which must fail rather
// than create an empty type.
// Inputs: a bogus slug on a nil-client cache. Outputs: a non-nil error.
// Data choice: an obviously fake slug guarantees a library miss without
// depending on which builtins happen to exist.
func TestCreateDeviceTypeFromLocal_ReturnsErrorForUnknownSlug(t *testing.T) {
	cache := NewLookupCache(nil)
	if _, err := cache.CreateDeviceTypeFromLocal("definitely-not-a-real-slug"); err == nil {
		t.Fatal("expected an error for a slug missing from the local library")
	}
}

// TestCreateDeviceTypeFromLocal_CreatesFromLibrary verifies that
// CreateDeviceTypeFromLocal looks up a local device-type by slug, creates it in
// Nautobot, and caches it under that slug.
//
// Why it matters: cani ships a curated hardware library; exporting from it
// gives Nautobot accurate device-types without manual data entry.
// Inputs: a test-only registered slug with manufacturer "Acme" (seeded so it
// resolves from cache) and a 201 response. Outputs: the created *CachedItem plus
// a deviceTypes[slug] entry.
// Data choice: registering a unique local type keeps the test independent of
// embedded fixture contents while still driving the local-library lookup path.
func TestCreateDeviceTypeFromLocal_CreatesFromLibrary(t *testing.T) {
	const slug = "export-local-test-device-type"
	const manufacturer = "Acme"
	const partNumber = "EXPORT-LOCAL-TEST"
	devicetypes.RegisterDeviceType(devicetypes.CaniDeviceType{
		Slug:         slug,
		Model:        "Export Local Test Device",
		Manufacturer: manufacturer,
		PartNumber:   partNumber,
		UHeight:      2,
	})
	t.Cleanup(func() {
		delete(devicetypes.All(), slug)
		delete(devicetypes.ByPartNumber(), partNumber)
	})

	dtID := uuid.New()
	var calls int
	e, cleanup := newExporterWithServer(t, jsonHandler(&calls, http.StatusCreated,
		deviceTypeCreatedJSON(dtID, "Library Model")))
	defer cleanup()
	seedManufacturer(e, manufacturer)

	item, err := e.Cache.CreateDeviceTypeFromLocal(slug)
	if err != nil {
		t.Fatalf("CreateDeviceTypeFromLocal() error = %v", err)
	}
	if item == nil || item.ID != dtID {
		t.Errorf("expected device type %s, got %+v", dtID, item)
	}

	e.Cache.deviceTypesMu.RLock()
	_, ok := e.Cache.deviceTypes[slug]
	e.Cache.deviceTypesMu.RUnlock()
	if !ok {
		t.Errorf("expected device type cached under slug %q", slug)
	}
}
