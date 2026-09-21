package devicetypes

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// TestRemoveDeviceCleansOwnedReferences verifies deletion follows forward FKs
// and detaches shared network objects without deleting unrelated hardware.
//
// Why it matters: successful removal must leave an exportable inventory.
// Inputs: a device tree with a module, nested FRUs, ports, shared IP and VRF.
// Outputs: only peer hardware remains; shared objects reference only the peer.
// Data choice: stale child and interface caches must not decide ownership.
func TestRemoveDeviceCleansOwnedReferences(t *testing.T) {
	fixture := newRemovalReferenceFixture(t)
	inventory := fixture.inventory
	inventory.Devices[fixture.root].Children = []uuid.UUID{fixture.peer}
	inventory.Interfaces[fixture.peerInterface].DeviceID = fixture.root

	if err := inventory.RemoveDevice(fixture.root); err != nil {
		t.Fatalf("RemoveDevice() error = %v", err)
	}

	assertOnlyPeerHardware(t, fixture)
	if got := inventory.IPAddresses[fixture.address].Interfaces; !reflect.DeepEqual(got, []uuid.UUID{fixture.peerInterface}) {
		t.Errorf("IP assignments = %v, want only peer interface", got)
	}
	if got := inventory.VRFs[fixture.vrf].Devices; !reflect.DeepEqual(got, []uuid.UUID{fixture.peer}) {
		t.Errorf("VRF devices = %v, want only peer", got)
	}
	if inventory.Devices[fixture.peer].BMCParent != uuid.Nil {
		t.Error("peer retained a BMC parent reference to the removed device")
	}
	if inventory.lookupProviderKey("source", "id", "root") != uuid.Nil {
		t.Error("provider-key cache retained the deleted device")
	}
	assertRemovalRelationships(t, fixture)
}

// TestRemoveModuleCleansOwnedReferences verifies module deletion removes only
// that module's ports and FRUs, including cables addressed through its device.
//
// Why it matters: modules share a parent device with interfaces that must remain.
// Inputs: a module port referenced by a shared IP and a device-addressed cable.
// Outputs: device ports and VRF membership survive; the module references vanish.
// Data choice: the cable names the parent device, so owner-ID matching is insufficient.
func TestRemoveModuleCleansOwnedReferences(t *testing.T) {
	fixture := newRemovalReferenceFixture(t)
	inventory := fixture.inventory

	if err := inventory.RemoveModule(fixture.module); err != nil {
		t.Fatalf("RemoveModule() error = %v", err)
	}

	if len(inventory.Devices) != 3 || len(inventory.Modules) != 0 || len(inventory.Cables) != 0 {
		t.Errorf("incorrect surviving collections: devices=%d modules=%d cables=%d", len(inventory.Devices), len(inventory.Modules), len(inventory.Cables))
	}
	if inventory.Frus[fixture.moduleFRU] != nil || inventory.Frus[fixture.peerFRU] == nil || len(inventory.Frus) != 3 {
		t.Errorf("module FRU cleanup affected the wrong objects: %v", inventory.Frus)
	}
	if inventory.Interfaces[fixture.moduleInterface] != nil || len(inventory.Interfaces) != 3 {
		t.Errorf("module interface cleanup affected the wrong objects: %v", inventory.Interfaces)
	}
	wantInterfaces := []uuid.UUID{fixture.rootInterface, fixture.childInterface, fixture.peerInterface}
	if got := inventory.IPAddresses[fixture.address].Interfaces; !reflect.DeepEqual(got, wantInterfaces) {
		t.Errorf("IP assignments = %v, want %v", got, wantInterfaces)
	}
	wantDevices := []uuid.UUID{fixture.root, fixture.child, fixture.peer}
	if got := inventory.VRFs[fixture.vrf].Devices; !reflect.DeepEqual(got, wantDevices) {
		t.Errorf("VRF assignments = %v, want %v", got, wantDevices)
	}
	assertRemovalRelationships(t, fixture)
}

type removalReferenceFixture struct {
	inventory                                                     *Inventory
	root, child, peer, module                                     uuid.UUID
	rootInterface, childInterface, peerInterface, moduleInterface uuid.UUID
	peerFRU, moduleFRU, address, vrf                              uuid.UUID
}

