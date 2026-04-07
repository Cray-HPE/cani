package devicetypes

// Test coverage for constructors.go
//
// | Function               | Happy-path test                              | Failure test                                  |
// |------------------------|----------------------------------------------|-----------------------------------------------|
// | NewLocation            | TestNewLocationReturnsEmptyStruct             | TestNewLocationFieldsAreZeroValues            |
// | NewDefaultLocation     | TestNewDefaultLocationSetsDefaults            | TestNewDefaultLocationIDIsUnique              |
// | generateCaniName       | TestGenerateCaniNameMatchesPrefix             | TestGenerateCaniNameIsNotEmpty                |
// | NewDeviceFromSlug      | TestNewDeviceFromSlugHappyPath                | TestNewDeviceFromSlugUnknownSlug              |
// | NewDeviceFromPartNumber| TestNewDeviceFromPartNumberHappyPath          | TestNewDeviceFromPartNumberUnknownPN          |
// | NewRackFromSlug        | TestNewRackFromSlugHappyPath                  | TestNewRackFromSlugUnknownSlug                |
// | NewRackFromPartNumber  | TestNewRackFromPartNumberHappyPath            | TestNewRackFromPartNumberUnknownPN            |
// | NewModuleFromSlug      | TestNewModuleFromSlugHappyPath                | TestNewModuleFromSlugUnknownSlug              |
// | NewModuleFromPartNumber| TestNewModuleFromPartNumberHappyPath          | TestNewModuleFromPartNumberUnknownPN          |
// | NewCableFromSlug       | TestNewCableFromSlugHappyPath                 | TestNewCableFromSlugUnknownSlug               |
// | NewCableFromPartNumber | TestNewCableFromPartNumberHappyPath           | TestNewCableFromPartNumberUnknownPN           |
// | NewFruFromSlug         | TestNewFruFromSlugHappyPath                   | TestNewFruFromSlugUnknownSlug                 |
// | NewFruFromPartNumber   | TestNewFruFromPartNumberHappyPath             | TestNewFruFromPartNumberUnknownPN             |

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ---------- NewLocation ----------

func TestNewLocationReturnsEmptyStruct(t *testing.T) {
	loc := NewLocation()
	if loc == nil {
		t.Fatal("expected non-nil location")
	}
}

func TestNewLocationFieldsAreZeroValues(t *testing.T) {
	loc := NewLocation()
	if loc.ID != uuid.Nil {
		t.Errorf("expected zero UUID, got %s", loc.ID)
	}
	if loc.Name != "" {
		t.Errorf("expected empty name, got %q", loc.Name)
	}
	if loc.Status != "" {
		t.Errorf("expected empty status, got %q", loc.Status)
	}
}

// ---------- NewDefaultLocation ----------

func TestNewDefaultLocationSetsDefaults(t *testing.T) {
	loc := NewDefaultLocation()
	if loc == nil {
		t.Fatal("expected non-nil location")
	}
	if loc.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if loc.Name != "default-cani" {
		t.Errorf("expected name %q, got %q", "default-cani", loc.Name)
	}
	if loc.LocationType != "site" {
		t.Errorf("expected location type %q, got %q", "site", loc.LocationType)
	}
	if loc.Status != "Active" {
		t.Errorf("expected status %q, got %q", "Active", loc.Status)
	}
}

func TestNewDefaultLocationIDIsUnique(t *testing.T) {
	a := NewDefaultLocation()
	b := NewDefaultLocation()
	if a.ID == b.ID {
		t.Error("expected two calls to produce different UUIDs")
	}
}

// ---------- generateCaniName ----------

func TestGenerateCaniNameMatchesPrefix(t *testing.T) {
	name := generateCaniName()
	if !strings.HasPrefix(name, "cani-device-") {
		t.Errorf("expected prefix %q, got %q", "cani-device-", name)
	}
}

func TestGenerateCaniNameIsNotEmpty(t *testing.T) {
	name := generateCaniName()
	if name == "" {
		t.Error("expected non-empty name")
	}
	if len(name) <= len("cani-device-") {
		t.Errorf("expected name longer than prefix, got %q", name)
	}
}

// ---------- NewDeviceFromSlug ----------

func TestNewDeviceFromSlugHappyPath(t *testing.T) {
	resetRegistries()
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-server",
		PartNumber:   "PN-SRV",
		Manufacturer: "Acme",
		Model:        "Server1",
	})

	dev, err := NewDeviceFromSlug("test-server")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dev.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if dev.Status != "Staged" {
		t.Errorf("expected status %q, got %q", "Staged", dev.Status)
	}
	if !strings.HasPrefix(dev.Name, "cani-device-") {
		t.Errorf("expected generated name with prefix, got %q", dev.Name)
	}
	if dev.Manufacturer != "Acme" {
		t.Errorf("expected manufacturer %q, got %q", "Acme", dev.Manufacturer)
	}
}

