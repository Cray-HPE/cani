package datastores

// | Function     | Happy-path test                    | Failure test              |
// |--------------|------------------------------------|---------------------------|
// | NewJSONStore | TestNewJSONStoreHappyPath           | TestNewJSONStoreNilConfig |
// | Load         | TestLoadHappyPath                  | TestLoadInvalidJSON       |
// | Save         | TestSaveHappyPath                  | TestSaveInvalidPath       |
// | Save/Load    | TestJSONStorePersistsCRUDMutations |                           |

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

type crudFixtureIDs struct {
	locationID uuid.UUID
	rackID     uuid.UUID
	nodeID     uuid.UUID
	switchID   uuid.UUID
	cableID    uuid.UUID
	vlanID     uuid.UUID
	prefixID   uuid.UUID
	ipID       uuid.UUID
}

func newCRUDInventory(t *testing.T) (*devicetypes.Inventory, crudFixtureIDs) {
	t.Helper()

	ids := crudFixtureIDs{
		locationID: uuid.MustParse("10000000-0000-0000-0000-000000000001"),
		rackID:     uuid.MustParse("20000000-0000-0000-0000-000000000001"),
		nodeID:     uuid.MustParse("30000000-0000-0000-0000-000000000001"),
		switchID:   uuid.MustParse("40000000-0000-0000-0000-000000000001"),
		cableID:    uuid.MustParse("50000000-0000-0000-0000-000000000001"),
		vlanID:     uuid.MustParse("60000000-0000-0000-0000-000000000001"),
		prefixID:   uuid.MustParse("70000000-0000-0000-0000-000000000001"),
		ipID:       uuid.MustParse("80000000-0000-0000-0000-000000000001"),
	}

	inv := devicetypes.NewInventory()
	location := &devicetypes.CaniLocationType{
		ID:           ids.locationID,
		Name:         "site-a",
		LocationType: "site",
		ObjectMeta:   devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive)},
	}
	if err := inv.AddLocation(location); err != nil {
		t.Fatalf("adding location: %v", err)
	}

	rack := &devicetypes.CaniRackType{
		ID:         ids.rackID,
		Name:       "rack-a",
		Location:   ids.locationID,
		UHeight:    42,
		ObjectMeta: devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive)},
	}
	if err := inv.AddRack(rack); err != nil {
		t.Fatalf("adding rack: %v", err)
	}

	node := &devicetypes.CaniDeviceType{
		ID:           ids.nodeID,
		Name:         "node-a",
		Type:         devicetypes.TypeNode,
		Parent:       ids.rackID,
		Rack:         ids.rackID,
		RackPosition: 10,
		Face:         "front",
		ObjectMeta: devicetypes.ObjectMeta{
			Status: string(devicetypes.StatusActive),
			ProviderMetadata: map[string]any{
				"csm": map[string]any{
					"xname": "x3000c0s1b0n0",
					"nid":   1001,
				},
			},
		},
	}
	switchDevice := &devicetypes.CaniDeviceType{
		ID:           ids.switchID,
		Name:         "leaf-a",
		Type:         devicetypes.TypeMgmtSwitch,
		Parent:       ids.rackID,
		Rack:         ids.rackID,
		RackPosition: 20,
		Face:         "rear",
		ObjectMeta:   devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive)},
	}
	if err := inv.AddDevices(map[uuid.UUID]*devicetypes.CaniDeviceType{
		ids.nodeID:   node,
		ids.switchID: switchDevice,
	}); err != nil {
		t.Fatalf("adding devices: %v", err)
	}

	cable := devicetypes.NewCable("dac-passive", "node-a-to-leaf-a")
	cable.ID = ids.cableID
	cable.SetDeviceTerminations(ids.nodeID, ids.switchID, "eth0", "1/1/1")
	if err := inv.AddCable(cable); err != nil {
		t.Fatalf("adding cable: %v", err)
	}

	vlan := &devicetypes.CaniVLAN{
		ID:         ids.vlanID,
		VID:        1513,
		Name:       "HMN",
		Location:   ids.locationID,
		ObjectMeta: devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive)},
	}
	if err := inv.AddVLAN(vlan); err != nil {
		t.Fatalf("adding vlan: %v", err)
	}

	prefix := &devicetypes.CaniPrefix{
		ID:         ids.prefixID,
		Prefix:     "10.1.0.0/24",
		Type:       devicetypes.PrefixTypeNetwork,
		Location:   ids.locationID,
		VLAN:       ids.vlanID,
		ObjectMeta: devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive)},
	}
	if err := inv.AddPrefix(prefix); err != nil {
		t.Fatalf("adding prefix: %v", err)
	}

	address := &devicetypes.CaniIPAddress{
		ID:         ids.ipID,
		Address:    "10.1.0.10/24",
		Type:       devicetypes.IPAddressTypeHost,
		DNSName:    "node-a.example.com",
		ObjectMeta: devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive)},
	}
	if err := inv.AddIPAddress(address); err != nil {
		t.Fatalf("adding ip address: %v", err)
	}

	return inv, ids
}