func newRemovalReferenceFixture(t *testing.T) removalReferenceFixture {
	t.Helper()
	fixture := removalReferenceFixture{
		inventory: NewInventory(), root: uuid.New(), child: uuid.New(), peer: uuid.New(), module: uuid.New(),
		rootInterface: uuid.New(), childInterface: uuid.New(), peerInterface: uuid.New(), moduleInterface: uuid.New(),
		peerFRU: uuid.New(), moduleFRU: uuid.New(), address: uuid.New(), vrf: uuid.New(),
	}
	inventory := fixture.inventory
	inventory.Devices[fixture.root] = &CaniDeviceType{
		ID: fixture.root, Name: "root",
		Interfaces: []InterfaceSpec{{ID: fixture.rootInterface, Name: "root-port"}},
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{"source": map[string]any{"id": "root"}}},
	}
	inventory.Devices[fixture.child] = &CaniDeviceType{
		ID: fixture.child, Name: "child", Parent: fixture.root,
		Interfaces: []InterfaceSpec{{ID: fixture.childInterface, Name: "child-port"}},
	}
	cableID := uuid.New()
	inventory.Devices[fixture.peer] = &CaniDeviceType{
		ID: fixture.peer, Name: "peer", BMCParent: fixture.root,
		Interfaces: []InterfaceSpec{{ID: fixture.peerInterface, Name: "peer-port", ConnectedCable: &cableID}},
	}
	inventory.Modules[fixture.module] = &CaniModuleType{
		ID: fixture.module, Name: "nic", ParentDevice: fixture.root,
		Interfaces: []InterfaceSpec{{ID: fixture.moduleInterface, Name: "nic-port"}},
	}
	rootFRU, nestedFRU := uuid.New(), uuid.New()
	inventory.Frus[rootFRU] = &CaniFruType{ID: rootFRU, Name: "root-fru", Device: fixture.root}
	inventory.Frus[nestedFRU] = &CaniFruType{ID: nestedFRU, Name: "nested-fru", Parent: rootFRU}
	inventory.Frus[fixture.moduleFRU] = &CaniFruType{ID: fixture.moduleFRU, Name: "nic-fru", Device: fixture.module}
	inventory.Frus[fixture.peerFRU] = &CaniFruType{ID: fixture.peerFRU, Name: "peer-fru", Device: fixture.peer}
	inventory.Cables[cableID] = &CaniCableType{
		ID: cableID, Label: "module-to-peer",
		TerminationADevice: fixture.root, TerminationAPort: "nic-port", TerminationA: fixture.moduleInterface,
		TerminationBDevice: fixture.peer, TerminationBPort: "peer-port", TerminationB: fixture.peerInterface,
	}
	inventory.IPAddresses[fixture.address] = &CaniIPAddress{
		ID: fixture.address, Address: "10.0.0.1/24",
		Interfaces: []uuid.UUID{fixture.rootInterface, fixture.childInterface, fixture.moduleInterface, fixture.peerInterface},
	}
	inventory.VRFs[fixture.vrf] = &CaniVRF{ID: fixture.vrf, Name: "shared", Devices: []uuid.UUID{fixture.root, fixture.child, fixture.peer}}
	if err := inventory.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("invalid fixture: %v", err)
	}
	inventory.RebuildProviderKeyIndex()
	return fixture
}

func assertRemovalRelationships(t *testing.T, fixture removalReferenceFixture) {
	t.Helper()
	inventory := fixture.inventory
	if inventory.Devices[fixture.peer].Interfaces[0].ConnectedCable != nil {
		t.Error("surviving port retained a reference to the deleted cable")
	}
	if err := inventory.RebuildDerivedState().Err(); err != nil {
		t.Errorf("relationships invalid after removal: %v", err)
	}
}

func assertOnlyPeerHardware(t *testing.T, fixture removalReferenceFixture) {
	t.Helper()
	inventory := fixture.inventory
	if len(inventory.Devices) != 1 || inventory.Devices[fixture.peer] == nil {
		t.Fatalf("device deletion did not preserve only the peer: %v", inventory.Devices)
	}
	if len(inventory.Modules) != 0 || len(inventory.Cables) != 0 {
		t.Errorf("owned modules or cables remain: modules=%v cables=%v", inventory.Modules, inventory.Cables)
	}
	if len(inventory.Frus) != 1 || inventory.Frus[fixture.peerFRU] == nil {
		t.Errorf("FRU deletion did not preserve only the peer FRU: %v", inventory.Frus)
	}
	if len(inventory.Interfaces) != 1 || inventory.Interfaces[fixture.peerInterface] == nil {
		t.Errorf("interface deletion did not preserve only the peer interface: %v", inventory.Interfaces)
	}
}