func TestNewDeviceFromSlugUnknownSlug(t *testing.T) {
	resetRegistries()

	dev, err := NewDeviceFromSlug("nonexistent-slug")
	if err == nil {
		t.Fatal("expected error for unknown slug")
	}
	if dev != nil {
		t.Error("expected nil device on error")
	}
	if !strings.Contains(err.Error(), "nonexistent-slug") {
		t.Errorf("error should mention the slug, got: %v", err)
	}
}

// ---------- NewDeviceFromPartNumber ----------

func TestNewDeviceFromPartNumberHappyPath(t *testing.T) {
	resetRegistries()
	RegisterDeviceType(CaniDeviceType{
		Slug:         "test-switch",
		PartNumber:   "PN-SW-100",
		Manufacturer: "Acme",
		Model:        "Switch100",
	})

	dev, err := NewDeviceFromPartNumber("PN-SW-100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dev.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if dev.Status != "Staged" {
		t.Errorf("expected status %q, got %q", "Staged", dev.Status)
	}
	if !strings.HasPrefix(dev.Name, "cani-device-") {
		t.Errorf("expected generated name with prefix, got %q", dev.Name)
	}
}

func TestNewDeviceFromPartNumberUnknownPN(t *testing.T) {
	resetRegistries()

	dev, err := NewDeviceFromPartNumber("BOGUS-PN")
	if err == nil {
		t.Fatal("expected error for unknown part number")
	}
	if dev != nil {
		t.Error("expected nil device on error")
	}
	if !strings.Contains(err.Error(), "BOGUS-PN") {
		t.Errorf("error should mention the part number, got: %v", err)
	}
}

// ---------- NewRackFromSlug ----------

func TestNewRackFromSlugHappyPath(t *testing.T) {
	resetRegistries()
	RegisterRackType(CaniRackType{
		Slug:       "test-rack-42u",
		PartNumber: "PN-RACK-42",
		UHeight:    42,
	})

	rack, err := NewRackFromSlug("test-rack-42u")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rack.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if rack.Status != "Active" {
		t.Errorf("expected status %q, got %q", "Active", rack.Status)
	}
	if rack.OccupiedSlots == nil {
		t.Error("expected OccupiedSlots map to be initialized")
	}
	if rack.UHeight != 42 {
		t.Errorf("expected UHeight 42, got %d", rack.UHeight)
	}
}

func TestNewRackFromSlugUnknownSlug(t *testing.T) {
	resetRegistries()

	rack, err := NewRackFromSlug("nonexistent-rack")
	if err == nil {
		t.Fatal("expected error for unknown slug")
	}
	if rack != nil {
		t.Error("expected nil rack on error")
	}
	if !strings.Contains(err.Error(), "nonexistent-rack") {
		t.Errorf("error should mention the slug, got: %v", err)
	}
}

// ---------- NewRackFromPartNumber ----------

func TestNewRackFromPartNumberHappyPath(t *testing.T) {
	resetRegistries()
	RegisterRackType(CaniRackType{
		Slug:       "test-rack-48u",
		PartNumber: "PN-RACK-48",
		UHeight:    48,
	})

	rack, err := NewRackFromPartNumber("PN-RACK-48")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rack.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if rack.Status != "Active" {
		t.Errorf("expected status %q, got %q", "Active", rack.Status)
	}
	if rack.OccupiedSlots == nil {
		t.Error("expected OccupiedSlots map to be initialized")
	}
}

func TestNewRackFromPartNumberUnknownPN(t *testing.T) {
	resetRegistries()

	rack, err := NewRackFromPartNumber("BOGUS-RACK-PN")
	if err == nil {
		t.Fatal("expected error for unknown part number")
	}
	if rack != nil {
		t.Error("expected nil rack on error")
	}
	if !strings.Contains(err.Error(), "BOGUS-RACK-PN") {
		t.Errorf("error should mention the part number, got: %v", err)
	}
}

// ---------- NewModuleFromSlug ----------

func TestNewModuleFromSlugHappyPath(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{
		Slug:         "test-psu",
		PartNumber:   "PN-PSU-1",
		Manufacturer: "Acme",
		Model:        "PSU-1600W",
	})

	mod, err := NewModuleFromSlug("test-psu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mod.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if mod.Status != "Active" {
		t.Errorf("expected status %q, got %q", "Active", mod.Status)
	}
	if mod.Name != "PSU-1600W" {
		t.Errorf("expected name %q from model, got %q", "PSU-1600W", mod.Name)
	}
}

func TestNewModuleFromSlugUnknownSlug(t *testing.T) {
	resetRegistries()

	mod, err := NewModuleFromSlug("nonexistent-module")
	if err == nil {
		t.Fatal("expected error for unknown slug")
	}
	if mod != nil {
		t.Error("expected nil module on error")
	}
	if !strings.Contains(err.Error(), "nonexistent-module") {
		t.Errorf("error should mention the slug, got: %v", err)
	}
}

// ---------- NewModuleFromPartNumber ----------