func assertCreatedCRUDInventory(t *testing.T, inv *devicetypes.Inventory, ids crudFixtureIDs) {
	t.Helper()

	if inv.SchemaVersion != devicetypes.SchemaVersionV1Alpha3 {
		t.Fatalf("SchemaVersion = %q, want %q", inv.SchemaVersion, devicetypes.SchemaVersionV1Alpha3)
	}
	assertCreatedLocation(t, inv, ids)
	assertCreatedRack(t, inv, ids)
	assertCreatedDevices(t, inv, ids)
	assertCreatedCable(t, inv, ids)
	assertCreatedIPAM(t, inv, ids)
}

func assertCreatedLocation(t *testing.T, inv *devicetypes.Inventory, ids crudFixtureIDs) {
	t.Helper()

	location := inv.Locations[ids.locationID]
	if location == nil {
		t.Fatalf("location %s was not loaded", ids.locationID)
	}
	if location.Name != "site-a" || location.LocationType != "site" {
		t.Fatalf("location = %+v, want site-a site", location)
	}
	if !hasUUID(location.Racks, ids.rackID) {
		t.Fatalf("location racks = %v, want %s", location.Racks, ids.rackID)
	}
}

func assertCreatedRack(t *testing.T, inv *devicetypes.Inventory, ids crudFixtureIDs) {
	t.Helper()

	rack := inv.Racks[ids.rackID]
	if rack == nil {
		t.Fatalf("rack %s was not loaded", ids.rackID)
	}
	if rack.Location != ids.locationID || rack.UHeight != 42 {
		t.Fatalf("rack Location/UHeight = %s/%d, want %s/42", rack.Location, rack.UHeight, ids.locationID)
	}
	if !hasUUID(rack.Devices, ids.nodeID) || !hasUUID(rack.Devices, ids.switchID) {
		t.Fatalf("rack devices = %v, want node %s and switch %s", rack.Devices, ids.nodeID, ids.switchID)
	}
}

func assertCreatedDevices(t *testing.T, inv *devicetypes.Inventory, ids crudFixtureIDs) {
	t.Helper()

	node := inv.Devices[ids.nodeID]
	if node == nil {
		t.Fatalf("node %s was not loaded", ids.nodeID)
	}
	if node.Name != "node-a" || node.Parent != ids.rackID || node.Rack != ids.rackID {
		t.Fatalf("node identity/reference = %+v, want node-a in rack %s", node, ids.rackID)
	}
	if found := inv.FindDeviceByProviderKey("csm", "xname", "x3000c0s1b0n0"); found == nil || found.ID != ids.nodeID {
		t.Fatalf("provider-key lookup returned %+v, want node %s", found, ids.nodeID)
	}
}

func assertCreatedCable(t *testing.T, inv *devicetypes.Inventory, ids crudFixtureIDs) {
	t.Helper()

	cable := inv.Cables[ids.cableID]
	if cable == nil {
		t.Fatalf("cable %s was not loaded", ids.cableID)
	}
	if cable.TerminationADevice != ids.nodeID || cable.TerminationBDevice != ids.switchID {
		t.Fatalf("cable terminations = %s/%s, want %s/%s", cable.TerminationADevice, cable.TerminationBDevice, ids.nodeID, ids.switchID)
	}
	if cable.TerminationAPort != "eth0" || cable.TerminationBPort != "1/1/1" {
		t.Fatalf("cable ports = %q/%q, want eth0/1/1/1", cable.TerminationAPort, cable.TerminationBPort)
	}
}

func assertCreatedIPAM(t *testing.T, inv *devicetypes.Inventory, ids crudFixtureIDs) {
	t.Helper()

	vlan := inv.VLANs[ids.vlanID]
	if vlan == nil || vlan.VID != 1513 || vlan.Location != ids.locationID {
		t.Fatalf("vlan = %+v, want VID 1513 at location %s", vlan, ids.locationID)
	}
	prefix := inv.Prefixes[ids.prefixID]
	if prefix == nil || prefix.Prefix != "10.1.0.0/24" || prefix.VLAN != ids.vlanID {
		t.Fatalf("prefix = %+v, want 10.1.0.0/24 on VLAN %s", prefix, ids.vlanID)
	}
	address := inv.IPAddresses[ids.ipID]
	if address == nil || address.Address != "10.1.0.10/24" || address.Parent != ids.prefixID {
		t.Fatalf("ip address = %+v, want 10.1.0.10/24 under prefix %s", address, ids.prefixID)
	}
}

