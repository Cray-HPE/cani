package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// inventoryWithRedfishMeta builds an inventory holding one device. When
// redfishMeta is nil the device carries a non-redfish provider sub-map so that
// GetProviderSubMap("redfish") misses; otherwise the map is nested under the
// "redfish" provider key.
func inventoryWithRedfishMeta(id uuid.UUID, redfishMeta map[string]any) *devicetypes.Inventory {
	pm := map[string]any{"csm": map[string]any{"foo": "bar"}}
	if redfishMeta != nil {
		pm = map[string]any{providerKeyRedfish: redfishMeta}
	}
	return &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {
				ID:         id,
				Name:       "existing",
				ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: pm},
			},
		},
	}
}

// assertMatch checks a (uuid, ok) match result against expectations. On a wanted
// match the id must equal wantID; otherwise the id must be uuid.Nil.
func assertMatch(t *testing.T, got uuid.UUID, ok bool, wantID uuid.UUID, wantOK bool) {
	t.Helper()
	if ok != wantOK {
		t.Fatalf("ok = %v, want %v", ok, wantOK)
	}
	if wantOK && got != wantID {
		t.Errorf("id = %s, want %s", got, wantID)
	}
	if !wantOK && got != uuid.Nil {
		t.Errorf("id = %s, want Nil", got)
	}
}

// TestResolveExistingID_NilInventory_NewID verifies resolveExistingID mints a
// fresh UUID when there is no existing inventory to match against.
//
// Why it matters: a first-time import has no prior device, so every root must
// receive a brand-new identity rather than collapsing onto uuid.Nil.
// Inputs: a testRoot() and a nil inventory pointer. Outputs: two generated UUIDs.
// Data choice: calling twice and comparing proves the result is freshly
// generated each time, not a fixed zero value.
func TestResolveExistingID_NilInventory_NewID(t *testing.T) {
	root := testRoot()
	id1 := resolveExistingID(root, nil)
	id2 := resolveExistingID(root, nil)
	if id1 == uuid.Nil || id2 == uuid.Nil {
		t.Fatal("expected non-nil generated UUIDs")
	}
	if id1 == id2 {
		t.Errorf("expected distinct fresh UUIDs, got %s twice", id1)
	}
}

// TestResolveExistingID_BMCIdentityMatch_ReusesID verifies an existing device
// matching both redfish_uuid and BMC FQDN has its UUID reused.
//
// Why it matters: re-importing the same BMC must update the existing device in
// place, keeping the import idempotent instead of duplicating inventory.
// Inputs: a testRoot() and an inventory whose device carries the same
// redfish_uuid and bmc_fqdn. Outputs: the existing device's UUID.
// Data choice: matching on the root's real UUID and FQDN exercises the
// require-both-to-match path that guards against UUID-only collisions.
func TestResolveExistingID_BMCIdentityMatch_ReusesID(t *testing.T) {
	root := testRoot()
	id := uuid.New()
	existing := inventoryWithRedfishMeta(id, map[string]any{
		metaKeyRedfishUUID: root.UUID,
		"bmc_fqdn":         root.ManagerFQDN(),
	})
	if got := resolveExistingID(root, existing); got != id {
		t.Errorf("resolveExistingID() = %s, want existing %s", got, id)
	}
}

// TestResolveExistingID_UUIDMatchButBMCDiffers_NewID verifies that when a root
// has a BMC identity, a UUID-only match with a different BMC does not collapse
// into the existing device.
//
// Why it matters: two distinct endpoints can legitimately share a Redfish UUID;
// treating them as one would silently merge separate hardware.
// Inputs: a testRoot() and an inventory device with the same redfish_uuid but a
// different bmc_fqdn/bmc_hostname. Outputs: a fresh UUID, not the existing one.
// Data choice: identical UUID plus mismatched BMC isolates exactly the
// anti-collapse guard.
func TestResolveExistingID_UUIDMatchButBMCDiffers_NewID(t *testing.T) {
	root := testRoot()
	existingID := uuid.New()
	existing := inventoryWithRedfishMeta(existingID, map[string]any{
		metaKeyRedfishUUID: root.UUID,
		"bmc_fqdn":         "different-host.example.com",
		"bmc_hostname":     "different-host",
	})
	got := resolveExistingID(root, existing)
	if got == existingID {
		t.Error("must not reuse UUID of a device with a different BMC identity")
	}
	if got == uuid.Nil {
		t.Error("expected a freshly generated UUID")
	}
}

// TestResolveExistingID_FallbackUUIDMatch verifies that when the root exposes no
// BMC identity, an existing device is matched on redfish_uuid alone.
//
// Why it matters: roots without OEM manager data still need idempotent matching,
// falling back to the UUID so re-imports do not duplicate the device.
// Inputs: a testRoot() with its OEM Manager stripped and an inventory device
// sharing the redfish_uuid. Outputs: the existing device's UUID.
// Data choice: clearing Manager makes ManagerFQDN/HostName empty, forcing the
// UUID-only fallback branch.
func TestResolveExistingID_FallbackUUIDMatch(t *testing.T) {
	root := testRoot()
	root.Oem.Hpe.Manager = nil
	existingID := uuid.New()
	existing := inventoryWithRedfishMeta(existingID, map[string]any{
		metaKeyRedfishUUID: root.UUID,
	})
	if got := resolveExistingID(root, existing); got != existingID {
		t.Errorf("resolveExistingID() = %s, want existing %s", got, existingID)
	}
}