func TestNewModuleFromPartNumberHappyPath(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{
		Slug:         "test-nic",
		PartNumber:   "PN-NIC-25G",
		Manufacturer: "Acme",
		Model:        "NIC-25G",
	})

	mod, err := NewModuleFromPartNumber("PN-NIC-25G")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mod.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if mod.Status != "Active" {
		t.Errorf("expected status %q, got %q", "Active", mod.Status)
	}
	if mod.Name != "NIC-25G" {
		t.Errorf("expected name %q from model, got %q", "NIC-25G", mod.Name)
	}
}

func TestNewModuleFromPartNumberUnknownPN(t *testing.T) {
	resetRegistries()

	mod, err := NewModuleFromPartNumber("BOGUS-MOD-PN")
	if err == nil {
		t.Fatal("expected error for unknown part number")
	}
	if mod != nil {
		t.Error("expected nil module on error")
	}
	if !strings.Contains(err.Error(), "BOGUS-MOD-PN") {
		t.Errorf("error should mention the part number, got: %v", err)
	}
}

// ---------- NewCableFromSlug ----------

func TestNewCableFromSlugHappyPath(t *testing.T) {
	resetRegistries()
	RegisterCableType(CaniCableType{
		Slug:       "test-cat6-cable",
		PartNumber: "PN-CAT6",
		Model:      "CAT6-3m",
	})

	cable, err := NewCableFromSlug("test-cat6-cable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cable.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if cable.Status != "Connected" {
		t.Errorf("expected status %q, got %q", "Connected", cable.Status)
	}
}

func TestNewCableFromSlugUnknownSlug(t *testing.T) {
	resetRegistries()

	cable, err := NewCableFromSlug("nonexistent-cable")
	if err == nil {
		t.Fatal("expected error for unknown slug")
	}
	if cable != nil {
		t.Error("expected nil cable on error")
	}
	if !strings.Contains(err.Error(), "nonexistent-cable") {
		t.Errorf("error should mention the slug, got: %v", err)
	}
}

// ---------- NewCableFromPartNumber ----------

func TestNewCableFromPartNumberHappyPath(t *testing.T) {
	resetRegistries()
	RegisterCableType(CaniCableType{
		Slug:       "test-fiber-cable",
		PartNumber: "PN-FIBER-10",
		Model:      "Fiber-10m",
	})

	cable, err := NewCableFromPartNumber("PN-FIBER-10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cable.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if cable.Status != "Connected" {
		t.Errorf("expected status %q, got %q", "Connected", cable.Status)
	}
}

func TestNewCableFromPartNumberUnknownPN(t *testing.T) {
	resetRegistries()

	cable, err := NewCableFromPartNumber("BOGUS-CABLE-PN")
	if err == nil {
		t.Fatal("expected error for unknown part number")
	}
	if cable != nil {
		t.Error("expected nil cable on error")
	}
	if !strings.Contains(err.Error(), "BOGUS-CABLE-PN") {
		t.Errorf("error should mention the part number, got: %v", err)
	}
}

// ---------- NewFruFromSlug ----------

func TestNewFruFromSlugHappyPath(t *testing.T) {
	resetRegistries()
	RegisterFruType(CaniFruType{
		Slug:       "test-ballast-kit",
		PartNumber: "PN-BALLAST",
		Model:      "BallastKit",
	})

	fru, err := NewFruFromSlug("test-ballast-kit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fru.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if fru.Status != "Active" {
		t.Errorf("expected status %q, got %q", "Active", fru.Status)
	}
}

func TestNewFruFromSlugUnknownSlug(t *testing.T) {
	resetRegistries()

	fru, err := NewFruFromSlug("nonexistent-fru")
	if err == nil {
		t.Fatal("expected error for unknown slug")
	}
	if fru != nil {
		t.Error("expected nil FRU on error")
	}
	if !strings.Contains(err.Error(), "nonexistent-fru") {
		t.Errorf("error should mention the slug, got: %v", err)
	}
}

// ---------- NewFruFromPartNumber ----------

func TestNewFruFromPartNumberHappyPath(t *testing.T) {
	resetRegistries()
	RegisterFruType(CaniFruType{
		Slug:       "test-rail-kit",
		PartNumber: "PN-RAIL",
		Model:      "RailKit",
	})

	fru, err := NewFruFromPartNumber("PN-RAIL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fru.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if fru.Status != "Active" {
		t.Errorf("expected status %q, got %q", "Active", fru.Status)
	}
}

func TestNewFruFromPartNumberUnknownPN(t *testing.T) {
	resetRegistries()

	fru, err := NewFruFromPartNumber("BOGUS-FRU-PN")
	if err == nil {
		t.Fatal("expected error for unknown part number")
	}
	if fru != nil {
		t.Error("expected nil FRU on error")
	}
	if !strings.Contains(err.Error(), "BOGUS-FRU-PN") {
		t.Errorf("error should mention the part number, got: %v", err)
	}
}