func assertUpdatedCRUDInventory(t *testing.T, inv *devicetypes.Inventory, ids crudFixtureIDs) {
	t.Helper()

	assertUpdatedNode(t, inv, ids)
	if _, ok := inv.Devices[ids.switchID]; ok {
		t.Fatalf("deleted switch %s was loaded", ids.switchID)
	}
	if _, ok := inv.Cables[ids.cableID]; ok {
		t.Fatalf("cable %s referencing deleted switch was loaded", ids.cableID)
	}
	if rack := inv.Racks[ids.rackID]; rack == nil || hasUUID(rack.Devices, ids.switchID) {
		t.Fatalf("rack after delete = %+v, want switch %s absent", rack, ids.switchID)
	}
	if address := inv.IPAddresses[ids.ipID]; address == nil || address.Parent != ids.prefixID {
		t.Fatalf("ip address after update = %+v, want parent %s", address, ids.prefixID)
	}
}

func assertUpdatedNode(t *testing.T, inv *devicetypes.Inventory, ids crudFixtureIDs) {
	t.Helper()

	node := inv.Devices[ids.nodeID]
	if node == nil {
		t.Fatalf("updated node %s was not loaded", ids.nodeID)
	}
	if node.Name != "node-a-renamed" || node.Status != string(devicetypes.StatusPlanned) {
		t.Fatalf("node update = %q/%q, want node-a-renamed/%q", node.Name, node.Status, devicetypes.StatusPlanned)
	}
	if node.Parent != ids.rackID || node.Rack != ids.rackID {
		t.Fatalf("node rack reference = parent %s rack %s, want %s", node.Parent, node.Rack, ids.rackID)
	}
	if found := inv.FindDeviceByProviderKey("csm", "xname", "x3000c0s1b0n1"); found == nil || found.ID != ids.nodeID {
		t.Fatalf("updated provider-key lookup returned %+v, want node %s", found, ids.nodeID)
	}
}

func hasUUID(values []uuid.UUID, want uuid.UUID) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// TestNewJSONStoreHappyPath verifies relative datastore paths are resolved next
// to the loaded config file.
//
// Why it matters: commands depend on the JSON store using the same datastore
// path the user configured instead of the process working directory.
// Inputs: a config path under /tmp/cani-test and a relative inventory file.
// Outputs: a JSONStore whose Path points beside the config file.
// Data choice: the relative filename is the common cani config shape.
func TestNewJSONStoreHappyPath(t *testing.T) {
	original := config.Cfg
	config.Cfg = &config.Config{
		Path:      "/tmp/cani-test/config.yaml",
		Datastore: "inventory.json",
	}
	defer func() { config.Cfg = original }()

	store := NewJSONStore()

	expected := filepath.Join("/tmp/cani-test", "inventory.json")
	if store.Path != expected {
		t.Errorf("expected path %q, got %q", expected, store.Path)
	}
}

// TestNewJSONStoreAbsolutePath verifies absolute datastore paths are preserved.
//
// Why it matters: callers can intentionally keep the datastore outside the
// config directory, and path resolution must not rewrite that choice.
// Inputs: a config path and an absolute datastore path. Outputs: a JSONStore
// with the same absolute datastore path.
// Data choice: /tmp/override/test.json makes the absolute path branch explicit.
func TestNewJSONStoreAbsolutePath(t *testing.T) {
	original := config.Cfg
	config.Cfg = &config.Config{
		Path:      "/tmp/cani-test/config.yaml",
		Datastore: "/tmp/override/test.json",
	}
	defer func() { config.Cfg = original }()

	store := NewJSONStore()

	expected := "/tmp/override/test.json"
	if store.Path != expected {
		t.Errorf("expected path %q, got %q", expected, store.Path)
	}
}

// TestNewJSONStoreNilConfig verifies NewJSONStore panics when global config is nil.
//
// Why it matters: the constructor currently requires config initialization before
// datastore setup, so the test documents that failure mode.
// Inputs: a nil config.Cfg. Outputs: a recovered non-nil panic value.
// Data choice: nil is the only config state that exercises this dereference path.
func TestNewJSONStoreNilConfig(t *testing.T) {
	original := config.Cfg
	config.Cfg = nil
	defer func() { config.Cfg = original }()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when config.Cfg is nil, but did not panic")
		}
	}()

	NewJSONStore()
}