// TestResolveExistingID_FallbackUUIDNoMatch_NewID verifies a root with no BMC
// identity and no UUID match receives a fresh UUID.
//
// Why it matters: an unrelated existing device must never be reused for a new
// root, or distinct hardware would be conflated.
// Inputs: a testRoot() with OEM Manager stripped and an inventory device with a
// different redfish_uuid. Outputs: a generated UUID distinct from the existing
// device.
// Data choice: a deliberately different stored UUID proves the fallback rejects
// non-matches rather than returning the first device it sees.
func TestResolveExistingID_FallbackUUIDNoMatch_NewID(t *testing.T) {
	root := testRoot()
	root.Oem.Hpe.Manager = nil
	existingID := uuid.New()
	existing := inventoryWithRedfishMeta(existingID, map[string]any{
		metaKeyRedfishUUID: "00000000-0000-0000-0000-000000000000",
	})
	if got := resolveExistingID(root, existing); got == existingID || got == uuid.Nil {
		t.Errorf("expected a fresh UUID distinct from existing, got %s", got)
	}
}

// TestMatchByBMCIdentity verifies BMC-identity matching requires redfish_uuid to
// match AND either the FQDN or hostname to match, skipping devices that miss any
// of these.
//
// Why it matters: this is the dedup core that keeps re-imports idempotent while
// refusing to merge endpoints that merely share a UUID.
// Inputs: a testRoot() and, per case, an inventory device with a crafted redfish
// sub-map. Outputs: the matched (uuid, ok) pair asserted by assertMatch.
// Data choice: the cases isolate each branch — fqdn hit, hostname hit, uuid
// mismatch, BMC mismatch, and a device with no redfish sub-map.
func TestMatchByBMCIdentity(t *testing.T) {
	root := testRoot()
	fqdn := root.ManagerFQDN()
	host := root.ManagerHostName()
	cases := []struct {
		name   string
		meta   map[string]any
		wantOK bool
	}{
		{"fqdn match", map[string]any{metaKeyRedfishUUID: root.UUID, "bmc_fqdn": fqdn}, true},
		{"hostname match", map[string]any{metaKeyRedfishUUID: root.UUID, "bmc_hostname": host}, true},
		{"uuid mismatch", map[string]any{metaKeyRedfishUUID: "other", "bmc_fqdn": fqdn}, false},
		{"bmc mismatch", map[string]any{metaKeyRedfishUUID: root.UUID, "bmc_fqdn": "x", "bmc_hostname": "y"}, false},
		{"no redfish meta", nil, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()
			existing := inventoryWithRedfishMeta(id, tt.meta)
			got, ok := matchByBMCIdentity(root, existing, fqdn, host)
			assertMatch(t, got, ok, id, tt.wantOK)
		})
	}
}

// TestMatchByRedfishUUID verifies UUID-only matching returns the device whose
// redfish_uuid equals the root's, and reports no match otherwise.
//
// Why it matters: this is the fallback dedup path for roots that expose no BMC
// identity, so it must be precise about UUID equality.
// Inputs: a testRoot() and an inventory device whose redfish_uuid either equals
// the root's or differs. Outputs: the matched (uuid, ok) pair.
// Data choice: one exact-UUID device and one different-UUID device prove both the
// hit and miss branches.
func TestMatchByRedfishUUID(t *testing.T) {
	root := testRoot()
	t.Run("match", func(t *testing.T) {
		id := uuid.New()
		existing := inventoryWithRedfishMeta(id, map[string]any{metaKeyRedfishUUID: root.UUID})
		got, ok := matchByRedfishUUID(root, existing)
		assertMatch(t, got, ok, id, true)
	})
	t.Run("no match", func(t *testing.T) {
		existing := inventoryWithRedfishMeta(uuid.New(), map[string]any{metaKeyRedfishUUID: "different"})
		got, ok := matchByRedfishUUID(root, existing)
		assertMatch(t, got, ok, uuid.Nil, false)
	})
}

// TestProviderValueEquals verifies the helper compares a metadata value to a
// string only when the key exists and holds a string.
//
// Why it matters: provider metadata is an untyped map, so matching logic must
// guard against missing keys and non-string values before comparing.
// Inputs: a fixed metadata map and, per case, a key/value pair. Outputs: the
// boolean equality result.
// Data choice: the map mixes a string and an int so the cases cover equal,
// unequal, missing-key, and wrong-type outcomes.
func TestProviderValueEquals(t *testing.T) {
	meta := map[string]any{"k": "v", "n": 42}
	cases := []struct {
		name, key, val string
		want           bool
	}{
		{"equal", "k", "v", true},
		{"not equal", "k", "x", false},
		{"missing key", "missing", "v", false},
		{"non-string value", "n", "42", false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := providerValueEquals(meta, tt.key, tt.val); got != tt.want {
				t.Errorf("providerValueEquals(meta, %q, %q) = %v, want %v", tt.key, tt.val, got, tt.want)
			}
		})
	}
}