// TestLoadHappyPath verifies Load restores persisted inventory content.
//
// Why it matters: datastore reads must preserve the CRUD state produced by the
// inventory layer, including IDs, relationships, IPAM objects, and metadata.
// Inputs: an on-disk JSON inventory with locations, racks, devices, a cable,
// VLAN, prefix, and IP address. Outputs: a loaded Inventory with matching data.
// Data choice: the fixture combines physical and IPAM references to catch data
// loss beyond a non-nil inventory pointer.
func TestLoadHappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inventory.json")

	inv, ids := newCRUDInventory(t)
	data, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		t.Fatalf("marshaling test inventory: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	store := &JSONStore{Path: path}
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load() returned nil inventory")
	}
	assertCreatedCRUDInventory(t, loaded, ids)
}

// TestLoadInvalidJSON verifies Load rejects malformed datastore JSON.
//
// Why it matters: callers need a clear error instead of silently accepting a
// corrupted datastore and losing inventory data.
// Inputs: a datastore file containing invalid JSON. Outputs: an error from Load.
// Data choice: a short invalid object exercises JSON parsing before migration or
// inventory post-processing can occur.
func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inventory.json")

	if err := os.WriteFile(path, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	store := &JSONStore{Path: path}
	_, err := store.Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parsing inventory") {
		t.Fatalf("Load() error = %v, want parsing inventory context", err)
	}
}

// TestSaveHappyPath verifies Save writes a complete JSON inventory to disk.
//
// Why it matters: datastore writes are the persistence boundary for inventory
// creates, reference changes, provider metadata, and IPAM state.
// Inputs: an inventory populated through CRUD helpers and a nested output path.
// Outputs: a created JSON file that unmarshals back to the same inventory data.
// Data choice: the nested path proves directory creation while the populated
// inventory proves Save does more than create an empty file.
func TestSaveHappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "inventory.json")

	store := &JSONStore{Path: path}
	inv, ids := newCRUDInventory(t)

	if err := store.Save(inv); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Save() did not create the file")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}

	loaded := devicetypes.NewInventory()
	if err := json.Unmarshal(data, loaded); err != nil {
		t.Fatalf("saved file contains invalid JSON: %v", err)
	}
	assertCreatedCRUDInventory(t, loaded, ids)
}

// TestJSONStorePersistsCRUDMutations verifies save-load cycles preserve creates,
// updates, and deletes.
//
// Why it matters: higher-level CRUD operations rely on JSONStore to persist the
// inventory state exactly as mutated between command invocations.
// Inputs: an inventory with physical and IPAM data, then an updated node and a
// removed switch. Outputs: reloads that keep the update and omit deleted data.
// Data choice: deleting the switch also removes its cable, proving reference
// cleanup survives the datastore round trip.
func TestJSONStorePersistsCRUDMutations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inventory.json")
	store := &JSONStore{Path: path}

	inv, ids := newCRUDInventory(t)
	if err := store.Save(inv); err != nil {
		t.Fatalf("initial Save() returned unexpected error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("initial Load() returned unexpected error: %v", err)
	}
	assertCreatedCRUDInventory(t, loaded, ids)

	loaded.MergeDevices(map[uuid.UUID]*devicetypes.CaniDeviceType{
		ids.nodeID: {
			ID:   ids.nodeID,
			Name: "node-a-renamed",
			Type: devicetypes.TypeNode,
			ObjectMeta: devicetypes.ObjectMeta{
				Status: string(devicetypes.StatusPlanned),
				ProviderMetadata: map[string]any{
					"csm": map[string]any{"xname": "x3000c0s1b0n1"},
				},
			},
		},
	})
	if err := loaded.RemoveDevice(ids.switchID); err != nil {
		t.Fatalf("removing switch: %v", err)
	}
	if err := store.Save(loaded); err != nil {
		t.Fatalf("updated Save() returned unexpected error: %v", err)
	}

	reloaded, err := store.Load()
	if err != nil {
		t.Fatalf("updated Load() returned unexpected error: %v", err)
	}

	assertUpdatedCRUDInventory(t, reloaded, ids)
}

// TestSaveInvalidPath verifies Save reports directory creation failures.
//
// Why it matters: callers need a hard failure when the datastore cannot be
// written, otherwise CRUD changes could appear to succeed but be lost.
// Inputs: an output path whose parent component is an existing file. Outputs: an
// error from Save.
// Data choice: the blocker file deterministically exercises the MkdirAll failure
// path without relying on host permissions.
func TestSaveInvalidPath(t *testing.T) {
	dir := t.TempDir()

	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0644); err != nil {
		t.Fatalf("creating blocker file: %v", err)
	}

	path := filepath.Join(blocker, "sub", "inventory.json")

	store := &JSONStore{Path: path}
	inv := devicetypes.NewInventory()

	err := store.Save(inv)
	if err == nil {
		t.Fatal("Save() expected error for invalid path, got nil")
	}
	if !strings.Contains(err.Error(), "creating inventory directory") {
		t.Fatalf("Save() error = %v, want creating inventory directory context", err)
	}
}
